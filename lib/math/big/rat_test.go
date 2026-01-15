// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package big

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestRat_IsEqual_unexported(t *testing.T) {
	type A struct {
		r *Rat
	}

	exp := &A{
		r: NewRat(10),
	}

	cases := []struct {
		got *A
	}{{
		got: &A{
			r: NewRat(10),
		},
	}}

	for _, c := range cases {
		test.Assert(t, "IsEqual unexported field", exp, c.got)
	}
}
