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
	//
	// Parts contains one or more message body.
	//
	Parts []*MIME // nolint: structcheck,unused
}

//
// Unpack the message's body using boundary.
//
func (body *Body) Unpack(raw, boundary []byte) (rest []byte, err error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if len(boundary) == 0 {
		part := &MIME{
			Content: raw,
		}
		body.Parts = append(body.Parts, part)
		return nil, nil
	}

	var (
		mime *MIME
		// Minimum length of end boundary line is length of boundary
		// plus four hyphens "-" plus CRLF.
		minlen = len(boundary) + 6
	)

	rest = body.skipPreamble(raw, boundary)
	for {
		mime, rest, err = ParseBodyPart(rest, boundary)
		if err != nil {
			return rest, err
		}
		if mime == nil {
			break
		}

		body.Parts = append(body.Parts, mime)

		if len(rest) < minlen {
			break
		}
	}

	return rest, nil
}

func (body *Body) skipPreamble(raw, boundary []byte) []byte {
	r := &libio.Reader{}
	r.InitBytes(raw)

	for {
		line := r.ReadLine()
		if len(line) == 0 {
			return r.Rest()
		}
		if len(line) < len(boundary)+4 {
			continue
		}
		if line[len(line)-2] != '\r' {
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
