package http

import (
	"fmt"
	"strconv"
	"strings"

	libstrings "github.com/shuLhan/share/lib/strings"
)

// RangePosition contains the parsed value of Content-Range header.
type RangePosition struct {
	unit string

	content []byte

	Start int64
	End   int64

	// Length of zero means read until the end.
	Length int64
}

// ParseContentRange parse Content-Range value, the following format,
//
//	unit SP position "/" size
//	SP       = " "
//	position = start "-" end / start "-" / "-" last
//	start, end, last, size = 1*DIGIT
//
// It will return nil if the v is invalid.
func ParseContentRange(v string) (pos *RangePosition) {
	var (
		p = libstrings.NewParser(v, ` `)

		tok   string
		delim rune
		err   error
	)

	pos = &RangePosition{}

	tok, delim = p.ReadNoSpace()
	if delim != ' ' {
		return nil
	}

	pos.unit = strings.ToLower(tok)

	p.SetDelimiters(`-/`)

	tok, delim = p.ReadNoSpace()
	if len(tok) == 0 && delim == '-' {
		// Probably "-last".
		tok, delim = p.ReadNoSpace()
		if delim != '/' {
			return nil
		}

		pos.Length, err = strconv.ParseInt(tok, 10, 64)
		if err != nil {
			return nil
		}

		pos.Start = -1 * pos.Length
	} else {
		if delim != '-' || delim == 0 {
			return nil
		}

		pos.Start, err = strconv.ParseInt(tok, 10, 64)
		if err != nil {
			return nil
		}

		tok, delim = p.ReadNoSpace()
		if delim != '/' {
			return nil
		}

		if len(tok) != 0 {
			// Case of "start-end/size".
			pos.End, err = strconv.ParseInt(tok, 10, 64)
			if err != nil {
				return nil
			}
			pos.Length = (pos.End - pos.Start) + 1
		}
	}

	// The size.
	tok, delim = p.ReadNoSpace()
	if delim != 0 {
		return nil
	}

	if tok != "*" {
		var size int64
		size, err = strconv.ParseInt(tok, 10, 64)
		if err != nil {
			return nil
		}
		if pos.End == 0 {
			// Case of "start-/size".
			pos.Length = (size - pos.Start)
		}
	}

	return pos
}

// Content return the range content body in multipart.
func (pos RangePosition) Content() []byte {
	return pos.content
}

// ContentRange return the string that can be used for HTTP Content-Range
// header value.
func (pos RangePosition) ContentRange(unit string, size int64) (v string) {
	if size == 0 {
		v = fmt.Sprintf(`%s %s/*`, unit, pos.String())
	} else {
		v = fmt.Sprintf(`%s %s/%d`, unit, pos.String(), size)
	}
	return v
}

func (pos RangePosition) String() string {
	if pos.Start < 0 {
		return fmt.Sprintf(`%d`, pos.Start)
	}
	if pos.Start > 0 && pos.End == 0 {
		return fmt.Sprintf(`%d-`, pos.Start)
	}
	return fmt.Sprintf(`%d-%d`, pos.Start, pos.End)
}
