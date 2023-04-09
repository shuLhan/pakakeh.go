// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewResponse(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		exp    *Response
		expErr string
	}{{
		desc: "With empty input",
	}, {
		desc:   "With invalid length",
		in:     "220\n",
		expErr: "invalid response length",
	}, {
		desc:   "With invalid code",
		in:     "22 -test\r\n",
		expErr: "invalid response code",
	}, {
		desc:   "With missing CRLF",
		in:     "220 test",
		expErr: "missing CRLF at message line",
	}, {
		desc:   "With trailing characters",
		in:     "220 test\r\n220",
		expErr: "trailing characters at message line",
	}, {
		desc: "Without message",
		in:   "220\r\n",
		exp: &Response{
			Code: 220,
		},
	}, {
		desc: "With message",
		in:   "220 test.local \r\n",
		exp: &Response{
			Code:    220,
			Message: "test.local",
		},
	}, {
		desc: "With multiline",
		in:   "220-test.local \r\n220-A\r\n220 B\r\n",
		exp: &Response{
			Code:    220,
			Message: "test.local",
			Body: []string{
				"A",
				"B",
			},
		},
	}, {
		desc:   "With inconsistent code on multiline",
		in:     "220-test.local \r\n210-A\r\n220 B\r\n",
		expErr: "inconsistent code",
	}, {
		desc:   `With invalid separator on multiline`,
		in:     "220-test.local \r\n210A\r\n220 B\r\n",
		expErr: `inconsistent code`,
	}, {
		desc:   "With missing CRLF on multiline",
		in:     "220-test.local \r\n220-A",
		expErr: "missing CRLF",
	}, {
		desc:   "With trailing characters on multiline",
		in:     "220-test.local \r\n220 A\r\n220",
		expErr: "trailing characters",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := NewResponse([]byte(c.in))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Response", c.exp, got)
	}
}
