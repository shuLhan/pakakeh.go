// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode"

	"github.com/shuLhan/share/lib/debug"
)

const (
	tokBackslash   = '\\'
	tokBackspace   = '\b'
	tokDot         = '.'
	tokDoubleQuote = '"'
	tokEqual       = '='
	tokHash        = '#'
	tokHyphen      = '-'
	tokNewLine     = '\n'
	tokSecEnd      = ']'
	tokSecStart    = '['
	tokSemiColon   = ';'
	tokSpace       = ' '
	tokTab         = '\t'
	tokUnderscore  = '_'
)

var (
	errBadConfig      = errors.New("bad config line %d at %s")
	errVarNoSection   = "variable without section, line %d at %s"
	errVarNameInvalid = errors.New("invalid variable name, line %d at %s")
	errValueInvalid   = errors.New("invalid value, line %d at %s")
)

// reader define the INI file reader.
type reader struct {
	br   *bytes.Reader
	_var *variable
	sec  *Section

	filename string

	buf        bytes.Buffer
	bufComment bytes.Buffer
	bufFormat  bytes.Buffer
	bufSpaces  bytes.Buffer

	lineNum int
	r       rune
	b       byte
}

// newReader create, initialize, and return new reader.
func newReader() (r *reader) {
	r = &reader{
		br: bytes.NewReader(nil),
	}
	r.reset(nil)

	return
}

// reset all reader attributes, excluding filename.
func (reader *reader) reset(src []byte) {
	reader.br.Reset(src)
	reader.b = 0
	reader.r = 0
	reader.lineNum = 0
	reader._var = &variable{
		mode: lineModeEmpty,
	}
	reader.sec = &Section{
		mode: lineModeEmpty,
	}
	reader.buf.Reset()
	reader.bufComment.Reset()
	reader.bufFormat.Reset()
	reader.bufSpaces.Reset()
}

// parseFile will open, read, and parse INI file `filename` and return an
// instance of Ini.
//
// On failure, it return nil and error.
func (reader *reader) parseFile(filename string) (in *Ini, err error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	reader.filename = filename

	in, err = reader.Parse(src)

	return
}

// Parse will parse INI config from slice of bytes `src` into `in`.
func (reader *reader) Parse(src []byte) (in *Ini, err error) {
	in = &Ini{}
	reader.reset(src)

	for {
		reader.lineNum++

		err = reader.parse()
		if err != nil {
			if err != io.EOF {
				err = fmt.Errorf(err.Error(), reader.lineNum,
					reader.filename)
				return nil, err
			}
			break
		}

		if debug.Value >= 3 {
			fmt.Print(reader._var)
		}

		reader._var.lineNum = reader.lineNum

		if isLineModeVar(reader._var.mode) {
			if reader.sec.mode == lineModeEmpty {
				err = fmt.Errorf(errVarNoSection,
					reader.lineNum,
					reader.filename)
				return nil, err
			}
		}

		if reader._var.mode&lineModeSection == lineModeSection ||
			reader._var.mode&lineModeSubsection == lineModeSubsection {
			in.addSection(reader.sec)

			reader.sec = &Section{
				mode:    reader._var.mode,
				lineNum: reader._var.lineNum,
				name:    reader._var.secName,
				sub:     reader._var.subName,
				format:  reader._var.format,
				others:  reader._var.others,
			}

			reader._var = &variable{
				mode: lineModeEmpty,
			}
			continue
		}

		reader.sec.addVariable(reader._var)

		reader._var = &variable{
			mode: lineModeEmpty,
		}
	}

	if reader._var.mode != lineModeEmpty {
		if debug.Value >= 3 {
			fmt.Println(reader._var)
		}

		reader.sec.addVariable(reader._var)
	}

	in.addSection(reader.sec)

	reader._var = nil
	reader.sec = nil

	return in, nil
}

func (reader *reader) parse() (err error) {
	var isNewline bool

	reader.bufFormat.Reset()

	for !isNewline {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			return err
		}
		switch reader.b {
		case tokNewLine:
			reader.bufFormat.WriteByte(reader.b)
			reader._var.format = reader.bufFormat.String()
			isNewline = true

		case tokSpace, tokTab:
			reader.bufFormat.WriteByte(reader.b)

		case tokHash, tokSemiColon:
			_ = reader.br.UnreadByte()
			return reader.parseComment()

		case tokSecStart:
			_ = reader.br.UnreadByte()
			return reader.parseSectionHeader()

		default:
			_ = reader.br.UnreadByte()
			return reader.parseVariable()
		}
	}

	return nil
}

