// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"errors"
	"strconv"

	libbytes "github.com/shuLhan/share/lib/bytes"
	libio "github.com/shuLhan/share/lib/io"
)

//
// Response represent a generic single or multilines response from server.
//
type Response struct {
	Code    int
	Message string
	Body    []string
}

//
// NewResponse create and initialize new Response from parsing the raw
// response text.
//
func NewResponse(raw []byte) (res *Response, err error) {
	if len(raw) == 0 {
		return nil, nil
	}

	// The minimum length is 3 + CRLF
	if len(raw) < 5 {
		return nil, errors.New("invalid response length")
	}

	res = &Response{}

	code, isMultiline, err := res.parseCode(raw)
	if err != nil {
		return nil, err
	}

	reader := &libio.Reader{}
	seps := []byte{'-', ' '}
	terms := []byte{'\n'}

	reader.Init(string(raw[4:]))

	err = res.parseMessage(reader, isMultiline, terms)
	if err != nil {
		return nil, err
	}

	if !isMultiline {
		return res, nil
	}

	err = res.parseBody(reader, code, seps, terms)

	return res, err
}

//
// parseCode parse the first response code.
//
func (res *Response) parseCode(raw []byte) (code []byte, isMultiline bool, err error) {
	code = raw[0:3]

	for _, b := range code {
		if !libbytes.IsDigit(b) {
			return code, false, errors.New("invalid response code")
		}
	}
	if raw[3] == '-' {
		isMultiline = true
	}

	res.Code, err = strconv.Atoi(string(code))

	return code, isMultiline, err
}

//
// parseMessage parse the first line of response as response Message.
//
func (res *Response) parseMessage(
	reader *libio.Reader,
	isMultiline bool,
	terms []byte,
) (err error) {
	bb, _, c := reader.ReadUntil(nil, terms)
	if c == 0 {
		// It should be '\n'
		return errors.New("missing CRLF at message line")
	}

	res.Message = string(bytes.TrimSpace(bb))

	c = reader.SkipSpace()
	if !isMultiline && c != 0 {
		return errors.New("trailing characters at message line")
	}

	return nil
}

func (res *Response) parseBody(reader *libio.Reader, code, seps, terms []byte) (err error) {
	var (
		bb         []byte
		isLastLine bool
		c          byte
	)

	for {
		bb, _, c = reader.ReadUntil(seps, terms)
		switch c {
		case '-':
		case ' ':
			isLastLine = true
		default:
			return errors.New("invalid separator after code")
		}
		if !bytes.Equal(bb, code) {
			return errors.New("inconsistent code")
		}

		bb, _, c = reader.ReadUntil(nil, terms)
		if c == 0 {
			return errors.New("missing CRLF")
		}

		bb = bytes.TrimSpace(bb)
		if len(bb) > 0 {
			res.Body = append(res.Body, string(bb))
		}
		if isLastLine {
			c = reader.SkipSpace()
			if c != 0 {
				return errors.New("trailing characters")
			}
			break
		}
	}
	return nil
}
