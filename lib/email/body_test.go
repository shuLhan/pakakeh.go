// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseBody(t *testing.T) {
	cases := []struct {
		desc     string
		in       string
		boundary string
		expErr   string
		exp      string
		expRest  string
	}{{
		desc: "With empty input",
	}, {
		desc: "With only preamble",
		in: "This is a preamble\r\non multiple\nline\r\n" +
			"--with fake boundary\r\n" +
			"--and leading hyphens\r\n",
		boundary: "boundary",
	}, {
		desc: "Without boundary",
		in: "preamble\r\n\r\n" +
			"--boundary\r\n",
		exp: "\r\npreamble\r\n\r\n" +
			"--boundary\r\n",
	}, {
		desc: "With invalid header",
		in: "preamble\r\n\r\n" +
			"--boundary\r\n" +
			"Content-Encoding:\r\n\r\n",
		boundary: "boundary",
		expErr:   "ParseField: invalid character at index 19",
	}, {
		desc: "With epilogue",
		in: "preamble\r\n\r\n" +
			"--boundary--\r\n\r\n" +
			"trailing text\r\n",
		boundary: "boundary",
		expRest:  "\r\ntrailing text\r\n",
	}, {
		desc: "With boundary, epilogue",
		in: "preamble\r\n\r\n" +
			"--boundary\r\n\r\n" +
			"First body.\r\n" +
			"--boundary--\r\n\r\n" +
			"--Trailing\r\n",
		boundary: "boundary",
		exp:      "\r\nFirst body.\r\n",
		expRest:  "\r\n--Trailing\r\n",
	}, {
		desc: "With preamble, no header",
		in: "preamble\r\n\r\n" +
			"--boundary\r\n\r\n" +
			"First body.\r\n",
		boundary: "boundary",
		exp:      "\r\nFirst body.\r\n",
	}, {
		desc: "With preamble, and header",
		in: "\r\npreamble\r\n\r\n" +
			"--boundary\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"First body.\r\n",
		boundary: "boundary",
		exp:      "content-type:text/plain\r\n\r\nFirst body.\r\n",
	}, {
		desc: "With preamble, header, and rest",
		in: "\r\npreamble\r\n\r\n" +
			"--boundary\r\nContent-Type: text/plain\r\n\r\n" +
			"First body.\r\nWith\r\n a long long\n--lines and fake boundary\r\n" +
			"--boundary\r\nContent-Type: text/plain\r\n\r\n" +
			"Second body.\r\n",
		boundary: "boundary",
		exp: "content-type:text/plain\r\n\r\n" +
			"First body.\r\nWith\r\n a long long\n--lines and fake boundary\r\n" +
			"content-type:text/plain\r\n\r\n" +
			"Second body.\r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		body, rest, err := ParseBody([]byte(c.in), []byte(c.boundary))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}
		if body == nil {
			continue
		}

		test.Assert(t, "rest", c.expRest, string(rest), true)
		test.Assert(t, "body", c.exp, body.String(), true)
	}
}
