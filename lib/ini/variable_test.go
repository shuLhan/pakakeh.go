// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestVariableEscape(t *testing.T) {
	cases := []struct {
		desc string
		in   string
		exp  string
	}{{
		desc: "With empty input",
		in:   "",
		exp:  `""`,
	}, {
		desc: "With escaped characters",
		in:   "x\b\n\t\\\"x",
		exp:  `"x\b\n\t\\\"x"`,
	}}

	for _, c := range cases {
		t.Log(c.desc)
		got := escape(c.in)
		test.Assert(t, "escape", c.exp, got)
	}
}

func TestVariableString(t *testing.T) {
	cases := []struct {
		desc string
		v    *variable
		exp  string
	}{{
		desc: "With mode empty #1",
		v: &variable{
			mode: lineModeEmpty,
		},
	}, {
		desc: "With mode empty #2",
		v: &variable{
			mode:   lineModeEmpty,
			format: "	",
		},
		exp: "	",
	}, {
		desc: "With line mode comment",
		v: &variable{
			mode:   lineModeComment,
			format: `  ; comment`,
		},
		exp: `  ; comment`,
	}, {
		desc: "With mode value",
		v: &variable{
			mode:  lineModeValue,
			key:   "name",
			value: "value",
		},
		exp: "name = value\n",
	}, {
		desc: `With mode value and comment`,
		v: &variable{
			mode:   lineModeValue,
			key:    `name`,
			value:  `value`,
			format: "%s = %s ; comment\n",
		},
		exp: "name = value ; comment\n",
	}, {
		desc: "With mode multi",
		v: &variable{
			mode:  lineModeMulti,
			key:   "name",
			value: "value",
		},
		exp: "name = value\n",
	}, {
		desc: "With mode multi and comment",
		v: &variable{
			mode:   lineModeMulti,
			key:    `name`,
			value:  `value`,
			format: "%s = %s ; comment\n",
		},
		exp: "name = value ; comment\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.v.String()

		test.Assert(t, "", c.exp, got)
	}
}
