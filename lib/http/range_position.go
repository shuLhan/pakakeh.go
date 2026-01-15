// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

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
func ParseContentRange(v string) (pos *RangePosition, err error) {
	var (
		logp = `ParseContentRange`
		p    = libstrings.NewParser(v, ` `)

		tok   string
		delim rune
	)

	pos = &RangePosition{}

	tok, delim = p.ReadNoSpace()
	if delim != ' ' {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}

	pos.unit = strings.ToLower(tok)
	if !(pos.unit == AcceptRangesBytes || pos.unit == AcceptRangesNone) {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}

	p.SetDelimiters(`-/`)

	tok, delim = p.ReadNoSpace()
	if len(tok) == 0 {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}
	if tok == `*` {
		if delim != '/' {
			return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
		}
		tok, delim = p.ReadNoSpace()
		if delim != 0 {
			return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
		}
		if tok == `*` {
			// "*/*": invalid range requested with unknown size.
			return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
		}

		pos = &RangePosition{}
		goto parselength
	}
	if delim != '-' {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}

	pos = &RangePosition{
		start:  new(int64),
		end:    new(int64),
		length: new(int64),
	}

	*pos.start, err = strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q: %w`, logp, v, err)
	}

	tok, delim = p.ReadNoSpace()
	if delim != '/' {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}
	*pos.end, err = strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q: %w`, logp, v, err)
	}
	if *pos.end < *pos.start {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}

	tok, delim = p.ReadNoSpace()
	if delim != 0 {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}
	if tok == `*` {
		// "x-y/*"
		return pos, nil
	}

parselength:
	*pos.length, err = strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q: %w`, logp, v, err)
	}
	if *pos.length < 0 {
		return nil, fmt.Errorf(`%s: invalid Content-Range %q`, logp, v)
	}
	return pos, nil
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
