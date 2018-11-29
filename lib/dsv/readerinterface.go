// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"bytes"
	"fmt"
	"io"
	"os"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/tabula"
)

//
// ReaderInterface is the interface for reading DSV file.
//
type ReaderInterface interface {
	ConfigInterface
	AddInputMetadata(*Metadata)
	AppendMetadata(MetadataInterface)
	GetInputMetadata() []MetadataInterface
	GetInputMetadataAt(idx int) MetadataInterface
	GetMaxRows() int
	SetMaxRows(max int)
	GetDatasetMode() string
	SetDatasetMode(mode string)
	GetNColumnIn() int
	GetInput() string
	SetInput(path string)
	GetRejected() string
	SetRejected(path string)
	GetSkip() int
	SetSkip(n int)
	IsTrimSpace() bool
	SetDefault()
	OpenInput() error
	OpenRejected() error
	SkipLines() error

	Reset() error
	Flush() error
	ReadLine() ([]byte, error)
	FetchNextLine([]byte) ([]byte, error)
	Reject(line []byte) (int, error)
	Close() error

	GetDataset() interface{}
	MergeColumns(ReaderInterface)
}

//
// Read row from input file.
//
func Read(reader ReaderInterface) (n int, e error) {
	var (
		row     *tabula.Row
		line    []byte
		linenum int
		eRead   *ReaderError
	)
	maxrows := reader.GetMaxRows()

	e = reader.Reset()
	if e != nil {
		return
	}

	dataset := reader.GetDataset().(tabula.DatasetInterface)

	// Loop until we reached MaxRows (> 0) or when all rows has been
	// read (= -1)
	for {
		row, line, linenum, eRead = ReadRow(reader, linenum)
		if nil == eRead {
			dataset.PushRow(row)

			n++
			if maxrows > 0 && n >= maxrows {
				break
			}
			continue
		}

		if eRead.T&EReadEOF == EReadEOF {
			_ = reader.Flush()
			e = io.EOF
			return
		}

		eRead.N = linenum
		fmt.Fprintf(os.Stderr, "%s\n", eRead)

		// If error, save the rejected line.
		line = append(line, DefEOL)

		_, e = reader.Reject(line)
		if e != nil {
			break
		}
	}

	// remember to flush if we have rejected rows.
	e = reader.Flush()

	return n, e
}

//
// parsingLeftQuote parse the left-quote string from line.
//
func parsingLeftQuote(lq, line []byte, startAt int) (
	p int, eRead *ReaderError,
) {
	p = startAt

	// parsing until we found left quote token
	p, found := libbytes.SkipAfterToken(line, lq, p, false)

	if found {
		return p, nil
	}

	eRead = &ReaderError{
		T:    EReadMissLeftQuote,
		Func: "parsingLeftQuote",
		What: "Missing left-quote '" + string(lq) + "'",
		Line: string(line),
		Pos:  p,
		N:    0,
	}

	return p, eRead
}

//
// parsingSeparator parsing the line until we found the separator.
//
// Return the data and index of last parsed line, or error if separator is not
// found or not match with specification.
//
func parsingSeparator(sep, line []byte, startAt int) (
	v []byte, p int, eRead *ReaderError,
) {
	p = startAt

	v, p, found := libbytes.CutUntilToken(line, sep, p, false)

	if found {
		return v, p, nil
	}

	eRead = &ReaderError{
		Func: "parsingSeparator",
		What: "Missing separator '" + string(sep) + "'",
		Line: string(line),
		Pos:  p,
		N:    0,
	}

	return v, p, eRead
}

//
// parsingRightQuote parsing the line until we found the right quote or separator.
//
// Return the data and index of last parsed line, or error if right-quote is not
// found or not match with specification.
//
func parsingRightQuote(reader ReaderInterface, rq, line []byte, startAt int) (
	v, lines []byte, p int, eRead *ReaderError,
) {
	var e error
	var content []byte
	p = startAt
	var found bool

	// (2.2.1)
	for {
		content, p, found = libbytes.CutUntilToken(line, rq, p, true)

		v = append(v, content...)

		if found {
			return v, line, p, nil
		}

		// EOL before finding right-quote.
		// Read and join with the next line.
		line, e = reader.FetchNextLine(line)

		if e != nil {
			break
		}
	}

	eRead = &ReaderError{
		T:    EReadMissRightQuote,
		Func: "parsingRightQuote",
		What: "Missing right-quote '" + string(rq) + "'",
		Line: string(line),
		Pos:  p,
		N:    0,
	}

	if e == io.EOF {
		eRead.T &= EReadEOF
	}

	return v, line, p, eRead
}

