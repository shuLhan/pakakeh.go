// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"strings"

	libio "github.com/shuLhan/share/lib/io"
)

//
// Body represent single or multiple message body parts.
//
type Body struct {
	// Parts contains one or more message body.
	Parts []*MIME
	// raw contains original message body.
	raw []byte
}

//
// ParseBody parse the the raw message's body using boundary.
//
func ParseBody(raw, boundary []byte) (body *Body, rest []byte, err error) {
	if len(raw) == 0 {
		return nil, nil, nil
	}

	body = &Body{
		raw: raw,
	}

	if len(boundary) == 0 {
		body.Parts = append(body.Parts, &MIME{
			Content: raw,
		})
		return body, nil, nil
	}

	var (
		mime *MIME
		// Minimum length of end boundary line is length of boundary
		// plus four hyphens "-" plus CRLF.
		minlen = len(boundary) + 6
	)

	rest = skipPreamble(raw, boundary)
	for {
		mime, rest, err = ParseBodyPart(rest, boundary)
		if err != nil {
			return nil, rest, err
		}
		if mime == nil {
			break
		}

		body.Parts = append(body.Parts, mime)

		if len(rest) < minlen {
			break
		}
	}

	return body, rest, nil
}

func skipPreamble(raw, boundary []byte) []byte {
	r := &libio.Reader{}
	r.Init(raw)

	for {
		line := r.ReadLine()
		if len(line) == 0 {
			return r.Rest()
		}
		if len(line) < len(boundary)+4 {
			continue
		}
		if line[len(line)-2] != cr {
			continue
		}
		if !bytes.Equal(line[:2], boundSeps) {
			continue
		}
		if !bytes.Equal(line[2:2+len(boundary)], boundary) {
			continue
		}
		r.UnreadN(len(line))
		break
	}
	return r.Rest()
}

//
// String return text representation of Body.
//
func (body *Body) String() string {
	var sb strings.Builder

	for _, part := range body.Parts {
		sb.WriteString(part.String())
	}

	return sb.String()
}

//
// Relaxed canonicalize the original body with "relaxed" algorithm as defined
// in RFC 6376 section 3.4.4.
// It remove all trailing whitespaces, reduce sequence of whitespaces inside
// line into single space, and remove all empty line at the end of body.
// If body is not empty and not end with CRLF, a CRLF is added.
//
// This function is expensive for message with large body, its better if we
// call it once and store it somewhere.
//
func (body *Body) Relaxed() (out []byte) { //nolint: gocognit
	if len(body.raw) == 0 {
		return
	}

	out = make([]byte, 0, len(body.raw))
	x := len(body.raw) - 1

	// Remove trailing whitespaces.
	for ; x >= 0; x-- {
		if body.raw[x] == '\t' || body.raw[x] == ' ' {
			continue
		}
		break
	}

	// Remove empty lines ...
	hasCRLF := false
	for x > 2 {
		if body.raw[x-1] == cr && body.raw[x] == lf {
			hasCRLF = true
			x -= 2
			continue
		}
		break
	}

	// Reduce sequence of WSP.
	end := x
	hasSpace := 0
	for x = 0; x <= end; x++ {
		if body.raw[x] == '\t' || body.raw[x] == ' ' || body.raw[x] == '\n' {
			hasSpace++
			continue
		}
		if body.raw[x] == '\r' {
			x++
			if body.raw[x] == '\n' {
				if hasSpace > 1 {
					out = append(out, ' ')
				}
				out = append(out, cr)
				out = append(out, lf)
				hasSpace = 0
				continue
			}
			hasSpace++
			continue
		}
		if hasSpace > 0 {
			out = append(out, ' ')
			hasSpace = 0
		}
		out = append(out, body.raw[x])
	}
	if len(out) >= 2 {
		if out[len(out)-2] == cr && out[len(out)-1] == lf {
			return out
		}
	}
	if hasCRLF {
		out = append(out, "\r\n"...)
	}

	return out
}

//
// Simple canonicalize the original body with "simple" algorithm as defined in
// RFC 6376 section 3.4.3.
// Basically, it converts "*CRLF" at the end of body to a single CRLF.
// If no message body or no trailing CRLF, a CRLF is added.
//
func (body *Body) Simple() (out []byte) {
	if len(body.raw) == 0 {
		return []byte{cr, lf}
	}

	out = make([]byte, len(body.raw))
	copy(out, body.raw)

	x := len(out) - 1
	for x > 2 {
		if out[x-1] == cr && out[x] == lf {
			out = out[:len(out)-2]
			x -= 2
			continue
		}
		break
	}
	switch x {
	case 1:
	default:
		if out[x-1] != cr && out[x] != lf {
			out = append(out, "\r\n"...)
		}
	}

	return out
}
