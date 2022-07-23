// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"reflect"
	"testing"

	libreflect "github.com/shuLhan/share/lib/reflect"
	"github.com/shuLhan/share/lib/test"
)

func TestIsValueBoolTrue(t *testing.T) {
	cases := []struct {
		desc string
		v    string
		exp  bool
	}{{
		desc: "With empty value",
	}, {
		desc: "With value in all caps",
		v:    "TRUE",
		exp:  true,
	}, {
		desc: "With value is yes",
		v:    "YES",
		exp:  true,
	}, {
		desc: "With value is ya",
		v:    "yA",
		exp:  true,
	}, {
		desc: "With value is 1",
		v:    "1",
		exp:  true,
	}, {
		desc: "With value is 11",
		v:    "11",
		exp:  false,
	}, {
		desc: "With value is tru (typo)",
		v:    "tru",
		exp:  false,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := IsValueBoolTrue(c.v)

		test.Assert(t, "", c.exp, got)
	}
}

func TestParseTag(t *testing.T) {
	type testCase struct {
		in  string
		exp []string
	}

	var cases = []testCase{{
		in:  `sec`,
		exp: []string{`sec`, ``, ``, ``},
	}, {
		in:  `sec:sub`,
		exp: []string{`sec`, `sub`, ``, ``},
	}, {
		in:  `sec:sub:var`,
		exp: []string{`sec`, `sub`, `var`, ``},
	}, {
		in:  `sec:sub:var:def`,
		exp: []string{`sec`, `sub`, `var`, `def`},
	}, {
		in:  `sec:sub \"\:\\ name:var`,
		exp: []string{`sec`, `sub ":\ name`, `var`, ``},
	}}

	var (
		c   testCase
		got []string
	)
	for _, c = range cases {
		got = parseTag(c.in)
		test.Assert(t, c.in, c.exp, got)
	}
}

func TestParseTag_fromStruct(t *testing.T) {
	type ADT struct {
		F1 int `ini:"a"`
		F2 int `ini:"a:b"`
		F3 int `ini:"a:b:c"`
		F4 int `ini:"a:b:c:d"`
		F5 int `ini:"a:b \\\"\\: c:d"`
	}

	var (
		exp = [][]string{
			{`a`, ``, ``, ``},
			{`a`, `b`, ``, ``},
			{`a`, `b`, `c`, ``},
			{`a`, `b`, `c`, `d`},
			{`a`, `b ": c`, `d`, ``},
		}

		adt   ADT
		vtype reflect.Type
		field reflect.StructField
		tag   string
		got   []string
		x     int
	)

	vtype = reflect.TypeOf(adt)

	for x = 0; x < vtype.NumField(); x++ {
		field = vtype.Field(x)

		tag, _, _ = libreflect.Tag(field, "ini")

		got = parseTag(tag)
		test.Assert(t, tag, exp[x], got)
	}
}
