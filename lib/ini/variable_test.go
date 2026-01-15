// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package ini

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestVariableString(t *testing.T) {
	type testCase struct {
		desc string
		v    *variable
		exp  string
	}

	var cases = []testCase{{
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
			mode:  lineModeKeyValue,
			key:   "name",
			value: "value",
		},
		exp: "name = value\n",
	}, {
		desc: `With mode value and comment`,
		v: &variable{
			mode:     lineModeKeyValue,
			key:      `name`,
			value:    `value`,
			rawValue: []byte(` value `),
			format:   "%s =%s; comment\n",
		},
		exp: "name = value ; comment\n",
	}, {
		desc: `With mode multi`,
		v: &variable{
			mode:  lineModeKeyValue,
			key:   `name`,
			value: `value`,
		},
		exp: "name = value\n",
	}, {
		desc: `With mode multi and comment`,
		v: &variable{
			mode:     lineModeKeyValue,
			key:      `name`,
			value:    `value`,
			rawValue: []byte(` value `),
			format:   "%s =%s; comment\n",
		},
		exp: "name = value ; comment\n",
	}}

	var (
		c   testCase
		got string
	)

	for _, c = range cases {
		t.Log(c.desc)

		got = c.v.String()

		test.Assert(t, "", c.exp, got)
	}
}
