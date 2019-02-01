// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestHeaderUnpack(t *testing.T) {
	cases := []struct {
		desc    string
		raw     []byte
		expErr  string
		exp     string
		expRest []byte
	}{{
		desc: "With empty input",
	}, {
		desc:   "With whitespaces only",
		raw:    []byte(" \t"),
		expErr: "Header.Unpack: invalid end of header: ' \t'",
	}, {
		desc:    "With CRLF only",
		raw:     crlf,
		expRest: []byte{},
	}, {
		desc:   "With invalid end",
		raw:    []byte("a: 1\r\nx"),
		expErr: "Header.Unpack: invalid end of header: 'x'",
	}, {
		desc:   "With invalid field: missing value",
		raw:    []byte("a:\r\n\t"),
		expErr: "ParseField: invalid input",
	}, {
		desc: "With single field",
		raw:  []byte("a:1\r\n"),
		exp:  "a:1\r\n",
	}, {
		desc: "With multiple fields",
		raw:  []byte("a:1\r\nb : 2\r\n"),
		exp:  "a:1\r\nb:2\r\n",
	}, {
		desc:    "With empty line at the end",
		raw:     []byte("a:1\r\nb : 2\r\n\r\n"),
		exp:     "a:1\r\nb:2\r\n",
		expRest: []byte{},
	}, {
		desc:    "With body",
		raw:     []byte("a:1\r\nb : 2\r\n\r\nBody."),
		exp:     "a:1\r\nb:2\r\n",
		expRest: []byte("Body."),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		header := &Header{}

		rest, err := header.Unpack(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Header.String", c.exp, header.String(), true)
		test.Assert(t, "rest", c.expRest, rest, true)
	}
}
