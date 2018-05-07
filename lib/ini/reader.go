package ini

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"unicode"
)

const (
	tokBackslash   = '\\'
	tokHash        = '#'
	tokSecEnd      = ']'
	tokSecStart    = '['
	tokSemiColon   = ';'
	tokDoubleQuote = '"'
)

var (
	errBadConfig      = "bad config line %d at %s"
	errVarNoSection   = "variable without section, line %d at %s"
	errVarNameInvalid = "invalid variable name, line %d at %s"
	errValueInvalid   = "invalid value, line %d at %s"

	sepSubsection = []byte{' '}
	sepNewline    = []byte{'\n'}
	sepVar        = []byte{'='}
)

//
// Reader define the INI file reader.
//
type Reader struct {
	filename  string
	bb        []byte
	lines     []parsedLine
	sec       *section
	buf       bytes.Buffer
	bufSpaces bytes.Buffer
	bufCom    bytes.Buffer
}

//
// NewReader will open file `filename` for reading and return the reader
// without error.
// On fail, it will return nil reader and error.
//
func NewReader(filename string) (reader *Reader, err error) {
	reader = &Reader{
		filename: filename,
		sec: &section{
			m: sectionModeNone,
		},
	}

	reader.bb, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	return
}

//
// parse will parse INI config from slice of bytes into `in`.
//
// nolint: gocyclo
func (reader *Reader) parse(in *Ini) (err error) {
	var ok bool

	err = reader.normalized()
	if err != nil {
		return
	}

	for x := 0; x < len(reader.lines); x++ {
		switch reader.lines[x].m {
		case lineModeNewline:
			reader.sec.pushVar(varModeNewline, nil, nil, nil)

		case lineModeComment:
			reader.sec.pushVar(varModeComment, nil, nil,
				reader.lines[x].v)

		case lineModeVar:
			// S.4.0 variable must belong to section
			if reader.sec.m == sectionModeNone {
				err = fmt.Errorf(errVarNoSection, x,
					reader.filename)
				return
			}

			err = reader.parseVar(reader.lines[x].v, x)

		case lineModeVarMulti:
			// S.4.0 variable must belong to section
			if reader.sec.m == sectionModeNone {
				err = fmt.Errorf(errVarNoSection, x,
					reader.filename)
				return
			}

			x, err = reader.parseMultilineVar(x)
			x--

		case lineModeSection:
			in.secs = append(in.secs, reader.sec)
			reader.sec = &section{
				m: sectionModeNormal,
			}
			ok = reader.parseSection(reader.lines[x].v, x, true)
			if !ok {
				err = fmt.Errorf(errBadConfig, x,
					reader.filename)
			}

		case lineModeSubsection:
			in.secs = append(in.secs, reader.sec)
			reader.sec = &section{
				m: sectionModeSub,
			}
			ok = reader.parseSubsection(reader.lines[x].v, x)
			if !ok {
				err = fmt.Errorf(errBadConfig, x,
					reader.filename)
			}
		}

		if err != nil {
			return
		}
	}

	return
}

//
// normalized will split source by lines.
//
// nolint: gocyclo
func (reader *Reader) normalized() (err error) {
	// (0)
	multi := false
	lines := bytes.Split(reader.bb, sepNewline)

	for x := 0; x < len(lines); x++ {
		orgLine := lines[x]
		line := bytes.TrimSpace(orgLine)

		if len(line) == 0 {
			multi = false
			reader.addLine(lineModeNewline, x, nil)
			continue
		}

		b0 := line[0]
		blast := line[len(line)-1]

		if multi {
			reader.addLine(lineModeVarMulti, x, line)

			if blast != tokBackslash {
				multi = false
			}
			continue
		}

		if b0 == tokHash || b0 == tokSemiColon {
			reader.addLine(lineModeComment, x, orgLine)
			continue
		}

		if b0 == tokSecStart {
			if blast != tokSecEnd {
				err = fmt.Errorf(errBadConfig, x+1,
					reader.filename)
				return
			}

			reader.addLine(lineModeSection, x, line)
			continue
		}

		if blast != tokBackslash {
			reader.addLine(lineModeVar, x, line)
			continue
		}

		reader.addLine(lineModeVarMulti, x, line)
		multi = true
	}

	if debug >= debugL2 {
		for _, line := range reader.lines {
			fmt.Printf("%1d %4d %s\n", line.m, line.n, line.v)
		}
	}

	return
}

//
// addLine will add input line `in` to reader.
//
// (1) If line mode is section,
// (1.1) If it's contain space, change their mode to subsection.
//
func (reader *Reader) addLine(mode lineMode, num int, in []byte) {
	// (1)
	if mode == lineModeSection {
		itHaveSub := bytes.Index(in, sepSubsection)
		if itHaveSub > 0 {
			mode = lineModeSubsection
		}
	}

	line := parsedLine{
		m: mode,
		n: num,
		v: in,
	}

	reader.lines = append(reader.lines, line)
}

