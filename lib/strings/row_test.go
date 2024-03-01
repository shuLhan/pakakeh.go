// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestRowIsEqual(t *testing.T) {
	type testCase struct {
		a   Row
		b   Row
		exp bool
	}

	var cases = []testCase{{
		a:   Row{{`a`}, {`b`, `c`}},
		b:   Row{{`a`}, {`b`, `c`}},
		exp: true,
	}, {
		a:   Row{{`a`}, {`b`, `c`}},
		b:   Row{{`a`}, {`c`, `b`}},
		exp: true,
	}, {
		a:   Row{{`a`}, {`b`, `c`}},
		b:   Row{{`c`, `b`}, {`a`}},
		exp: true,
	}, {
		a: Row{{`a`}, {`b`, `c`}},
		b: Row{{`a`}, {`b`, `a`}},
	}}

	var (
		c   testCase
		got bool
	)
	for _, c = range cases {
		got = c.a.IsEqual(c.b)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestRowJoin(t *testing.T) {
	type testCase struct {
		lsep string
		ssep string
		exp  string
		row  Row
	}
	var cases = []testCase{{
		// Empty input.
	}, {
		lsep: `;`,
		ssep: `,`,
		exp:  ``,
	}, {
		row:  Row{{`a`}, {}},
		lsep: `;`,
		ssep: `,`,
		exp:  `a;`,
	}, {
		row:  Row{{`a`}, {`b`, `c`}},
		lsep: `;`,
		ssep: `,`,
		exp:  `a;b,c`,
	}}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		got = c.row.Join(c.lsep, c.ssep)
		test.Assert(t, ``, c.exp, got)
	}
}
