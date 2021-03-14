// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestHeaderBoundary(t *testing.T) {
	cases := []struct {
		desc string
		in   string
		exp  []byte
	}{{
		desc: "With no content-type",
		in: "From: Nathaniel Borenstein <nsb@bellcore.com>\r\n" +
			"To: Ned Freed <ned@innosoft.com>\r\n" +
			"Date: Sun, 21 Mar 1993 23:56:48 -0800 (PST)\r\n" +
			"Subject: Sample message\r\n" +
			"\r\n",
	}, {
		desc: "With invalid content-type",
		in: "From: Nathaniel Borenstein <nsb@bellcore.com>\r\n" +
			"To: Ned Freed <ned@innosoft.com>\r\n" +
			"Date: Sun, 21 Mar 1993 23:56:48 -0800 (PST)\r\n" +
			"Subject: Sample message\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-type: multipart/mixed; boundary=simple:boundary\r\n" +
			"\r\n",
	}, {
		desc: "With boundary",
		in: "From: Nathaniel Borenstein <nsb@bellcore.com>\r\n" +
			"To: Ned Freed <ned@innosoft.com>\r\n" +
			"Date: Sun, 21 Mar 1993 23:56:48 -0800 (PST)\r\n" +
			"Subject: Sample message\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-type: multipart/mixed; boundary=\"simple boundary\"\r\n" +
			"\r\n",
		exp: []byte("simple boundary"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		header, _, err := ParseHeader([]byte(c.in))
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Boundary", c.exp, header.Boundary())
	}
}

func TestParseHeader(t *testing.T) {
	cases := []struct {
		desc       string
		raw        []byte
		expErr     string
		exp        string
		expRelaxed string
		expSimple  string
		expRest    []byte
	}{{
		desc: "With empty input",
	}, {
		desc:   "With whitespaces only",
		raw:    []byte(" \t"),
		expErr: "email: invalid field at ' \t'",
	}, {
		desc:    "With CRLF only",
		raw:     []byte("\r\n"),
		expRest: []byte{},
	}, {
		desc:   "With invalid end",
		raw:    []byte("a: 1\r\nx"),
		expErr: "ParseHeader: invalid end of header: 'x'",
	}, {
		desc:   "With invalid field: missing value",
		raw:    []byte("a:\r\n\t"),
		expErr: "email: empty field value at 'a:\r\n\t'",
	}, {
		desc:       "With single field",
		raw:        []byte("a:1\r\n"),
		exp:        "a:1\r\n",
		expRelaxed: "a:1\r\n",
		expSimple:  "a:1\r\n",
	}, {
		desc:       "With multiple fields",
		raw:        []byte("a:1\r\nb : 2\r\n"),
		exp:        "a:1\r\nb:2\r\n",
		expRelaxed: "a:1\r\nb:2\r\n",
		expSimple:  "a:1\r\nb : 2\r\n",
	}, {
		desc:       "With empty line at the end",
		raw:        []byte("a:1\r\nb : 2\r\n\r\n"),
		exp:        "a:1\r\nb:2\r\n",
		expRest:    []byte{},
		expRelaxed: "a:1\r\nb:2\r\n",
		expSimple:  "a:1\r\nb : 2\r\n",
	}, {
		desc:       "With body",
		raw:        []byte("a:1\r\nb : 2\r\n\r\nBody."),
		exp:        "a:1\r\nb:2\r\n",
		expRest:    []byte("Body."),
		expRelaxed: "a:1\r\nb:2\r\n",
		expSimple:  "a:1\r\nb : 2\r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		header, rest, err := ParseHeader(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}
		if header == nil {
			continue
		}

		test.Assert(t, "Header.Relaxed", []byte(c.exp), header.Relaxed())
		test.Assert(t, "rest", c.expRest, rest)

		test.Assert(t, "Header.Relaxed", []byte(c.expRelaxed), header.Relaxed())
		test.Assert(t, "Header.Simple", []byte(c.expSimple), header.Simple())
	}
}
