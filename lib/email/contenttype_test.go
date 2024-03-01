// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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
		in:     `text ;`,
		expErr: `ParseContentType: missing subtype`,
	}, {
		in:     `text /plain;`,
		expErr: `ParseContentType: invalid type 'text '`,
	}, {
		in:     `text/ plain;`,
		expErr: `ParseContentType: invalid subtype ' plain'`,
	}, {
		in:     `text/plain/;`,
		expErr: `ParseContentType: invalid character '/'`,
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
		in:  `text/plain`,
		exp: `text/plain`,
	}, {
		in:  `text/plain;`,
		exp: `text/plain`,
	}, {
		in:  `text/plain; key`,
		exp: `text/plain`,
	}, {
		in:  `text/plain; key=val;`,
		exp: `text/plain; key=val`,
	}, {
		in:  `text/plain; key="value ?"`,
		exp: `text/plain; key="value ?"`,
	}, {
		in:  `text/plain; key="value ?"; key2="b=c;d"; key3=";e=f"`,
		exp: `text/plain; key="value ?"; key2="b=c;d"; key3=";e=f"`,
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

func TestContentType_isEqual(t *testing.T) {
	var (
		ct = &ContentType{
			Top: `text`,
			Sub: `plain`,
		}
		got bool
	)

	// Case: with nil.
	got = ct.isEqual(nil)
	if got != false {
		t.Fatalf(`want false, got %v`, got)
	}

	// Case: with Top not match.
	got = ct.isEqual(&ContentType{Top: `TEX`})
	if got != false {
		t.Fatalf(`want false, got %v`, got)
	}

	// Case: with Sub not match.
	got = ct.isEqual(&ContentType{Top: `TEXT`, Sub: `PLAI`})
	if got != false {
		t.Fatalf(`want false, got %v`, got)
	}

	got = ct.isEqual(&ContentType{Top: `TEXT`, Sub: `PLAIN`})
	if got != true {
		t.Fatalf(`want true, got %v`, got)
	}
}
