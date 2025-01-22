// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package test

import (
	"testing"
)

func TestAssert(t *testing.T) {
	type testCase struct {
		a    any
		b    any
		exp  string
		desc string
	}

	var str = `a string`

	var cases = []testCase{{
		desc: `nil any`,
		a:    nil,
		b:    &str,
		exp:  `!!! Assert: IsValid: expecting <invalid Value>(false), got <*string Value>(true)`,
	}}

	var (
		c   testCase
		bw  BufferWriter
		got string
	)

	for _, c = range cases {
		Assert(&bw, ``, c.a, c.b)
		got = bw.String()
		if c.exp != got {
			t.Fatalf(`want: %s, got: %s`, c.exp, got)
		}

		bw.Reset()
	}
}
