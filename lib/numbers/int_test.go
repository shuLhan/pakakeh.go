// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestIntCreateSeq(t *testing.T) {
	exp := []int{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}
	got := IntCreateSeq(-5, 5)

	test.Assert(t, "", exp, got)
}

func TestIntPickRandPositive(t *testing.T) {
	pickedListID := []int{0, 1, 2, 3, 4, 5, 7}
	exsListID := []int{8, 9}
	exp := 6

	// Pick random without duplicate.
	got := IntPickRandPositive(7, false, pickedListID, nil)

	test.Assert(t, "", exp, got)

	// Pick random with exclude indices.
	got = IntPickRandPositive(9, false, pickedListID, exsListID)
	test.Assert(t, "", exp, got)
}
