// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package email

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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
		expErr:   `ParseField: parseValue: empty field value`,
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
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}
		if body == nil {
			continue
		}

		test.Assert(t, "rest", c.expRest, string(rest))
		test.Assert(t, "body", c.exp, body.String())
	}
}

func TestBodyRelaxed(t *testing.T) {
	cases := []struct {
		desc string
		in   string
		exp  string
	}{{
		desc: "With empty body",
		in:   "",
		exp:  "",
	}, {
		desc: "With only CRL",
		in:   "\r\n",
		exp:  "\r\n",
	}, {
		desc: "With space at the end",
		in:   "\r\n\t \t \t",
		exp:  "\r\n",
	}, {
		desc: "With content and single CRLF",
		in:   "T\r\n",
		exp:  "T\r\n",
	}, {
		desc: "With content, space, and single CRLF",
		in:   "Th \t \r\nis\r\n",
		exp:  "Th \r\nis\r\n",
	}, {
		desc: "With text and multiple CRLF",
		in:   "Th\r\nis\r\n\r\n\r\n",
		exp:  "Th\r\nis\r\n",
	}, {
		desc: "With multiple spaces",
		in:   " C \r\nD \t E\r\nF \r G\r\nH \n I\r\n",
		exp:  " C\r\nD E\r\nF G\r\nH I\r\n",
	}}

	body := &Body{}

	for _, c := range cases {
		t.Log(c.desc)

		body.raw = []byte(c.in)
		got := body.Relaxed()

		test.Assert(t, "Relaxed", c.exp, string(got))
	}
}

func TestBodySimple(t *testing.T) {
	cases := []struct {
		desc string
		in   string
		exp  string
	}{{
		desc: "With empty body",
		in:   "",
		exp:  "\r\n",
	}, {
		desc: "With empty body and multiple CRLF",
		in:   "\r\n\r\n\r\n",
		exp:  "\r\n",
	}, {
		desc: "With content and single CRLF",
		in:   "T\r\n",
		exp:  "T\r\n",
	}, {
		desc: "With content and single CRLF",
		in:   "Th\r\nis\r\n",
		exp:  "Th\r\nis\r\n",
	}, {
		desc: "With text and multiple CRLF",
		in:   "Th\r\nis\r\n\r\n\r\n",
		exp:  "Th\r\nis\r\n",
	}}

	body := &Body{}

	for _, c := range cases {
		t.Log(c.desc)

		body.raw = []byte(c.in)
		got := body.Simple()

		test.Assert(t, "Simple", []byte(c.exp), got)
	}
}