func (reader *reader) parseComment() (err error) {
	reader.bufComment.Reset()

	reader._var.mode |= lineModeComment

	reader.bufFormat.Write([]byte{'%', 's'})

	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}
		if reader.b == tokNewLine {
			reader.bufFormat.WriteByte(reader.b)
			break
		}
		_ = reader.bufComment.WriteByte(reader.b)
	}

	reader._var.format = reader.bufFormat.String()
	reader._var.others = reader.bufComment.String()

	return
}

func (reader *reader) parseSectionHeader() (err error) {
	reader.buf.Reset()

	reader.b, err = reader.br.ReadByte()
	if err != nil {
		return errBadConfig
	}

	if reader.b != tokSecStart {
		return errBadConfig
	}

	reader.bufFormat.WriteByte(tokSecStart)
	reader._var.mode = lineModeSection

	reader.r, _, err = reader.br.ReadRune()
	if err != nil {
		return errBadConfig
	}

	if !unicode.IsLetter(reader.r) {
		return errBadConfig
	}

	var isNewline bool
	reader.bufFormat.Write([]byte{'%', 's'})
	reader.buf.WriteRune(reader.r)

	for !isNewline {
		reader.r, _, err = reader.br.ReadRune()
		if err != nil {
			return errBadConfig
		}
		switch {
		case reader.r == tokSpace, reader.r == tokTab:
			isNewline = true

		case reader.r == tokSecEnd:
			reader.bufFormat.WriteRune(reader.r)
			reader._var.secName = reader.buf.String()
			return reader.parsePossibleComment()

		case unicode.IsLetter(reader.r), unicode.IsDigit(reader.r),
			reader.r == tokHyphen, reader.r == tokDot:
			reader.buf.WriteRune(reader.r)

		default:
			return errBadConfig
		}
	}

	reader.bufFormat.WriteRune(reader.r)
	reader._var.secName = reader.buf.String()

	return reader.parseSubsection()
}

func (reader *reader) parseSubsection() (err error) {
	reader.buf.Reset()

	reader._var.mode |= lineModeSubsection

	// Skip white-spaces
	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			return errBadConfig
		}
		if reader.b == tokSpace || reader.b == tokTab {
			reader.bufFormat.WriteByte(reader.b)
			continue
		}
		if reader.b != tokDoubleQuote {
			return errBadConfig
		}
		break
	}

	reader.bufFormat.WriteByte(reader.b) // == tokDoubleQuote
	reader.bufFormat.Write([]byte{'%', 's'})

	var esc bool
	var end bool

	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			return errBadConfig
		}
		if end {
			if reader.b == tokSecEnd {
				reader.bufFormat.WriteByte(reader.b)
				break
			}
			return errBadConfig
		}
		if esc {
			reader.buf.WriteByte(reader.b)
			esc = false
			continue
		}
		if reader.b == tokBackslash {
			esc = true
			continue
		}
		if reader.b == tokDoubleQuote {
			reader.bufFormat.WriteByte(reader.b)
			end = true
			continue
		}
		reader.buf.WriteByte(reader.b)
	}

	reader._var.subName = reader.buf.String()

	return reader.parsePossibleComment()
}

// parsePossibleComment will check only for whitespace and comment start
// character.
func (reader *reader) parsePossibleComment() (err error) {
	var isNewline bool

	for !isNewline {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}
		switch reader.b {
		case tokNewLine:
			reader.bufFormat.WriteByte(reader.b)
			isNewline = true
		case tokSpace, tokTab:
			reader.bufFormat.WriteByte(reader.b)
		case tokHash, tokSemiColon:
			_ = reader.br.UnreadByte()
			return reader.parseComment()

		default:
			return errBadConfig
		}
	}

	reader._var.format = reader.bufFormat.String()

	return
}

func (reader *reader) parseVariable() (err error) {
	reader.buf.Reset()

	reader.r, _, err = reader.br.ReadRune()
	if err != nil {
		return errVarNameInvalid
	}

	if !unicode.IsLetter(reader.r) {
		return errVarNameInvalid
	}

	var isNewline bool
	reader.bufFormat.Write([]byte{'%', 's'})
	reader.buf.WriteRune(reader.r)

	for !isNewline {
		reader.r, _, err = reader.br.ReadRune()
		if err != nil {
			break
		}
		switch {
		case reader.r == tokEqual:
			reader.bufFormat.WriteRune(reader.r)

			reader._var.mode = lineModeValue
			reader._var.key = reader.buf.String()

			return reader.parseVarValue()

		case reader.r == tokNewLine:
			reader.bufFormat.WriteRune(reader.r)
			isNewline = true

		case unicode.IsLetter(reader.r), unicode.IsDigit(reader.r),
			reader.r == tokHyphen, reader.r == tokDot,
			reader.r == tokUnderscore:
			reader.buf.WriteRune(reader.r)

		case reader.r == tokHash, reader.r == tokSemiColon:
			_ = reader.br.UnreadRune()

			reader._var.mode = lineModeValue
			reader._var.key = reader.buf.String()

			return reader.parseComment()

		case unicode.IsSpace(reader.r):
			reader.bufFormat.WriteRune(reader.r)

			reader._var.mode = lineModeValue
			reader._var.key = reader.buf.String()

			return reader.parsePossibleValue()

		default:
			return errVarNameInvalid
		}
	}

	reader._var.mode = lineModeValue
	reader._var.format = reader.bufFormat.String()
	reader._var.key = reader.buf.String()

	return nil
}

