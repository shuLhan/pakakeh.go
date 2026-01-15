// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package http

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
)

// DefRangeLimit limit of content served by server when [Range] request
// without end, in example "0-".
const DefRangeLimit = 8388608

// Range define the unit and list of start-end positions for resource.
type Range struct {
	unit      string
	positions []*RangePosition
}

// NewRange create new Range with specified unit.
// The default unit is "bytes" if its empty.
func NewRange(unit string) (r *Range) {
	if len(unit) == 0 {
		unit = AcceptRangesBytes
	} else {
		unit = strings.ToLower(unit)
	}
	r = &Range{unit: unit}
	return r
}

// ParseMultipartRange parse "multipart/byteranges" response body.
// Each Content-Range position and body part in the multipart will be stored
// under [RangePosition].
func ParseMultipartRange(body io.Reader, boundary string) (r *Range, err error) {
	var (
		logp   = `ParseMultipartRange`
		reader = multipart.NewReader(body, boundary)
	)
	r = NewRange(``)
	for {
		var part *multipart.Part

		part, err = reader.NextPart()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf(`%s: on NextPart: %w`, logp, err)
		}

		var contentRange = part.Header.Get(HeaderContentRange)

		if len(contentRange) == 0 {
			continue
		}

		var pos *RangePosition
		pos, err = ParseContentRange(contentRange)
		if err != nil {
			continue
		}

		pos.content, err = io.ReadAll(part)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf(`%s: on ReadAll part: %w`, logp, err)
		}

		r.positions = append(r.positions, pos)
	}
	return r, nil
}

// ParseRange parses raw range value in the following format,
//
//	range    = unit "=" position *("," position)
//	unit     = 1*VCHAR
//	position = "-" last / start "-" / start "-" end
//	last     = 1*DIGIT
//	start    = 1*DIGIT
//	end      = 1*DIGIT
//
// An invalid position will be skipped.
func ParseRange(v string) (r Range) {
	var (
		par = libstrings.NewParser(v, `=`)

		tok   string
		delim rune
		err   error
	)

	for {
		tok, delim = par.ReadNoSpace()
		if delim == 0 {
			// No '=' found.
			return r
		}
		if len(tok) > 0 {
			break
		}
	}

	r.unit = strings.ToLower(tok)

	par.SetDelimiters(`-,`)
	for delim != 0 {
		tok, delim = par.ReadNoSpace()
		if len(tok) == 0 {
			if delim == 0 {
				break
			}
			if delim == ',' {
				// Empty range ", ..."
				continue
			}
			if delim == '-' {
				// Probably "-last".
				tok, delim = par.ReadNoSpace()
				if delim == '-' {
					// Invalid "-start-" or "-start-end".
					skipPosition(par, delim)
					continue
				}

				var end int64
				end, err = strconv.ParseInt(tok, 10, 64)
				if err != nil {
					continue
				}
				if end == 0 {
					// Invalid range "-0".
					continue
				}

				r.Add(nil, &end)
				continue
			}
		}
		if delim == ',' {
			// Invalid range "start,".
			continue
		}
		if delim == 0 {
			// Invalid range with "start" only.
			break
		}
		// delim is '-'.
		var start int64
		start, err = strconv.ParseInt(tok, 10, 64)
		if err != nil {
			skipPosition(par, delim)
			continue
		}

		tok, delim = par.ReadNoSpace()
		if delim == '-' {
			// Invalid range "start-end-"
			skipPosition(par, delim)
			continue
		}
		if len(tok) == 0 {
			// Range is "start-".
			r.Add(&start, nil)
		} else {
			// Range is "start-end".
			var end int64
			end, err = strconv.ParseInt(tok, 10, 64)
			if err != nil {
				skipPosition(par, delim)
				continue
			}
			r.Add(&start, &end)
		}
	}

	return r
}

// skipPosition Ignore any string until ','.
func skipPosition(par *libstrings.Parser, delim rune) {
	for delim == '-' {
		_, delim = par.Read()
	}
}

// Add start and end as requested position to Range.
// The start and end position is inclusive, closed interval [start, end],
// with end position must equal or greater than start position, unless its
// zero.
// For example,
//
//   - [0,+x] is valid, from offset 0 until x+1.
//   - [0,0] is valid and equal to first byte (but unusual).
//   - [+x,+y] is valid iff x <= y.
//   - [+x,-y] is invalid.
//   - [-x,+y] is invalid.
//
// The start or end can be nil, but not both.
// For example,
//
//   - [nil,+x] is valid, equal to "-x" or the last x bytes.
//   - [nil,0] is invalid.
//   - [nil,-x] is invalid.
//   - [x,nil] is valid, equal to "x-" or from offset x until end of file.
//   - [-x,nil] is invalid.
//
// The new position will be added and return true iff it does not overlap
// with existing list.
func (r *Range) Add(start, end *int64) bool {
	if start == nil && end == nil {
		return false
	}
	switch {
	case start == nil:
		if *end <= 0 {
			return false
		}
	case end == nil:
		if *start < 0 {
			return false
		}
	default:
		if *start < 0 || *end < 0 || *end < *start {
			return false
		}
	}

	var lastpos *RangePosition

	if len(r.positions) == 0 {
		goto ok
	}

	lastpos = r.positions[len(r.positions)-1]
	if lastpos.end == nil {
		return false
	}
	if lastpos.start == nil {
		if start == nil {
			// last=[nil,+b] vs. pos=[nil,+y]
			// The pos will always overlap with previous.
			return false
		}
		if end == nil {
			// last=[nil,+b] vs. pos=[+x,nil]
			// The pos will always overlap with previous.
			return false
		}
		goto ok
	}
	if start == nil {
		// [+a,+b] vs. [nil,+y]
		goto ok
	}
	if end == nil {
		// [+a,+b] vs. [+x,nil]
		if *lastpos.end >= *start {
			return false
		}
	}
	if *lastpos.end >= *start {
		return false
	}

ok:
	var pos = &RangePosition{}
	if start != nil {
		pos.start = new(int64)
		*pos.start = *start
	}
	if end != nil {
		pos.end = new(int64)
		*pos.end = *end
	}
	r.positions = append(r.positions, pos)
	return true
}

// IsEmpty return true if Range has no registered positions.
func (r *Range) IsEmpty() bool {
	return len(r.positions) == 0
}

// Positions return the list of range position.
func (r *Range) Positions() []*RangePosition {
	return r.positions
}

// String return the Range as value for HTTP header.
// It will return an empty string if no position registered.
func (r *Range) String() string {
	if r.IsEmpty() {
		return ``
	}

	var (
		sb  strings.Builder
		pos *RangePosition
		x   int
	)

	sb.WriteString(r.unit)

	sb.WriteByte('=')

	for x, pos = range r.positions {
		if x > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(pos.String())
	}
	return sb.String()
}
