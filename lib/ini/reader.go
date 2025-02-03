// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package ini

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
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

	buf       bytes.Buffer
	bufFormat bytes.Buffer

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
	reader.bufFormat.Reset()
}

// Parse will parse INI config from slice of bytes `src` into `in`.
func (reader *reader) Parse(src []byte) (in *Ini, err error) {
	in = &Ini{}
	reader.reset(src)

	for {
		reader.lineNum++

		err = reader.parse()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				err = fmt.Errorf(err.Error(), reader.lineNum,
					reader.filename)
				return nil, err
			}
			break
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
		reader.sec.addVariable(reader._var)
	}

	in.addSection(reader.sec)

	reader._var = nil
	reader.sec = nil

	return in, nil
}

func (reader *reader) parse() (err error) {
	reader.bufFormat.Reset()

	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}
		switch reader.b {
		case tokNewLine:
			reader.bufFormat.WriteByte(reader.b)
			reader._var.format = reader.bufFormat.String()
			return nil

		case tokSpace, tokTab:
			reader.bufFormat.WriteByte(reader.b)

		case tokHash, tokSemiColon:
			reader.bufFormat.WriteByte(reader.b)
			return reader.parseComment()

		case tokSecStart:
			return reader.parseSectionHeader()

		default:
			_ = reader.br.UnreadByte()
			return reader.parseVariable()
		}
	}
	return err
}

func (reader *reader) parseComment() (err error) {
	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}
		_ = reader.bufFormat.WriteByte(reader.b)
		if reader.b == tokNewLine {
			break
		}
	}

	reader._var.format = reader.bufFormat.String()

	return
}

func (reader *reader) parseSectionHeader() (err error) {
	reader.buf.Reset()

	reader.bufFormat.WriteByte(tokSecStart)
	reader._var.mode = lineModeSection

	reader.r, _, err = reader.br.ReadRune()
	if err != nil {
		return errBadConfig
	}

	if !unicode.IsLetter(reader.r) {
		return errBadConfig
	}

	reader.bufFormat.WriteString("%s")
	reader.buf.WriteRune(reader.r)

	for {
		reader.r, _, err = reader.br.ReadRune()
		if err != nil {
			return errBadConfig
		}
		if reader.r == tokSpace || reader.r == tokTab {
			break
		}
		if reader.r == tokSecEnd {
			reader.bufFormat.WriteRune(reader.r)
			reader._var.secName = reader.buf.String()
			return reader.parsePossibleComment()
		}
		if unicode.IsLetter(reader.r) || unicode.IsDigit(reader.r) ||
			reader.r == tokHyphen || reader.r == tokDot {
			reader.buf.WriteRune(reader.r)
			continue
		}
		return errBadConfig
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

// parsePossibleComment parse possible comment at the end of section or
// variable.
func (reader *reader) parsePossibleComment() (err error) {
	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}
		if reader.b == tokNewLine {
			reader.bufFormat.WriteByte(reader.b)
			break
		}
		if reader.b == tokSpace || reader.b == tokTab {
			reader.bufFormat.WriteByte(reader.b)
			continue
		}
		if reader.b == tokHash || reader.b == tokSemiColon {
			reader.bufFormat.WriteByte(reader.b)
			return reader.parseComment()
		}
		return errBadConfig
	}

	reader._var.format = reader.bufFormat.String()

	return err
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
			// The only possible error here is [io.EOF], so we
			// end it.
			break
		}
		switch {
		case reader.r == tokEqual:
			reader.bufFormat.WriteRune(reader.r)

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
			reader._var.mode = lineModeKeyOnly
			reader._var.key = reader.buf.String()

			reader.bufFormat.WriteRune(reader.r)
			return reader.parseComment()

		case unicode.IsSpace(reader.r):
			reader.bufFormat.WriteRune(reader.r)

			reader._var.mode = lineModeKeyOnly
			reader._var.key = reader.buf.String()

			return reader.parsePossibleValue()

		default:
			return errVarNameInvalid
		}
	}

	reader._var.mode = lineModeKeyOnly
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
			// The only possible error here is [io.EOF], so we
			// end it.
			break
		}
		switch reader.b {
		case tokNewLine:
			reader.bufFormat.WriteByte(reader.b)
			isNewline = true

		case tokSpace, tokTab:
			reader.bufFormat.WriteByte(reader.b)

		case tokHash, tokSemiColon:
			reader.bufFormat.WriteByte(reader.b)
			return reader.parseComment()
		case tokEqual:
			reader.bufFormat.WriteByte(reader.b)
			return reader.parseVarValue()
		default:
			return errVarNameInvalid
		}
	}

	reader._var.mode = lineModeKeyOnly
	reader._var.format = reader.bufFormat.String()

	return nil
}

