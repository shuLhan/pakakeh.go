// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
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
