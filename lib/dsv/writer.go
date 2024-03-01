// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

const (
	// DefSeparator default separator that will be used if its not given
	// in config file.
	DefSeparator = ","
	// DefOutput file.
	DefOutput = "output.dat"
	// DefEscape default string to escape the right quote or separator.
	DefEscape = "\\"
)

// Writer write records from reader or slice using format configuration in
// metadata.
type Writer struct {
	// fWriter as write descriptor.
	fWriter *os.File

	// BufWriter for buffered writer.
	BufWriter *bufio.Writer

	Config `json:"-"`

	// Output file where the records will be written.
	Output string `json:"Output"`

	// OutputMetadata define format for each column.
	OutputMetadata []Metadata `json:"OutputMetadata"`
}

// NewWriter create a writer object.
// User must call Open after that to populate the output and metadata.
func NewWriter(config string) (writer *Writer, e error) {
	writer = &Writer{
		Output:         "",
		OutputMetadata: nil,
		fWriter:        nil,
		BufWriter:      nil,
	}

	if config == "" {
		return
	}

	e = OpenWriter(writer, config)
	if e != nil {
		return nil, e
	}

	return
}

// GetOutput return output filename.
func (writer *Writer) GetOutput() string {
	return writer.Output
}

// SetOutput will set the output file to path.
func (writer *Writer) SetOutput(path string) {
	writer.Output = path
}

// AddMetadata will add new output metadata to writer.
func (writer *Writer) AddMetadata(md Metadata) {
	writer.OutputMetadata = append(writer.OutputMetadata, md)
}

// open a generic method to open output file with specific flag.
func (writer *Writer) open(file string, flag int) (e error) {
	if file == "" {
		if writer.Output == "" {
			file = DefOutput
		} else {
			file = writer.Output
		}
	}

	writer.fWriter, e = os.OpenFile(file, flag, 0600)
	if nil != e {
		return e
	}

	writer.BufWriter = bufio.NewWriter(writer.fWriter)

	return nil
}

