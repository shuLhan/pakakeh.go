// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestPartition(t *testing.T) {
	type testCase struct {
		exp Table
		ss  []string
		k   int
	}

	var cases = []testCase{{
		ss: []string{`a`, `b`},
		k:  1,
		exp: Table{
			{{`a`, `b`}},
		},
	}, {
		ss: []string{`a`, `b`},
		k:  2,
		exp: Table{
			{{`a`}, {`b`}},
		},
	}, {
		ss: []string{`a`, `b`, `c`},
		k:  1,
		exp: Table{
			{{`a`, `b`, `c`}},
		},
	}, {
		ss: []string{`a`, `b`, `c`},
		k:  2,
		exp: Table{
			{{`b`, `a`}, {`c`}},
			{{`b`}, {`c`, `a`}},
			{{`b`, `c`}, {`a`}},
		},
	}, {
		ss: []string{`a`, `b`, `c`},
		k:  3,
		exp: Table{
			{{`a`}, {`b`}, {`c`}},
		},
	}}

	var (
		c   testCase
		got Table
	)

	for _, c = range cases {
		t.Logf(`Partition: %d`, c.k)

		got = Partition(c.ss, c.k)

		test.Assert(t, ``, c.exp, got)
	}
}

func TestSinglePartition(t *testing.T) {
	type testCase struct {
		exp Table
		ss  []string
	}

	var cases = []testCase{{
		ss: []string{`a`, `b`, `c`},
		exp: Table{
			{{`a`}, {`b`}, {`c`}},
		},
	}}

	var (
		c   testCase
		got Table
	)
	for _, c = range cases {
		got = SinglePartition(c.ss)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestTable_IsEqual(t *testing.T) {
	type testCase struct {
		tcmp Table
		exp  bool
	}

	var table = Table{
		{{`a`}, {`b`, `c`}},
		{{`b`}, {`a`, `c`}},
		{{`c`}, {`a`, `b`}},
	}

	var cases = []testCase{{
		// Empty.
	}, {
		tcmp: table,
		exp:  true,
	}, {
		tcmp: Table{
			{{`c`}, {`a`, `b`}},
			{{`a`}, {`b`, `c`}},
			{{`b`}, {`a`, `c`}},
		},
		exp: true,
	}, {
		tcmp: Table{
			{{`c`}, {`a`, `b`}},
			{{`a`}, {`b`, `c`}},
		},
	}, {
		tcmp: Table{
			{{`b`}, {`a`, `b`}},
			{{`c`}, {`a`, `b`}},
			{{`a`}, {`b`, `c`}},
		},
	}}

	var (
		c   testCase
		got bool
	)

	for _, c = range cases {
		got = table.IsEqual(c.tcmp)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestTable_JoinCombination(t *testing.T) {
	type testCase struct {
		s     string
		table Table
		exp   Table
	}

	var cases = []testCase{{
		table: Table{
			{{`a`}, {`b`}, {`c`}},
		},
		s: `X`,
		exp: Table{
			{{`a`, `X`}, {`b`}, {`c`}},
			{{`a`}, {`b`, `X`}, {`c`}},
			{{`a`}, {`b`}, {`c`, `X`}},
		},
	}, {
		table: Table{
			{{`a`}, {`b`}, {`c`}},
			{{`g`}, {`h`}},
		},
		s: `X`,
		exp: Table{
			{{`a`, `X`}, {`b`}, {`c`}},
			{{`a`}, {`b`, `X`}, {`c`}},
			{{`a`}, {`b`}, {`c`, `X`}},
			{{`g`, `X`}, {`h`}},
			{{`g`}, {`h`, `X`}},
		},
	}}

	var (
		c   testCase
		got Table
	)
	for _, c = range cases {
		got = c.table.JoinCombination(c.s)
		test.Assert(t, ``, c.exp, got)
	}
}
