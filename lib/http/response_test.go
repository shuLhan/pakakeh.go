// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseResponseHeader(t *testing.T) {
	cases := []struct {
		desc    string
		raw     []byte
		expResp *http.Response
		expRest []byte
		expErr  string
	}{{
		desc: "With empty input",
	}, {
		desc:   "With invalid length",
		raw:    []byte("HTTP/1.1 101\r\n"),
		expErr: "http: invalid response header length",
	}, {
		desc:   "With lower case HTTP name",
		raw:    []byte("Http/1.1 101\r\n\r\n"),
		expErr: "http: invalid protocol name 'Http'",
	}, {
		desc:   "With invalid protocol separator",
		raw:    []byte("HTTP /1.1 101\r\n\r\n"),
		expErr: "http: invalid protocol separator ' '",
	}, {
		desc:   "With invalid version separator",
		raw:    []byte("HTTP/1 .1 101\r\n\r\n"),
		expErr: "http: invalid version separator ' '",
	}, {
		desc:   "With missing CR",
		raw:    []byte("HTTP/1.1 101\nHeader: Value\n"),
		expErr: "http: missing CRLF on status line",
	}, {
		desc:   "With invalid major version",
		raw:    []byte("HTTP/0.1 101\r\n\r\n"),
		expErr: "http: invalid major version '0'",
	}, {
		desc:   "With invalid major version (2)",
		raw:    []byte("HTTP/a.1 101\r\n\r\n"),
		expErr: "http: invalid major version 'a'",
	}, {
		desc:   "With invalid minor version",
		raw:    []byte("HTTP/1.  101\r\n\r\n"),
		expErr: "http: invalid minor version ' '",
	}, {
		desc:   "With invalid minor version (2)",
		raw:    []byte("HTTP/1.a 101\r\n\r\n"),
		expErr: "http: invalid minor version 'a'",
	}, {
		desc:   "With invalid status code",
		raw:    []byte("HTTP/1.1 999\r\n\r\n"),
		expErr: "http: invalid status code '999'",
	}, {
		desc:   "With invalid status code (2)",
		raw:    []byte("HTTP/1.1 10a\r\n\r\n"),
		expErr: `http: status code: strconv.Atoi: parsing "10a": invalid syntax`,
	}, {
		desc:   "Without CRLF",
		raw:    []byte("HTTP/1.1 101 Switching protocol\r\n\n"),
		expErr: "http: missing CRLF at the end, found \"\n\"",
	}, {
		desc:   "Without CRLF (2)",
		raw:    []byte("HTTP/1.1 101 Switching protocol\r\nFi"),
		expErr: "http: missing field separator at line \"Fi\"",
	}, {
		desc: "With valid status line",
		raw:  []byte("HTTP/1.1 101 Switching protocol\r\n\r\n"),
		expResp: &http.Response{
			Status:     "101 Switching protocol",
			StatusCode: 101,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     http.Header{},
		},
		expRest: []byte{},
	}, {
		desc: "With valid status line and a header",
		raw:  []byte("HTTP/1.1 101 Switching protocol\r\nKey: Value\r\n\r\n"),
		expResp: &http.Response{
			Status:     "101 Switching protocol",
			StatusCode: 101,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				"Key": []string{
					"Value",
				},
			},
		},
		expRest: []byte{},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, rest, err := ParseResponseHeader(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "http.Response", c.expResp, got, true)
		test.Assert(t, "rest", c.expRest, rest, true)
	}
}

func TestParseHeaders(t *testing.T) {
	cases := []struct {
		desc    string
		raw     []byte
		exp     http.Header
		expErr  string
		expRest []byte
	}{{
		desc: "With empty input",
		exp:  make(http.Header),
	}, {
		desc:   "With single byte as input",
		raw:    []byte{'x'},
		expErr: `http: missing CRLF at the end, found "x"`,
	}, {
		desc:   "Without CRLF at the end",
		raw:    []byte("xx"),
		expErr: `http: missing field separator at line "xx"`,
	}, {
		desc:   "Without field separator",
		raw:    []byte("key value\r\n"),
		expErr: "http: missing field separator at line \"key value\r\n\"",
	}, {
		desc:   "Without line feed",
		raw:    []byte("key:value"),
		expErr: "http: missing CRLF at the end of field line",
	}, {
		desc:   "Without carriage return",
		raw:    []byte("key:value\n"),
		expErr: "http: missing CR at the end of line",
	}, {
		desc:   "Without field value",
		raw:    []byte("key:\r\n"),
		expErr: "http: key 'key' have empty value",
	}, {
		desc: "With valid field",
		raw:  []byte("key:value\r\n\r\n"),
		exp: http.Header{
			"Key": []string{
				"value",
			},
		},
		expRest: []byte{},
	}, {
		desc: "With valid field (2)",
		raw:  []byte("key:value\r\nkey: another value     \r\n\r\nbody"),
		exp: http.Header{
			"Key": []string{
				"value",
				"another value",
			},
		},
		expRest: []byte("body"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		header, rest, err := parseHeaders(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "header", c.exp, header, true)
		test.Assert(t, "rest", c.expRest, rest, true)
	}
}
