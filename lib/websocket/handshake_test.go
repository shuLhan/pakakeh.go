// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestHandshakeParseHTTPLine(t *testing.T) {
	type testCase struct {
		desc   string
		req    string
		expErr string
		expURI string
	}

	var cases = []testCase{{
		desc:   "Without GET",
		req:    "POST / HTTP/1.1\r\n",
		expErr: ErrInvalidHTTPMethod.Error(),
	}, {
		desc:   "Without HTTP version",
		req:    "GET /\r\n",
		expErr: fmt.Sprintf(`%s: invalid request path`, ErrBadRequest),
	}, {
		desc:   "With invalid HTTP pragma",
		req:    "GET / HTTPS/1.1\r\n",
		expErr: fmt.Sprintf(`%s: invalid HTTP pragma`, ErrBadRequest),
	}, {
		desc:   "With invalid HTTP version",
		req:    "GET / HTTP/1.0\r\n",
		expErr: fmt.Sprintf(`%s`, ErrInvalidHTTPVersion),
	}, {
		desc:   "With invalid line",
		req:    "GET / HTTP/1.1 \r\n",
		expErr: fmt.Sprintf(`%s`, ErrInvalidHTTPVersion),
	}, {
		desc:   "With valid line",
		req:    "GET / HTTP/1.1\r\n",
		expURI: "/",
	}, {
		desc:   "With query",
		req:    "GET /?ticket=abcd HTTP/1.1\r\n",
		expURI: "/?ticket=abcd",
	}}

	var (
		h   = Handshake{}
		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log(c.desc)
		h.reset([]byte(c.req))

		err = h.parseHTTPLine()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "URI", c.expURI, h.URL.String())
	}
}
