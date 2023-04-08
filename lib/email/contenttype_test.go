// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseContentType(t *testing.T) {
	cases := []struct {
		in     string
		expErr string
		exp    string
	}{{
		exp: `text/plain; charset=us-ascii`,
	}, {
		in:     `text/;`,
		expErr: `ParseContentType: invalid subtype ''`,
	}, {
		in:  `text/plain`,
		exp: `text/plain;`,
	}, {
		in:     `text ;`,
		expErr: `ParseContentType: missing subtype`,
	}, {
		in:     `text /plain;`,
		expErr: `ParseContentType: invalid type 'text '`,
	}, {
		in:     `text/ plain;`,
		expErr: `ParseContentType: invalid subtype ' plain'`,
	}, {
		in:  `text/plain; key`,
		exp: `text/plain;`,
	}, {
		in:     `text/plain; ke(y)=value`,
		expErr: `ParseContentType: invalid parameter key 'ke(y)'`,
	}, {
		in:     `text/plain; key=value?`,
		expErr: `ParseContentType: invalid parameter value 'value?'`,
	}, {
		in:     `text/plain; key"value"`,
		expErr: `ParseContentType: expecting '=', got '"'`,
	}, {
		in:     `text/plain; key=val "value"`,
		expErr: `ParseContentType: invalid parameter value 'val'`,
	}, {
		in:     `text/plain; key="value?`,
		expErr: `ParseContentType: missing closing quote`,
	}, {
		in:  `text/plain; key="value ?"`,
		exp: `text/plain; key="value ?"`,
	}}

	for _, c := range cases {
		t.Log(c.in)

		got, err := ParseContentType([]byte(c.in))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "ContentType", c.exp, got.String())
	}
}

func TestGetParamValue(t *testing.T) {
	paramValue := []byte("multipart/mixed; boundary=\"----=_Part_1245\"\r\n")
	ct, err := ParseContentType(paramValue)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		in  []byte
		exp []byte
	}{{
		in: []byte("notexist"),
	}, {
		in:  []byte("Boundary"),
		exp: []byte("----=_Part_1245"),
	}}

	for _, c := range cases {
		t.Log(c.in)

		got := ct.GetParamValue(c.in)

		test.Assert(t, "GetParamValue", c.exp, got)
	}
}