// At this point we found `=` on source, and we expect the rest of source will
// be variable value.
// This method consume all characters after '=' as rawValue and normalize it
// into value.
func (reader *reader) parseVarValue() (err error) {
	reader.buf.Reset()

	var (
		isQuoted bool
		isEsc    bool
	)

	reader.bufFormat.WriteString("%s")
	reader._var.mode = lineModeKeyValue

	for {
		reader.b, err = reader.br.ReadByte()
		if err != nil {
			break
		}
		if isEsc {
			if reader.b == tokNewLine {
				reader.buf.WriteByte(tokBackslash)
				reader.buf.WriteByte(tokNewLine)
				reader.lineNum++
				isEsc = false
				continue
			}
			if reader.b == tokBackslash ||
				reader.b == tokDoubleQuote ||
				reader.b == 'b' || reader.b == 'n' ||
				reader.b == 't' {
				reader.buf.WriteByte(tokBackslash)
				reader.buf.WriteByte(reader.b)
				isEsc = false
				continue
			}
			return errValueInvalid
		}
		if reader.b == tokDoubleQuote {
			reader.buf.WriteByte(reader.b)
			isQuoted = !isQuoted
			continue
		}
		if reader.b == tokBackslash {
			isEsc = true
			continue
		}
		if reader.b == tokNewLine {
			reader.bufFormat.WriteByte(tokNewLine)
			break
		}
		if reader.b == tokHash || reader.b == tokSemiColon {
			if isQuoted {
				reader.buf.WriteByte(reader.b)
				continue
			}

			reader.bufFormat.WriteByte(reader.b)

			reader._var.rawValue = bytes.Clone(reader.buf.Bytes())
			reader._var.value = parseRawValue(reader._var.rawValue)

			return reader.parseComment()
		}
		reader.buf.WriteByte(reader.b)
	}

	if isQuoted {
		return errValueInvalid
	}

	reader._var.rawValue = bytes.Clone(reader.buf.Bytes())
	reader._var.value = parseRawValue(reader._var.rawValue)
	reader._var.format = reader.bufFormat.String()

	return nil
}

// parseRawValue parse the multiline and double quoted raw value into single
// string.
func parseRawValue(raw []byte) (out string) {
	var (
		sb strings.Builder
		b  byte

		isEsc       bool
		isPrevSpace bool
		isQuoted    bool
	)

	raw = bytes.TrimSpace(raw)

	for _, b = range raw {
		if b == ' ' || b == '\t' {
			if isQuoted {
				sb.WriteByte(b)
				continue
			}
			if isPrevSpace {
				continue
			}
			if sb.Len() > 0 {
				// Only write space once if sb already
				// filled.
				sb.WriteByte(' ')
			}
			isPrevSpace = true
			continue
		}
		if isEsc {
			switch b {
			case 'b':
				sb.WriteByte(tokBackspace)
			case 'n':
				sb.WriteByte(tokNewLine)
			case 't':
				sb.WriteByte(tokTab)
			case '\\':
				sb.WriteByte(tokBackslash)
			case '"':
				sb.WriteByte(tokDoubleQuote)
			}
			isEsc = false
			continue
		}
		if b == tokBackslash {
			isEsc = true
			continue
		}
		if b == tokDoubleQuote {
			isQuoted = !isQuoted
			continue
		}
		sb.WriteByte(b)
		isPrevSpace = false
	}
	return sb.String()
}