//
// parsingSkipSeparator parse until we found separator or EOF
//
func parsingSkipSeparator(sep, line []byte, startAt int) (
	p int, eRead *ReaderError,
) {
	p = startAt

	p, found := libbytes.SkipAfterToken(line, sep, p, false)

	if found {
		return p, nil
	}

	eRead = &ReaderError{
		T:    EReadMissSeparator,
		Func: "parsingSkipSeparator",
		What: "Missing separator '" + string(sep) + "'",
		Line: string(line),
		Pos:  p,
		N:    0,
	}

	return p, eRead
}

//
// parsingSkipSpace skip all space starting from `startAt`.
//
func parsingSkipSpace(line []byte, startAt int) (p int) {
	linelen := len(line)

	for p = startAt; p < linelen; p++ {
		if line[p] == ' ' || line[p] == '\t' || line[p] == '\n' ||
			line[p] == '\r' {
			continue
		}
		break
	}
	return
}

//
// ParseLine parse a line containing records. The output is array of record
// (or single row).
//
// This is how the algorithm works
// (1) create n slice of record, where n is number of column metadata
// (2) for each metadata
// (2.0) Check if the next sequence matched with separator.
// (2.0.1) If its match, create empty record
// (2.1) If using left quote, skip until we found left-quote
// (2.2) If using right quote, append byte to buffer until right-quote
// 	(2.2.1) If using separator, skip until separator
// (2.3) If using separator, append byte to buffer until separator
// (2.4) else append all byte to buffer.
// (3) save buffer to record
//
func ParseLine(reader ReaderInterface, line []byte) (
	prow *tabula.Row, eRead *ReaderError,
) {
	p := 0
	rIdx := 0
	inputMd := reader.GetInputMetadata()
	row := make(tabula.Row, 0)

	for _, md := range inputMd {
		lq := md.GetLeftQuote()
		rq := md.GetRightQuote()
		sep := md.GetSeparator()
		v := []byte{}

		// (2.0)
		if sep != "" && sep != lq {
			match := libbytes.IsTokenAt(line, []byte(sep), p)

			// (2.0.1)
			if match {
				p += len(sep)
				goto empty
			}
		}

		// (2.1)
		if lq != "" {
			p, eRead = parsingLeftQuote([]byte(lq), line, p)

			if eRead != nil {
				return
			}
		}

		// (2.2)
		if rq != "" {
			v, line, p, eRead = parsingRightQuote(reader, []byte(rq),
				line, p)

			if eRead != nil {
				return
			}

			if sep != "" {
				p, eRead = parsingSkipSeparator([]byte(sep),
					line, p)

				if eRead != nil {
					return
				}

				// Handle multi space if separator is a single
				// space.
				if sep == " " {
					p = parsingSkipSpace(line, p)
				}
			}
		} else {
			if sep != "" {
				// Skip space at beginning if separator is a
				// single space.
				if sep == " " {
					p = parsingSkipSpace(line, p)
				}

				v, p, eRead = parsingSeparator([]byte(sep),
					line, p)

				if eRead != nil {
					return
				}

				// Handle multi space if separator is a single
				// space.
				if sep == " " {
					p = parsingSkipSpace(line, p)
				}
			} else {
				v = line[p:]
				p += len(line)
			}
		}

		if md.GetSkip() {
			continue
		}
	empty:
		r, e := tabula.NewRecordBy(string(v), md.GetType())

		if nil != e {
			msg := fmt.Sprintf("md %s: Type conversion error from %q to %s",
				md.GetName(), string(v), md.GetTypeName())

			return nil, &ReaderError{
				T:    ETypeConversion,
				Func: "ParseLine",
				What: msg,
				Line: string(line),
				Pos:  p,
				N:    0,
			}
		}

		row = append(row, r)
		rIdx++
	}

	return &row, nil
}

//
// ReadRow read one line at a time until we get one row or error when parsing the
// data.
//
func ReadRow(reader ReaderInterface, linenum int) (
	row *tabula.Row,
	line []byte,
	n int,
	eRead *ReaderError,
) {
	var e error
	n = linenum

	// Read one line, skip empty line.
	for {
		line, e = reader.ReadLine()
		n++

		if e != nil {
			goto err
		}

		// check for empty line
		linetrimed := bytes.TrimSpace(line)

		if len(linetrimed) > 0 {
			break
		}
	}

	if reader.IsTrimSpace() {
		line = bytes.TrimSpace(line)
	}

	row, eRead = ParseLine(reader, line)

	return row, line, n, eRead

err:
	eRead = &ReaderError{
		Func: "ReadRow",
		What: fmt.Sprint(e),
	}

	if e == io.EOF {
		eRead.T = EReadEOF
	} else {
		eRead.T = EReadLine
	}

	return nil, line, n, eRead
}
