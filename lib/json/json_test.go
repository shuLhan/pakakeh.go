// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestEscape(t *testing.T) {
	in := []byte("\"\\/\b\f\n\r\t")
	exp := []byte(`\"\\\/\b\f\n\r\t`)
	got := Escape(in)
	test.Assert(t, "Escape", exp, got)
}

func TestEscapeString(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in:  "",
		exp: "",
	}, {
		in: `	this\ is
		//\"☺"`,
		exp: `\tthis\\ is\n\t\t\/\/\\\"☺\"`,
	}, {
		in: ` `, exp: `\u0002\b\f\u000E\u000F\u0010\u0014\u001E\u001F `,
	}}

	var got string

	for _, c := range cases {
		t.Log(c)

		got = EscapeString(c.in)

		test.Assert(t, "", c.exp, got)
	}
}

func TestToMapStringFloat64(t *testing.T) {
	in := map[string]interface{}{
		"string": "1",
		"zero":   "0",
		"byte":   byte(3),
		"[]byte": []byte("4"),
	}

	exp := map[string]float64{
		"string": 1,
		"byte":   3,
		"[]byte": 4,
	}

	got, err := ToMapStringFloat64(in)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "ToMapStringFloat64", exp, got)
}

func TestUnescape(t *testing.T) {
	in := []byte(`\"\\\/\b\f\n\r\t`)
	exp := []byte("\"\\/\b\f\n\r\t")
	got, err := Unescape(in, false)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, "Unescape", exp, got)
}

func TestUnescapeString(t *testing.T) {
	cases := []struct {
		in     string
		strict bool
		exp    string
		expErr string
	}{{
		in:  "",
		exp: "",
	}, {
		in: `\tthis\\ is\n\t\t\/\/\\\"☺\"`,
		exp: `	this\ is
		//\"☺"`,
	}, {
		in: `\u0002\b\f\u000E\u000F\u0010\u0014\u001E\u001F\u263A `,
		exp: `☺ `}, {
		in:     `\uerror`,
		expErr: `strconv.ParseUint: parsing "erro": invalid syntax`,
	}, {
		in:  `\x`,
		exp: "x",
	}, {
		in:     `\x`,
		strict: true,
		expErr: `BytesJSONUnescape: invalid syntax at 1`,
	}}

	var (
		got string
		err error
	)

	for _, c := range cases {
		t.Log(c)

		got, err = UnescapeString(c.in, c.strict)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "value", c.exp, got)
	}
}
