// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"testing"
)

func TestAssert(t *testing.T) {
	type testCase struct {
		a    interface{}
		b    interface{}
		exp  string
		desc string
	}

	var str = `a string`

	var cases = []testCase{{
		desc: `nil interface{}`,
		a:    nil,
		b:    &str,
		exp:  `!!! Assert: IsValid: expecting <invalid Value>(false), got <*string Value>(true)`,
	}}

	var (
		c   testCase
		tw  TestWriter
		got string
	)

	for _, c = range cases {
		Assert(&tw, ``, c.a, c.b)
		got = tw.String()
		if c.exp != got {
			t.Fatalf(`want: %s, got: %s`, c.exp, got)
		}

		tw.Reset()
	}
}
