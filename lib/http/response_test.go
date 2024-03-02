// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParseResponseHeader(t *testing.T) {
	cases := []struct {
		expResp *http.Response
		expErr  string
		desc    string
		raw     string
		expRest string
	}{{
		desc: `With empty input`,
	}, {
		desc:   `With invalid length`,
		raw:    "HTTP/1.1 101\r\n",
		expErr: `http: invalid response header length`,
	}, {
		desc:   `With lower case HTTP name`,
		raw:    "Http/1.1 101\r\n\r\n",
		expErr: `http: invalid protocol name 'Http'`,
	}, {
		desc:   `With invalid protocol separator`,
		raw:    "HTTP /1.1 101\r\n\r\n",
		expErr: `http: invalid protocol separator ' '`,
	}, {
		desc:   `With invalid version separator`,
		raw:    "HTTP/1 .1 101\r\n\r\n",
		expErr: `http: invalid version separator ' '`,
	}, {
		desc:   `With missing CR`,
		raw:    "HTTP/1.1 101\nHeader: Value\n",
		expErr: `http: missing CRLF on status line`,
	}, {
		desc:   `With invalid major version`,
		raw:    "HTTP/0.1 101\r\n\r\n",
		expErr: `http: invalid major version '0'`,
	}, {
		desc:   `With invalid major version (2)`,
		raw:    "HTTP/a.1 101\r\n\r\n",
		expErr: `http: invalid major version 'a'`,
	}, {
		desc:   `With invalid minor version`,
		raw:    "HTTP/1.  101\r\n\r\n",
		expErr: `http: invalid minor version ' '`,
	}, {
		desc:   `With invalid minor version (2)`,
		raw:    "HTTP/1.a 101\r\n\r\n",
		expErr: `http: invalid minor version 'a'`,
	}, {
		desc:   `With invalid status code #0`,
		raw:    "HTTP/1.1 999\r\n\r\n",
		expErr: `http: invalid status code '999'`,
	}, {
		desc:   `With invalid status code #1`,
		raw:    "HTTP/1.1 10a\r\n\r\n",
		expErr: `http: status code: strconv.Atoi: parsing "10a": invalid syntax`,
	}, {
		desc:   `Without CRLF #0`,
		raw:    "HTTP/1.1 101 Switching protocol\r\n\n",
		expErr: `http: missing CRLF at the end`,
	}, {
		desc:   `Without CRLF #1`,
		raw:    "HTTP/1.1 101 Switching protocol\r\nFi",
		expErr: `http: missing field value at line 'Fi'`,
	}, {
		desc: `With valid status line`,
		raw:  "HTTP/1.1 101 Switching protocol\r\n\r\n",
		expResp: &http.Response{
			Status:     `101 Switching protocol`,
			StatusCode: 101,
			Proto:      `HTTP/1.1`,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     http.Header{},
		},
	}, {
		desc: `With valid status line and a header`,
		raw:  "HTTP/1.1 101 Switching protocol\r\nKey: Value\r\n\r\n",
		expResp: &http.Response{
			Status:     `101 Switching protocol`,
			StatusCode: 101,
			Proto:      `HTTP/1.1`,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				`Key`: []string{
					`Value`,
				},
			},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, rest, err := ParseResponseHeader([]byte(c.raw)) //nolint: bodyclose
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, `http.Response`, c.expResp, got)
		test.Assert(t, `rest`, c.expRest, string(rest))
	}
}

func TestParseHeaders(t *testing.T) {
	cases := []struct {
		desc    string
		raw     string
		exp     http.Header
		expErr  string
		expRest string
	}{{
		desc: `With empty input`,
		exp:  make(http.Header),
	}, {
		desc:   `With single byte as input`,
		raw:    `x`,
		expErr: `http: missing CRLF at the end`,
	}, {
		desc:   `Without CRLF at the end`,
		raw:    `xx`,
		expErr: `http: missing field value at line 'xx'`,
	}, {
		desc:   `Without field separator`,
		raw:    "key value\r\n",
		expErr: "http: missing field value at line 'key value\r\n'",
	}, {
		desc:   `Without line feed`,
		raw:    `key:value`,
		expErr: `http: missing CRLF at the end of field line`,
	}, {
		desc:   `Without carriage return`,
		raw:    "key:value\n",
		expErr: `http: missing CR at the end of line`,
	}, {
		desc:   `Without field value`,
		raw:    "key:\r\n",
		expErr: `http: key 'key' have empty value`,
	}, {
		desc: `With valid field`,
		raw:  "key:value\r\n\r\n",
		exp: http.Header{
			`Key`: []string{
				`value`,
			},
		},
	}, {
		desc: `With valid field #1`,
		raw:  "key:value\r\nkey: another value     \r\n\r\nbody",
		exp: http.Header{
			`Key`: []string{
				`value`,
				`another value`,
			},
		},
		expRest: `body`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		header, rest, err := parseHeaders([]byte(c.raw))
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		test.Assert(t, `header`, c.exp, header)
		test.Assert(t, `rest`, c.expRest, string(rest))
	}
}