func (reader *Reader) parseMultilineVar(start int) (end int, err error) {
	var (
		lastIdx int
		blast   byte
	)

	reader.buf.Reset()

	for end = start; end < len(reader.lines); end++ {
		if reader.lines[end].m != lineModeVarMulti {
			break
		}

		lastIdx = len(reader.lines[end].v) - 1
		blast = reader.lines[end].v[lastIdx]
		if blast == tokBackslash {
			reader.buf.Write(reader.lines[end].v[0:lastIdx])
		} else {
			reader.buf.Write(reader.lines[end].v)
		}
	}

	err = reader.parseVar(reader.buf.Bytes(), start)

	return
}

//
// parseVar will split line at line number `num` into key and value
// using `=` as separator.
//
// (S.5.4) Variable name without value is a short-hand to set the value to the
//         boolean "true".
//
func (reader *Reader) parseVar(line []byte, num int) (err error) {
	var v, comment []byte

	kv := bytes.SplitN(line, sepVar, 2)

	k, ok := reader.parseVarName(kv[0])
	if !ok {
		err = fmt.Errorf(errVarNameInvalid, num, reader.filename)
		return
	}

	// (S.5.4)
	if len(kv) == 1 {
		v = varValueTrue
	} else {
		v, comment, ok = reader.parseVarValue(kv[1])
		if !ok {
			err = fmt.Errorf(errValueInvalid, num, reader.filename)
			return
		}
	}

	reader.sec.pushVar(varModeNormal, k, v, comment)

	return
}

//
// parseVarName will parse variable name from input bytes as defined in rules
// S.5.
//
func (reader *Reader) parseVarName(in []byte) (out []byte, ok bool) {
	in = bytes.ToLower(bytes.TrimSpace(in))

	if len(in) == 0 {
		return
	}

	x := 0
	rr := bytes.Runes(in)

	if !unicode.IsLetter(rr[x]) {
		return
	}

	reader.buf.Reset()
	reader.buf.WriteRune(rr[x])

	for x++; x < len(rr); x++ {
		if rr[x] == '-' {
			reader.buf.WriteRune(rr[x])
			continue
		}
		if unicode.IsLetter(rr[x]) || unicode.IsDigit(rr[x]) {
			reader.buf.WriteRune(rr[x])
			continue
		}

		return
	}

	out = append(out, reader.buf.Bytes()...)
	ok = true

	return
}

//
// parseVarValue will parse variable value as defined in rules S.6.
//
// (0) Check for double-quote on the first rune.
//
// (1) If rune is space ' ' or tab '\t',
// (1.1) If `quoted`, write to buffer
// (1.2) If not `quoted`, write to whitespaces buffer, to be used later.
//
// (2) If rune is double-quote, reset quoted state, do not append the
//     quoted character.
//
// (3) If next rune is '#',
// (3.1) If we are on double-quoted, add it to buffer.
// (3.2) If we are not on double-quoted, the rest of must be comment
//
// (4) If `esc` is true, check if next rune is valid escaped character,
// otherwise return.
//
// (5) If next rune is '\',
// (5.1) If `quoted` is true, set `esc` to true and continue to the next rune.
// (5.3) If not `quoted`, return immediately.
//
// nolint: gocyclo
func (reader *Reader) parseVarValue(in []byte) (value, comment []byte, ok bool) {
	in = bytes.TrimSpace(in)

	// S.6.0
	if len(in) == 0 {
		value = varValueTrue
		ok = true
		return
	}

	var (
		quoted bool
		esc    bool
		x      int
	)

	rr := bytes.Runes(in)

	// (0)
	if rr[x] == tokDoubleQuote {
		quoted = true
		x++
	}

	reader.buf.Reset()
	reader.bufSpaces.Reset()
	reader.bufCom.Reset()

	for ; x < len(rr); x++ {
		if rr[x] == ' ' || rr[x] == '\t' {
			if quoted {
				_, _ = reader.buf.WriteRune(rr[x])
				continue
			}
			if reader.buf.Len() > 0 {
				reader.bufSpaces.WriteRune(rr[x])
			}
			continue
		}

		// (2)
		if rr[x] == tokDoubleQuote {
			if esc {
				if reader.bufSpaces.Len() > 0 {
					_, _ = reader.buf.Write(reader.bufSpaces.Bytes())
					reader.bufSpaces.Reset()
				}
				_, _ = reader.buf.WriteRune('"')
				esc = false
				continue
			}
			if quoted {
				if esc {
					_, _ = reader.buf.WriteRune('"')
					esc = false
					continue
				}
				if reader.bufSpaces.Len() > 0 {
					_, _ = reader.buf.Write(reader.bufSpaces.Bytes())
					reader.bufSpaces.Reset()
				}
				quoted = false
				continue
			}
			quoted = true
			continue
		}

		// (3)
		if rr[x] == tokHash || rr[x] == tokSemiColon {
			if quoted {
				if reader.bufSpaces.Len() > 0 {
					_, _ = reader.buf.Write(reader.bufSpaces.Bytes())
					reader.bufSpaces.Reset()
				}
				_, _ = reader.buf.WriteRune(rr[x])
				continue
			}

			if reader.bufSpaces.Len() > 0 {
				_, _ = reader.bufCom.Write(reader.bufSpaces.Bytes())
				reader.bufSpaces.Reset()
			}
			reader.bufCom.WriteString(string(rr[x:]))
			goto out
		}

		// (4)
		if esc {
			if rr[x] == 'n' || rr[x] == 't' || rr[x] == 'b' {
				_, _ = reader.buf.WriteRune(tokBackslash)
				_, _ = reader.buf.WriteRune(rr[x])
				esc = false
				continue
			}
			if rr[x] == '\\' {
				_, _ = reader.buf.WriteRune(rr[x])
				esc = false
				continue
			}
			return
		}

		// (5)
		if rr[x] == tokBackslash {
			if quoted {
				if reader.bufSpaces.Len() > 0 {
					_, _ = reader.buf.Write(reader.bufSpaces.Bytes())
					reader.bufSpaces.Reset()
				}
				esc = true
				continue
			}
			esc = true
			continue
		}

		if reader.bufSpaces.Len() > 0 {
			_, _ = reader.buf.Write(reader.bufSpaces.Bytes())
			reader.bufSpaces.Reset()
		}
		_, _ = reader.buf.WriteRune(rr[x])
	}

	if quoted || esc {
		return
	}
out:
	value = append(value, reader.buf.Bytes()...)
	comment = append(comment, reader.bufCom.Bytes()...)
	ok = true

	return
}

