package http

import (
	"fmt"
	"strconv"
	"strings"

	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
)

// RangePosition contains the parsed value of Content-Range header.
type RangePosition struct {
	unit string

	start *int64
	end   *int64

	// length of resources.
	// A nil length, or "*", indicated an unknown size.
	length *int64

	content []byte
}

// ParseContentRange parse the HTTP header "Content-Range" value, as
// response from server, with the following format,
//
//	Content-Range = unit SP valid-range / invalid-range
//	           SP = " "
//	  valid-range = position "/" size
//	invalid-range = "*" "/" size
//	     position = start "-" end
//	         size = 1*DIGIT / "*"
//	        start = 1*DIGIT
//	          end = 1*DIGIT
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
	if len(tok) == 0 {
		return nil
	}
	if tok == `*` {
		if delim != '/' {
			return nil
		}
		tok, delim = p.ReadNoSpace()
		if delim != 0 {
			return nil
		}
		if tok == `*` {
			// "*/*": invalid range requested with unknown size.
			pos = &RangePosition{}
			return pos
		}

		pos = &RangePosition{}
		goto parselength
	}
	if delim != '-' {
		return nil
	}

	pos = &RangePosition{
		start:  new(int64),
		end:    new(int64),
		length: new(int64),
	}

	*pos.start, err = strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return nil
	}

	tok, delim = p.ReadNoSpace()
	if delim != '/' {
		return nil
	}
	*pos.end, err = strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return nil
	}
	if *pos.end < *pos.start {
		return nil
	}

	tok, delim = p.ReadNoSpace()
	if delim != 0 {
		return nil
	}
	if tok == `*` {
		// "x-y/*"
		return pos
	}

parselength:
	*pos.length, err = strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return nil
	}
	if *pos.length < 0 {
		return nil
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
	if pos.start == nil {
		if pos.end == nil {
			return `*`
		}
		return fmt.Sprintf(`-%d`, *pos.end)
	}
	if pos.end == nil {
		return fmt.Sprintf(`%d-`, *pos.start)
	}
	return fmt.Sprintf(`%d-%d`, *pos.start, *pos.end)
}
