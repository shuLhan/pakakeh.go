// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseBodyPart(t *testing.T) {
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
		desc:     "With only spaces",
		in:       "\r\n\r\n",
		boundary: "boundary",
	}, {
		desc:     "With invalid boundary",
		in:       "--invalid boundary\n",
		boundary: "boundary",
		expErr:   "ParseBodyPart: invalid boundary line: missing CR",
	}, {
		desc: "With preamble",
		in: "preamble\r\non multiple\nlines\r\n" +
			"--and leading hyphens\r\n",
		boundary: "boundary",
		expErr:   "ParseBodyPart: invalid boundary line: missing '--'",
	}, {
		desc:     "With missing boundary",
		in:       "text\r\n--boundary\r\n",
		boundary: "boundary",
		expErr:   "ParseBodyPart: missing boundary line",
	}, {
		desc:     "With mismatch boundary",
		in:       "--invalid boundary\r\n",
		boundary: "boundary",
		expErr:   "ParseBodyPart: boundary mismatch",
	}, {
		desc: "Without boundary",
		in: "preamble\r\n\r\n" +
			"--boundary\r\n",
		expErr: "ParseBodyPart: boundary parameter is empty",
	}, {
		desc: "With invalid header",
		in: "--boundary\r\n" +
			"Content-Encoding:\r\n\r\n",
		boundary: "boundary",
		expErr:   "email: empty field value at 'Content-Encoding:\r\n'",
	}, {
		desc: "With end of body",
		in: "--boundary--\r\n\r\n" +
			"trailing text\r\n",
		boundary: "boundary",
		expRest:  "trailing text\r\n",
	}, {
		desc: "With no header",
		in: "--boundary\r\n\r\n" +
			"First body.\r\n",
		boundary: "boundary",
		exp:      "\r\nFirst body.\r\n",
	}, {
		desc: "With header",
		in: "--boundary\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"First body.\r\n",
		boundary: "boundary",
		exp:      "content-type:text/plain\r\n\r\nFirst body.\r\n",
	}, {
		desc: "With header and second body",
		in: "--boundary\r\nContent-Type: text/plain\r\n\r\n" +
			"First body.\r\nWith\r\n a long long\n--lines and fake boundary\r\n" +
			"--boundary\r\nContent-Type: text/plain\r\n\r\n" +
			"Second body.\r\n",
		boundary: "boundary",
		exp: "content-type:text/plain\r\n\r\n" +
			"First body.\r\nWith\r\n a long long\n--lines and fake boundary\r\n",
		expRest: "--boundary\r\nContent-Type: text/plain\r\n\r\n" +
			"Second body.\r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, rest, err := ParseBodyPart([]byte(c.in), []byte(c.boundary))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}
		if got == nil {
			continue
		}

		test.Assert(t, "Rest", c.expRest, string(rest))
		test.Assert(t, "MIME", c.exp, got.String())
	}
}
