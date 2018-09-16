// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	dInts = [][]int{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 1, 0, 1, 0, 1, 0, 1, 0},
		{1, 1, 2, 2, 3, 1, 2},
	}
	dIntsSorted = [][]int{
		{},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		{0, 0, 0, 0, 0, 1, 1, 1, 1},
		{1, 1, 1, 2, 2, 2, 3},
	}
	dIntsSortedDesc = [][]int{
		{},
		{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		{1, 1, 1, 1, 0, 0, 0, 0, 0},
		{3, 2, 2, 2, 1, 1, 1},
	}
)

func TestIntsFindMaxEmpty(t *testing.T) {
	maxv, maxi, ok := IntsFindMax(dInts[0])

	test.Assert(t, "", -1, maxv, true)
	test.Assert(t, "", -1, maxi, true)
	test.Assert(t, "", false, ok, true)
}

func TestIntsFindMax(t *testing.T) {
	maxv, maxi, ok := IntsFindMax(dInts[1])

	test.Assert(t, "", 9, maxv, true)
	test.Assert(t, "", 4, maxi, true)
	test.Assert(t, "", true, ok, true)
}

func TestIntsFindMinEmpty(t *testing.T) {
	minv, mini, ok := IntsFindMin(dInts[0])

	test.Assert(t, "", -1, minv, true)
	test.Assert(t, "", -1, mini, true)
	test.Assert(t, "", false, ok, true)
}

func TestIntsFindMin(t *testing.T) {
	minv, mini, ok := IntsFindMin(dInts[1])

	test.Assert(t, "", 0, minv, true)
	test.Assert(t, "", 5, mini, true)
	test.Assert(t, "", true, ok, true)
}

func TestIntsSum(t *testing.T) {
	got := IntsSum(dInts[1])

	test.Assert(t, "", 45, got, true)
}

func TestIntsCount(t *testing.T) {
	got := IntsCount(dInts[0], 0)

	test.Assert(t, "", 0, got, true)

	got = IntsCount(dInts[1], 1)

	test.Assert(t, "", 1, got, true)

	got = IntsCount(dInts[2], 1)

	test.Assert(t, "", 4, got, true)

	got = IntsCount(dInts[3], 0)

	test.Assert(t, "", 0, got, true)

	got = IntsCount(dInts[3], 3)

	test.Assert(t, "", 1, got, true)
}

func TestIntsCountsEmpty(t *testing.T) {
	classes := []int{1, 2, 3}
	exp := []int{0, 0, 0}

	got := IntsCounts(dInts[0], classes)

	test.Assert(t, "", exp, got, true)
}

func TestIntsCountsEmptyClasses(t *testing.T) {
	classes := []int{}
	var exp []int

	got := IntsCounts(dInts[1], classes)

	test.Assert(t, "", exp, got, true)
}

func TestIntsCounts(t *testing.T) {
	classes := []int{1, 2, 3}
	exp := []int{3, 3, 1}

	got := IntsCounts(dInts[3], classes)

	test.Assert(t, "", exp, got, true)
}

func TestIntsMaxCountOf(t *testing.T) {
	classes := []int{0, 1}
	exp := int(0)
	got, _ := IntsMaxCountOf(dInts[2], classes)

	test.Assert(t, "", exp, got, true)

	// Swap the class values.
	classes = []int{1, 0}
	got, _ = IntsMaxCountOf(dInts[2], classes)

	test.Assert(t, "", exp, got, true)
}

func TestIntsSwapEmpty(t *testing.T) {
	exp := []int{}

	IntsSwap(dInts[0], 1, 6)

	test.Assert(t, "", exp, dInts[0], true)
}

func TestIntsSwapEqual(t *testing.T) {
	in := make([]int, len(dInts[1]))
	copy(in, dInts[1])

	exp := make([]int, len(in))
	copy(exp, in)

	IntsSwap(in, 1, 1)

	test.Assert(t, "", exp, in, true)
}

func TestIntsSwapOutOfRange(t *testing.T) {
	in := make([]int, len(dInts[1]))
	copy(in, dInts[1])

	exp := make([]int, len(in))
	copy(exp, in)

	IntsSwap(in, 1, 100)

	test.Assert(t, "", exp, in, true)
}

func TestIntsSwap(t *testing.T) {
	in := make([]int, len(dInts[1]))
	copy(in, dInts[1])

	exp := make([]int, len(in))
	copy(exp, in)

	IntsSwap(in, 0, len(in)-1)

	test.Assert(t, "", exp, in, false)

	tmp := exp[0]
	exp[0] = exp[len(exp)-1]
	exp[len(exp)-1] = tmp

	test.Assert(t, "", exp, in, true)
}

func TestIntsIsExist(t *testing.T) {
	var s bool

	// True positive.
	for _, d := range dInts {
		for _, v := range d {
			s = IntsIsExist(d, v)

			test.Assert(t, "", true, s, true)
		}
	}

	// False positive.
	for _, d := range dInts {
		s = IntsIsExist(d, -1)
		test.Assert(t, "", false, s, true)
		s = IntsIsExist(d, 10)
		test.Assert(t, "", false, s, true)
	}
}

func TestIntsInsertionSort(t *testing.T) {
	for x := range dInts {
		d := make([]int, len(dInts[x]))

		copy(d, dInts[x])

		ids := make([]int, len(d))
		for x := range ids {
			ids[x] = x
		}

		IntsInsertionSort(d, ids, 0, len(ids), true)

		test.Assert(t, "", dIntsSorted[x], d, true)
	}
}

