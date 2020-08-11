// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestHandshakeParseHTTPLine(t *testing.T) {
	cases := []struct {
		desc   string
		req    string
		expErr error
		expURI string
	}{{
		desc:   "Without GET",
		req:    "POST / HTTP/1.1\r\n",
		expErr: ErrInvalidHTTPMethod,
	}, {
		desc:   "Without HTTP version",
		req:    "GET /\r\n",
		expErr: ErrBadRequest,
	}, {
		desc:   "With invalid HTTP in version",
		req:    "GET / HTTPS/1.1\r\n",
		expErr: ErrBadRequest,
	}, {
		desc:   "With invalid HTTP version",
		req:    "GET / HTTP/1.0\r\n",
		expErr: ErrInvalidHTTPVersion,
	}, {
		desc:   "With invalid line",
		req:    "GET / HTTP/1.1 \r\n",
		expErr: ErrInvalidHTTPVersion,
	}, {
		desc:   "With valid line",
		req:    "GET / HTTP/1.1\r\n",
		expURI: "/",
	}, {
		desc:   "With query",
		req:    "GET /?ticket=abcd HTTP/1.1\r\n",
		expURI: "/?ticket=abcd",
	}}

	h := Handshake{}

	for _, c := range cases {
		t.Log(c.desc)
		h.reset([]byte(c.req))

		err := h.parseHTTPLine()
		if err != nil {
			test.Assert(t, "err", c.expErr, err, true)
			continue
		}

		test.Assert(t, "URI", c.expURI, h.URL.String(), true)
	}
}