//
// parseSection will parse section name from line. Line is assumed to be a
// valid section, which is started with '[' and end with ']'.
//
// (0) Remove '[' and ']'
// (1) Section name must start with alphabetic character.
// (2) Section name must be alphanumeric, '-', or '.'.
//
func (reader *Reader) parseSection(line []byte, num int, trim bool) (ok bool) {
	// (0)
	if trim {
		line = bytes.TrimSpace(line[1 : len(line)-1])
	}
	if len(line) == 0 {
		return
	}

	x := 0
	runes := bytes.Runes(line)

	if !unicode.IsLetter(runes[x]) {
		return
	}

	reader.buf.Reset()

	for ; x < len(runes); x++ {
		if runes[x] == '-' || runes[x] == '.' {
			reader.buf.WriteRune(runes[x])
			continue
		}
		if unicode.IsLetter(runes[x]) || unicode.IsDigit(runes[x]) {
			reader.buf.WriteRune(runes[x])
			continue
		}

		return
	}

	ok = true

	reader.sec.name = nil
	reader.sec.name = append(reader.sec.name, reader.buf.Bytes()...)

	return
}

//
// parseSubsection will parse section name and subsection name from line. Line
// is assumed to be a valid section, which is started with '[' and end with
// ']'.
//
// (0) Remove '[' and ']'
// (1) Section and subsection is separated by single space ' '.
// (2) Subsection name enclosed by double-quote.
// (3) Subsection can contains only the following escape character: '\' and
// '"', other than that will be appended without '\' character.
//
func (reader *Reader) parseSubsection(line []byte, num int) (ok bool) {
	// (0)
	line = bytes.TrimSpace(line[1 : len(line)-1])
	if len(line) == 0 {
		return
	}

	// (1)
	names := bytes.SplitN(line, sepSubsection, 2)
	ok = reader.parseSection(names[0], num, false)
	if !ok {
		return
	}

	if debug >= debugL2 {
		log.Printf(">>> subsection names: %s", names)
	}

	// (2)
	bfirst := names[1][0]
	lastIdx := len(names[1]) - 1
	blast := names[1][lastIdx]
	if bfirst != tokDoubleQuote || blast != tokDoubleQuote {
		return
	}

	var (
		esc   bool
		runes = bytes.Runes(names[1][1:lastIdx])
	)

	if debug >= debugL2 {
		log.Printf(">>> subsection name: %s", string(runes))
	}

	reader.buf.Reset()

	for x := 0; x < len(runes); x++ {
		// (3)
		if esc {
			reader.buf.WriteRune(runes[x])
			esc = false
			continue
		}
		if runes[x] == tokBackslash {
			esc = true
			continue
		}
		if runes[x] == tokDoubleQuote {
			return
		}
		reader.buf.WriteRune(runes[x])
	}

	if esc {
		return
	}

	reader.sec.subName = nil
	reader.sec.subName = append(reader.sec.subName, reader.buf.Bytes()...)
	ok = true

	return
}