// parsePossibleValue will check if the next character after space is comment
// or `=`.
func (reader *reader) parsePossibleValue() (err error) {
	var isNewline bool
	for !isNewline {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}
		switch reader.b {
		case tokNewLine:
			reader.bufFormat.WriteByte(reader.b)
			isNewline = true

		case tokSpace, tokTab:
			reader.bufFormat.WriteByte(reader.b)

		case tokHash, tokSemiColon:
			_ = reader.br.UnreadByte()
			return reader.parseComment()
		case tokEqual:
			reader.bufFormat.WriteByte(reader.b)
			return reader.parseVarValue()
		default:
			return errVarNameInvalid
		}
	}

	reader._var.mode = lineModeValue
	reader._var.format = reader.bufFormat.String()

	return nil
}

// At this point we found `=` on source, and we expect the rest of source will
// be variable value.
func (reader *reader) parseVarValue() (err error) {
	reader.buf.Reset()
	reader.bufSpaces.Reset()

	// Consume leading white-spaces.
consume_spaces:
	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			reader._var.format = reader.bufFormat.String()
			reader._var.value = ""
			return err
		}
		switch reader.b {
		case tokSpace, tokTab:
			reader.bufFormat.WriteByte(reader.b)
			continue consume_spaces

		case tokHash, tokSemiColon:
			_ = reader.br.UnreadByte()
			reader._var.value = ""
			reader.bufFormat.WriteString("%s")
			return reader.parseComment()

		case tokNewLine:
			if len(reader._var.key) > 0 {
				reader.bufFormat.WriteString("%s")
			}
			reader.bufFormat.WriteByte(reader.b)
			reader._var.format = reader.bufFormat.String()
			reader._var.value = ""
			return nil
		}
		break
	}

	reader.bufFormat.Write([]byte{'%', 's'})
	reader._var.mode = lineModeValue
	_ = reader.br.UnreadByte()

	var (
		quoted    bool
		esc       bool
		isNewline bool
	)

	for !isNewline {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}

		if esc {
			switch reader.b {
			case tokNewLine:
				reader._var.mode = lineModeMulti
				reader.valueCommit(true)
				reader.lineNum++
				esc = false
				continue

			case tokBackslash, tokDoubleQuote:
				reader.valueWriteByte(reader.b)
				esc = false
				continue

			case 'b':
				reader.buf.WriteByte(tokBackspace)
				esc = false
				continue

			case 'n':
				reader.buf.WriteByte(tokNewLine)
				esc = false
				continue
			case 't':
				reader.buf.WriteByte(tokTab)
				esc = false
				continue
			}
			return errValueInvalid
		}

		switch reader.b {
		case tokSpace, tokTab:
			if quoted {
				reader.valueWriteByte(reader.b)
				continue
			}
			reader.bufSpaces.WriteByte(reader.b)

		case tokBackslash:
			esc = true

		case tokDoubleQuote:
			if quoted {
				quoted = false
			} else {
				reader._var.isQuoted = true
				quoted = true
			}

		case tokNewLine:
			reader.bufFormat.WriteByte(reader.b)
			isNewline = true

		case tokHash, tokSemiColon:
			if quoted {
				reader.valueWriteByte(reader.b)
				continue
			}

			reader.bufFormat.Write(reader.bufSpaces.Bytes())
			reader.valueCommit(false)

			_ = reader.br.UnreadByte()

			return reader.parseComment()

		default:
			reader.valueWriteByte(reader.b)
		}
	}

	if quoted {
		return errValueInvalid
	}

	reader.valueCommit(false)

	reader._var.format = reader.bufFormat.String()

	return nil
}

func (reader *reader) valueCommit(withSpaces bool) {
	val := reader.buf.String()

	if withSpaces {
		val += reader.bufSpaces.String()
	}

	reader._var.value += val

	reader.buf.Reset()
	reader.bufSpaces.Reset()
}

func (reader *reader) valueWriteByte(b byte) {
	if reader.bufSpaces.Len() > 0 {
		reader.buf.Write(reader.bufSpaces.Bytes())
		reader.bufSpaces.Reset()
	}

	reader.buf.WriteByte(b)
}