// OpenOutput file and buffered writer.
// File will be truncated if its exist.
func (writer *Writer) OpenOutput(file string) (e error) {
	return writer.open(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
}

// ReopenOutput will open the output file back without truncating the content.
func (writer *Writer) ReopenOutput(file string) (e error) {
	if e = writer.Close(); e != nil {
		return
	}
	return writer.open(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY)
}

// Flush output buffer to disk.
func (writer *Writer) Flush() error {
	return writer.BufWriter.Flush()
}

// Close all open descriptor.
func (writer *Writer) Close() (e error) {
	if nil != writer.BufWriter {
		e = writer.BufWriter.Flush()
		if e != nil {
			return
		}
	}
	if nil != writer.fWriter {
		e = writer.fWriter.Close()
	}
	return
}

// WriteRow dump content of Row to file using format in metadata.
func (writer *Writer) WriteRow(row *tabula.Row, recordMd []MetadataInterface) (
	e error,
) {
	nRecord := row.Len()
	v := []byte{}
	esc := []byte(DefEscape)

	for i := range writer.OutputMetadata {
		md := writer.OutputMetadata[i]

		// find the input index based on name on record metadata.
		rIdx, mdMatch := FindMetadata(&md, recordMd)

		// No input metadata matched? skip it too.
		if rIdx >= nRecord {
			continue
		}

		// If input column is ignored, continue to next record.
		if mdMatch != nil && mdMatch.GetSkip() {
			continue
		}

		recV := (*row)[rIdx].Bytes()
		lq := md.GetLeftQuote()

		if lq != "" {
			v = append(v, []byte(lq)...)
		}

		rq := md.GetRightQuote()
		sep := md.GetSeparator()

		// Escape the escape character itself.
		if md.T == tabula.TString {
			recV, _ = libbytes.EncloseToken(recV, esc, esc, nil)
		}

		// Escape the right quote in field content before writing it.
		if rq != "" && md.T == tabula.TString {
			recV, _ = libbytes.EncloseToken(recV, []byte(rq), esc, nil)
		} else if sep != "" && md.T == tabula.TString {
			// Escape the separator
			recV, _ = libbytes.EncloseToken(recV, []byte(sep), esc, nil)
		}

		v = append(v, recV...)

		if rq != "" {
			v = append(v, []byte(rq)...)
		}

		if sep != "" {
			v = append(v, []byte(sep)...)
		}
	}

	v = append(v, DefEOL)

	_, e = writer.BufWriter.Write(v)

	return e
}

// WriteRows will loop each row in the list of rows and write their content to
// output file.
// Return n for number of row written, and e if error happened.
func (writer *Writer) WriteRows(rows tabula.Rows, recordMd []MetadataInterface) (
	n int,
	e error,
) {
	for n = range rows {
		e = writer.WriteRow(rows[n], recordMd)
		if nil != e {
			break
		}
	}

	_ = writer.Flush()
	return
}

// WriteColumns will write content of columns to output file.
// Return n for number of row written, and e if error happened.
func (writer *Writer) WriteColumns(columns tabula.Columns,
	colMd []MetadataInterface,
) (
	n int,
	e error,
) {
	nColumns := len(columns)
	if nColumns <= 0 {
		return
	}

	emptyRec := tabula.NewRecordString("")

	// Get minimum and maximum length of all columns.
	// In case one of the column have different length (shorter or longer),
	// we will take the column with minimum length first and continue with
	// the maximum length.

	minlen, maxlen := columns.GetMinMaxLength()

	// If metadata is nil, generate it from column name.
	if colMd == nil {
		for _, col := range columns {
			md := &Metadata{
				Name: col.Name,
				T:    col.Type,
			}

			colMd = append(colMd, md)
		}
	}

	// First loop, iterate until minimum column length.
	row := make(tabula.Row, nColumns)

	for ; n < minlen; n++ {
		// Convert columns to record.
		for y, col := range columns {
			row[y] = col.Records[n]
		}

		e = writer.WriteRow(&row, colMd)
		if e != nil {
			goto err
		}
	}

	// Second loop, iterate until maximum column length.
	for ; n < maxlen; n++ {
		// Convert columns to record.
		for y, col := range columns {
			if col.Len() > n {
				row[y] = col.Records[n]
			} else {
				row[y] = emptyRec
			}
		}

		e = writer.WriteRow(&row, colMd)
		if e != nil {
			goto err
		}
	}

err:
	_ = writer.Flush()
	return n, e
}

// WriteRawRow will write row data using separator `sep` for each record.
func (writer *Writer) WriteRawRow(row *tabula.Row, sep, esc []byte) (e error) {
	if sep == nil {
		sep = []byte(DefSeparator)
	}
	if esc == nil {
		esc = []byte(DefEscape)
	}

	v := []byte{}
	for x, rec := range *row {
		if x > 0 {
			v = append(v, sep...)
		}

		recV := rec.Bytes()

		if rec.Type() == tabula.TString {
			recV, _ = libbytes.EncloseToken(recV, sep, esc, nil)
		}

		v = append(v, recV...)
	}

	v = append(v, DefEOL)

	_, e = writer.BufWriter.Write(v)

	_ = writer.Flush()

	return e
}

// WriteRawRows write rows data using separator `sep` for each record.
// We use pointer in separator parameter, so we can use empty string as
// separator.
func (writer *Writer) WriteRawRows(rows *tabula.Rows, sep *string) (
	nrow int,
	e error,
) {
	nrow = len(*rows)
	if nrow <= 0 {
		return
	}

	if sep == nil {
		sep = new(string)
		*sep = DefSeparator
	}

	escbytes := []byte(DefEscape)
	sepbytes := []byte(*sep)
	x := 0

	for ; x < nrow; x++ {
		e = writer.WriteRawRow((*rows)[x], sepbytes, escbytes)
		if nil != e {
			break
		}
	}

	return x, e
}

// WriteRawColumns write raw columns using separator `sep` for each record to
// file.
//
// We use pointer in separator parameter, so we can use empty string as
// separator.
func (writer *Writer) WriteRawColumns(cols *tabula.Columns, sep *string) (
	nrow int,
	e error,
) {
	ncol := len(*cols)
	if ncol <= 0 {
		return
	}

	if sep == nil {
		sep = new(string)
		*sep = DefSeparator
	}

	// Find minimum and maximum column length.
	minlen, maxlen := cols.GetMinMaxLength()

	esc := []byte(DefEscape)
	sepbytes := []byte(*sep)
	x := 0

	// First, write until minimum column length.
	for ; x < minlen; x++ {
		v := cols.Join(x, sepbytes, esc)
		v = append(v, DefEOL)

		_, e = writer.BufWriter.Write(v)

		if nil != e {
			return x, e
		}
	}

	// and then write column until max length.
	for ; x < maxlen; x++ {
		v := cols.Join(x, sepbytes, esc)
		v = append(v, DefEOL)

		_, e = writer.BufWriter.Write(v)

		if nil != e {
			break
		}
	}

	_ = writer.Flush()
	return x, e
}

// WriteRawDataset will write content of dataset to file without metadata but
// using separator `sep` for each record.
//
// We use pointer in separator parameter, so we can use empty string as
// separator.
func (writer *Writer) WriteRawDataset(dataset tabula.DatasetInterface,
	sep *string,
) (
	int, error,
) {
	if nil == writer.fWriter {
		return 0, ErrNotOpen
	}
	if nil == dataset {
		return 0, nil
	}
	if sep == nil {
		sep = new(string)
		*sep = DefSeparator
	}

	var rows *tabula.Rows

	switch dataset.GetMode() {
	case tabula.DatasetModeColumns:
		cols := dataset.GetDataAsColumns()
		return writer.WriteRawColumns(cols, sep)
	case tabula.DatasetModeRows, tabula.DatasetModeMatrix:
		fallthrough
	default:
		rows = dataset.GetDataAsRows()
	}

	return writer.WriteRawRows(rows, sep)
}

// Write rows from Reader to file.
// Return n for number of row written, or e if error happened.
func (writer *Writer) Write(reader ReaderInterface) (int, error) {
	if nil == reader {
		return 0, ErrNilReader
	}
	if nil == writer.fWriter {
		return 0, ErrNotOpen
	}

	ds := reader.GetDataset().(tabula.DatasetInterface)

	var rows *tabula.Rows

	switch ds.GetMode() {
	case tabula.DatasetModeColumns:
		cols := ds.GetDataAsColumns()
		return writer.WriteColumns(*cols, reader.GetInputMetadata())
	case tabula.DatasetModeRows, tabula.DatasetModeMatrix:
		fallthrough
	default:
		rows = ds.GetDataAsRows()
	}

	return writer.WriteRows(*rows, reader.GetInputMetadata())
}

// String yes, it will print it in JSON like format.
func (writer *Writer) String() string {
	r, e := json.MarshalIndent(writer, "", "\t")

	if nil != e {
		log.Print(e)
	}

	return string(r)
}
