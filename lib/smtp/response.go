// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package smtp

import (
	"bytes"
	"errors"
	"strconv"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

// Response represent a generic single or multilines response from server.
type Response struct {
	Message string
	Body    []string
	Code    int
}

// NewResponse create and initialize new Response from parsing the raw
// response text.
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

	var parser = libbytes.NewParser(raw[4:], []byte{'-', ' ', '\n'})

	err = res.parseMessage(parser, isMultiline)
	if err != nil {
		return nil, err
	}

	if !isMultiline {
		return res, nil
	}

	err = res.parseBody(parser, code)

	return res, err
}

// parseCode parse the first response code.
func (res *Response) parseCode(raw []byte) (code []byte, isMultiline bool, err error) {
	code = raw[0:3]

	for _, b := range code {
		if !ascii.IsDigit(b) {
			return code, false, errors.New("invalid response code")
		}
	}
	if raw[3] == '-' {
		isMultiline = true
	}

	res.Code, err = strconv.Atoi(string(code))

	return code, isMultiline, err
}

// parseMessage parse the first line of response as response Message.
func (res *Response) parseMessage(parser *libbytes.Parser, isMultiline bool) (err error) {
	var (
		tok []byte
		c   byte
	)

	parser.SetDelimiters([]byte{'\n'})

	tok, c = parser.Read()
	if c == 0 {
		// It should be '\n'
		return errors.New("missing CRLF at message line")
	}

	res.Message = string(bytes.TrimSpace(tok))

	_, c = parser.SkipSpaces()
	if !isMultiline && c != 0 {
		return errors.New("trailing characters at message line")
	}

	return nil
}

func (res *Response) parseBody(parser *libbytes.Parser, code []byte) (err error) {
	var (
		tok        []byte
		c          byte
		isLastLine bool
	)

	parser.SetDelimiters([]byte{'-', ' '})

	for {
		tok, c = parser.Read()
		if c == ' ' {
			isLastLine = true
		}
		if !bytes.Equal(tok, code) {
			return errors.New("inconsistent code")
		}

		tok, c = parser.ReadLine()
		if c == 0 {
			return errors.New("missing CRLF")
		}

		tok = bytes.TrimSpace(tok)
		if len(tok) > 0 {
			res.Body = append(res.Body, string(tok))
		}
		if isLastLine {
			_, c = parser.SkipSpaces()
			if c != 0 {
				return errors.New("trailing characters")
			}
			break
		}
	}
	return nil
}
