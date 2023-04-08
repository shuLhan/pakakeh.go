// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

// ParseResponseHeader parse HTTP response header and return it as standard
// HTTP Response with unreaded packet.
func ParseResponseHeader(raw []byte) (resp *http.Response, rest []byte, err error) {
	if len(raw) == 0 {
		return nil, nil, nil
	}
	// The minimum HTTP response without header is 16 bytes:
	// "HTTP/X.X" SP 3DIGITS CRLF CRLF
	if len(raw) < 16 {
		return nil, raw, fmt.Errorf("http: invalid response header length")
	}
	// The HTTP-name is case sensitive: "HTTP".
	if !bytes.Equal(raw[:4], []byte("HTTP")) {
		return nil, raw, fmt.Errorf("http: invalid protocol name '%s'", raw[:4])
	}
	if raw[4] != '/' {
		return nil, raw, fmt.Errorf("http: invalid protocol separator '%c'", raw[4])
	}
	if raw[6] != '.' {
		return nil, raw, fmt.Errorf("http: invalid version separator '%c'", raw[6])
	}
	ilf := bytes.Index(raw, []byte{'\n'})
	if ilf < 0 || raw[ilf-1] != '\r' {
		return nil, raw, fmt.Errorf("http: missing CRLF on status line")
	}

	resp = &http.Response{
		Proto:      string(raw[:8]),
		ProtoMajor: int(raw[5] - 48),
		ProtoMinor: int(raw[7] - 48),
		Status:     string(raw[9 : ilf-1]),
	}

	if resp.ProtoMajor <= 0 || resp.ProtoMajor > 2 {
		return nil, raw, fmt.Errorf("http: invalid major version '%c'", raw[5])
	}
	if resp.ProtoMinor < 0 || resp.ProtoMinor > 1 {
		return nil, raw, fmt.Errorf("http: invalid minor version '%c'", raw[7])
	}

	resp.StatusCode, err = strconv.Atoi(string(raw[9:12]))
	if err != nil {
		return nil, raw, fmt.Errorf("http: status code: " + err.Error())
	}
	if resp.StatusCode < 100 || resp.StatusCode >= 600 {
		return nil, raw, fmt.Errorf("http: invalid status code '%s'", raw[9:12])
	}

	rest = raw[ilf+1:]

	resp.Header, rest, err = parseHeaders(rest)
	if err != nil {
		return nil, raw, err
	}

	return resp, rest, nil
}

func parseHeaders(raw []byte) (header http.Header, rest []byte, err error) {
	var (
		parser = libbytes.NewParser(raw, []byte{':', '\n'})

		key string
		tok []byte
		c   byte
	)

	header = make(http.Header)
	rest = raw

	// Loop until we found an empty line with CRLF.
	for len(rest) > 0 {
		switch len(rest) {
		case 1:
			return nil, rest, fmt.Errorf(`http: missing CRLF at the end`)
		default:
			if rest[0] == '\r' && rest[1] == '\n' {
				rest = rest[2:]
				return header, rest, nil
			}
		}

		// Get the field name.
		tok, c = parser.Read()
		if c != ':' {
			return nil, nil, fmt.Errorf(`http: missing field value at line '%s'`, rest)
		}

		key = string(tok)

		tok, c = parser.Read()
		if c != '\n' {
			return nil, nil, fmt.Errorf(`http: missing CRLF at the end of field line`)
		}
		if tok[len(tok)-1] != '\r' {
			return nil, nil, fmt.Errorf(`http: missing CR at the end of line`)
		}

		tok = bytes.TrimSpace(tok)
		if len(tok) == 0 {
			return nil, nil, fmt.Errorf(`http: key '%s' have empty value`, key)
		}

		header.Add(key, string(tok))

		rest = parser.Remaining()
	}

	return header, rest, nil
}