func TestIntsInsertionSortDesc(t *testing.T) {
	for x := range dInts {
		d := make([]int, len(dInts[x]))

		copy(d, dInts[x])

		ids := make([]int, len(d))
		for x := range ids {
			ids[x] = x
		}

		IntsInsertionSort(d, ids, 0, len(ids), false)

		test.Assert(t, "", dIntsSortedDesc[x], d, true)
	}
}

func TestIntsSortByIndex(t *testing.T) {
	ids := [][]int{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 2, 4, 6, 8, 1, 3, 5, 7},
		{0, 1, 5, 6, 2, 3, 4},
	}

	for x := range dInts {
		d := make([]int, len(dInts[x]))

		copy(d, dInts[x])

		IntsSortByIndex(&d, ids[x])

		test.Assert(t, "", dIntsSorted[x], d, true)
	}
}

var intsInSorts = [][]int{
	{9, 8, 7, 6, 5, 4, 3},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 6, 7, 8, 5, 1, 2, 3, 4, 9},
	{9, 8, 7, 6, 5, 4, 3, 2, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{51, 50, 56, 55, 55, 58, 55, 55, 58, 56,
		57, 50, 56, 59, 62, 60, 49, 63, 61, 56,
		58, 67, 61, 59, 60, 49, 56, 52, 61, 64,
		70, 57, 65, 69, 57, 64, 62, 66, 63, 62,
		54, 67, 61, 57, 55, 60, 30, 66, 57, 60,
		68, 60, 61, 63, 58, 58, 56, 57, 60, 69,
		69, 64, 63, 63, 67, 65, 58, 63, 64, 67,
		59, 72, 63, 63, 65, 71, 67, 76, 73, 64,
		67, 74, 60, 68, 65, 64, 67, 64, 65, 69,
		77, 67, 72, 77, 72, 77, 61, 79, 77, 68,
		62},
}

var intsExpSorts = [][]int{
	{3, 4, 5, 6, 7, 8, 9},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{30, 49, 49, 50, 50, 51, 52, 54, 55, 55,
		55, 55, 55, 56, 56, 56, 56, 56, 56, 57,
		57, 57, 57, 57, 57, 58, 58, 58, 58, 58,
		58, 59, 59, 59, 60, 60, 60, 60, 60, 60,
		60, 61, 61, 61, 61, 61, 61, 62, 62, 62,
		62, 63, 63, 63, 63, 63, 63, 63, 63, 64,
		64, 64, 64, 64, 64, 64, 65, 65, 65, 65,
		65, 66, 66, 67, 67, 67, 67, 67, 67, 67,
		67, 68, 68, 68, 69, 69, 69, 69, 70, 71,
		72, 72, 72, 73, 74, 76, 77, 77, 77, 77,
		79},
}

var intsExpSortsDesc = [][]int{
	{9, 8, 7, 6, 5, 4, 3},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{9, 8, 7, 6, 5, 4, 3, 2, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{79, 77, 77, 77, 77, 76, 74, 73, 72, 72,
		72, 71, 70, 69, 69, 69, 69, 68, 68, 68,
		67, 67, 67, 67, 67, 67, 67, 67, 66, 66,
		65, 65, 65, 65, 65, 64, 64, 64, 64, 64,
		64, 64, 63, 63, 63, 63, 63, 63, 63, 63,
		62, 62, 62, 62, 61, 61, 61, 61, 61, 61,
		60, 60, 60, 60, 60, 60, 60, 59, 59, 59,
		58, 58, 58, 58, 58, 58, 57, 57, 57, 57,
		57, 57, 56, 56, 56, 56, 56, 56, 55, 55,
		55, 55, 55, 54, 52, 51, 50, 50, 49, 49,
		30},
}

func TestIntsIndirectSort(t *testing.T) {
	var res, exp string

	for i := range intsInSorts {
		IntsIndirectSort(intsInSorts[i], true)

		res = fmt.Sprint(intsInSorts[i])
		exp = fmt.Sprint(intsExpSorts[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestIntsIndirectSortDesc(t *testing.T) {
	var res, exp string

	for i := range intsInSorts {
		IntsIndirectSort(intsInSorts[i], false)

		res = fmt.Sprint(intsInSorts[i])
		exp = fmt.Sprint(intsExpSortsDesc[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestIntsIndirectSort_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IntsIndirectSort(intsInSorts[5], true)

	test.Assert(t, "", exp, got, true)
}

func TestIntsIndirectSortDesc_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IntsIndirectSort(intsInSorts[5], false)

	test.Assert(t, "", exp, got, true)
}

func TestIntsInplaceMergesort(t *testing.T) {
	size := len(intsInSorts[6])
	idx := make([]int, size)

	IntsInplaceMergesort(intsInSorts[6], idx, 0, size, true)

	test.Assert(t, "", intsExpSorts[6], intsInSorts[6], true)
}

func TestIntsIndirectSort_SortByIndex(t *testing.T) {
	expIds := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in1 := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in2 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	exp := fmt.Sprint(in1)

	sortedIds := IntsIndirectSort(in1, true)

	test.Assert(t, "", expIds, sortedIds, true)

	// Reverse the sort.
	IntsSortByIndex(&in2, sortedIds)

	got := fmt.Sprint(in2)

	test.Assert(t, "", exp, got, true)
}
