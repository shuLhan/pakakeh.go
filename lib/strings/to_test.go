// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package strings

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestToFloat64(t *testing.T) {
	var (
		in  = []string{`0`, `1.1`, `e`, `3`}
		exp = []float64{0, 1.1, 0, 3}
		got = ToFloat64(in)
	)

	test.Assert(t, ``, exp, got)
}

func TestToInt64(t *testing.T) {
	var (
		in  = []string{`0`, `1`, `e`, `3.3`}
		exp = []int64{0, 1, 0, 3}
		got = ToInt64(in)
	)

	test.Assert(t, ``, exp, got)
}

func TestToStrings(t *testing.T) {
	var (
		is  = make([]any, 0)
		i64 = []int64{0, 1, 2, 3}
		exp = []string{`0`, `1`, `2`, `3`}

		v   int64
		got []string
	)

	for _, v = range i64 {
		is = append(is, v)
	}

	got = ToStrings(is)

	test.Assert(t, ``, exp, got)
}
