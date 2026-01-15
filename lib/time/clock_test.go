// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package time

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestCreateClock(t *testing.T) {
	type testCase struct {
		desc           string
		hour, min, sec int
		exp            Clock
	}

	var cases = []testCase{{
		desc: `minus values`,
		hour: -1,
		min:  -1,
		sec:  -1,
		exp:  Clock{},
	}, {
		desc: `overflow values`,
		hour: 24,
		min:  60,
		sec:  60,
		exp:  Clock{},
	}, {
		desc: `valid values`,
		hour: 0,
		min:  1,
		sec:  2,
		exp:  Clock{min: 1, sec: 2},
	}}

	var (
		c   testCase
		got Clock
	)
	for _, c = range cases {
		got = CreateClock(c.hour, c.min, c.sec)

		test.Assert(t, c.desc, c.exp, got)
	}
}

func TestParseClock(t *testing.T) {
	type testCase struct {
		v   string
		exp Clock
	}

	var cases = []testCase{{
		v:   ``,
		exp: Clock{},
	}, {
		v:   `::`,
		exp: Clock{},
	}, {
		v:   `:03`,
		exp: Clock{min: 3},
	}, {
		v:   `012`,
		exp: Clock{hour: 12},
	}, {
		v:   `012:34`,
		exp: Clock{hour: 12, min: 34},
	}, {
		v:   `024:061:88`,
		exp: Clock{},
	}, {
		v:   `23:59:59`,
		exp: Clock{hour: 23, min: 59, sec: 59},
	}}

	var (
		c   testCase
		got Clock
	)
	for _, c = range cases {
		got = ParseClock(c.v)
		test.Assert(t, c.v, c.exp, got)
	}
}
