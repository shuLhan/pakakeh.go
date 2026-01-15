// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package numbers

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestInt64CreateSeq(t *testing.T) {
	exp := []int64{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}
	got := Int64CreateSeq(-5, 5)

	test.Assert(t, "", exp, got)
}
