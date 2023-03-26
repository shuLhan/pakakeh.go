package http

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/parser"
)

// Range define the unit and list of start-end positions for resource.
type Range struct {
	unit      string
	positions []RangePosition
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

// ParseMultipartRange parse multipart/byteranges response body.
// Each Content-Range position and body part in the multipart will be stored
// under RangePosition.
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

		var pos = ParseContentRange(contentRange)
		if pos == nil {
			continue
		}

		pos.content, err = io.ReadAll(part)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf(`%s: on ReadAll part: %s`, logp, err)
		}

		r.positions = append(r.positions, *pos)
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
		par = parser.New(v, `=`)

		tok   string
		delim rune
		err   error
	)

	for {
		tok, delim = par.TokenTrimSpace()
		if delim == 0 {
			// No '=' found.
			return r
		}
		if len(tok) > 0 {
			break
		}
	}

	r.unit = strings.ToLower(tok)

	var (
		start, end int64
	)
	par.SetDelimiters(`-,`)
	for delim != 0 {
		tok, delim = par.TokenTrimSpace()
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
				tok, delim = par.TokenTrimSpace()
				if delim != 0 && delim != ',' {
					// Invalid "-start-" or "-start-end".
					skipPosition(par, delim)
					continue
				}

				start, err = strconv.ParseInt(tok, 10, 64)
				if err != nil {
					skipPosition(par, delim)
					continue
				}

				r.Add(-1*start, 0)
				skipPosition(par, delim)
				continue
			}
		}
		if delim == ',' || delim == 0 {
			// Invalid range "start,..." or "start$".
			continue
		}

		// delim == '-'
		start, err = strconv.ParseInt(tok, 10, 64)
		if err != nil {
			skipPosition(par, delim)
			continue
		}

		tok, delim = par.TokenTrimSpace()
		if delim == '-' {
			// Invalid range "start-end-"
			skipPosition(par, delim)
			continue
		}
		if len(tok) == 0 {
			if start == 0 {
				// Invalid range, "0-" equal to whole body.
				continue
			}

			// Range "start-".
			end = 0
		} else {
			// Range "start-end".
			end, err = strconv.ParseInt(tok, 10, 64)
			if err != nil {
				skipPosition(par, delim)
				continue
			}
		}
		r.Add(start, end)
	}

	return r
}

// skipPosition Ignore any string until ','.
func skipPosition(par *parser.Parser, delim rune) {
	for delim != ',' && delim != 0 {
		_, delim = par.Token()
	}
}

// Add start and end as requested position to Range.
// The start and end position is inclusive, closed interval [start, end],
// with end position must equal or greater than start position, unless its
// zero.
// For example,
//
//   - [0,0] is valid and equal to first byte (but unusual)
//   - [0,9] is valid and equal to the first 10 bytes.
//   - [10,0] is valid and equal to the bytes from offset 10 until the end.
//   - [-10,0] is valid and equal to the last 10 bytes.
//   - [10,1] or [0,-10] or [-10,10] is not valid position.
//
// The new position will be added and return true if only if it does not
// overlap with existing list.
func (r *Range) Add(start, end int64) bool {
	if end != 0 && end < start {
		// [10,1] or [0,-10]
		return false
	}
	if start < 0 && end != 0 {
		// [-10,10]
		return false
	}

	var pos RangePosition
	for _, pos = range r.positions {
		if pos.Start < 0 {
			if start < 0 {
				// Given pos:[-10,0], adding another negative
				// start like -20 or -5 will always cause
				// overlap.
				return false
			}
		} else if pos.Start == 0 {
			if start >= 0 && start <= pos.End {
				// pos:[0,+y], start<y.
				return false
			}
		} else {
			if pos.End == 0 {
				// pos:[+x,0] already accept until the end.
				return false
			}
			if start >= 0 && start <= pos.End {
				// pos:[+x,+y], start<y.
				return false
			}
		}
	}

	pos = RangePosition{
		Start: start,
		End:   end,
	}
	if start < 0 {
		pos.Length = start * -1
	} else if start >= 0 && end >= 0 {
		pos.Length = (end - start) + 1
	}
	r.positions = append(r.positions, pos)

	return true
}

// IsEmpty return true if Range has no registered positions.
func (r *Range) IsEmpty() bool {
	return len(r.positions) == 0
}

// Positions return the list of range position.
func (r *Range) Positions() []RangePosition {
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
		pos RangePosition
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
