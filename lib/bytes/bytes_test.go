// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bytes

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestTrimNull(t *testing.T) {
	type testCase struct {
		in  []byte
		exp []byte
	}

	var (
		cases = []testCase{{
			in: []byte{0},
		}, {
			in:  []byte{0, 'H'},
			exp: []byte{'H'},
		}, {
			in:  []byte{'H', 0},
			exp: []byte{'H'},
		}, {
			in:  []byte{'H'},
			exp: []byte{'H'},
		}}

		x   int
		c   testCase
		got []byte
	)

	for x, c = range cases {
		got = TrimNull(c.in)
		test.Assert(t, fmt.Sprintf(`TrimNull #%d`, x), c.exp, got)
	}
}
