// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIntCreateSeq(t *testing.T) {
	exp := []int{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}
	got := IntCreateSeq(-5, 5)

	test.Assert(t, "", exp, got, true)
}

func TestIntPickRandPositive(t *testing.T) {
	pickedIds := []int{0, 1, 2, 3, 4, 5, 7}
	exsIds := []int{8, 9}
	exp := 6

	// Pick random without duplicate.
	got := IntPickRandPositive(7, false, pickedIds, nil)

	test.Assert(t, "", exp, got, true)

	// Pick random with exclude indices.
	got = IntPickRandPositive(9, false, pickedIds, exsIds)
	test.Assert(t, "", exp, got, true)
}
