// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ints

import (
	"fmt"
	"sort"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	d = [][]int{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 1, 0, 1, 0, 1, 0, 1, 0},
		{1, 1, 2, 2, 3, 1, 2},
	}
	dSorted = [][]int{
		{},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		{0, 0, 0, 0, 0, 1, 1, 1, 1},
		{1, 1, 1, 2, 2, 2, 3},
	}
	dSortedDesc = [][]int{
		{},
		{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		{1, 1, 1, 1, 0, 0, 0, 0, 0},
		{3, 2, 2, 2, 1, 1, 1},
	}
)

func TestMaxEmpty(t *testing.T) {
	maxv, maxi, ok := Max(d[0])

	test.Assert(t, "", 0, maxv, true)
	test.Assert(t, "", 0, maxi, true)
	test.Assert(t, "", false, ok, true)
}

func TestMax(t *testing.T) {
	maxv, maxi, ok := Max(d[1])

	test.Assert(t, "", 9, maxv, true)
	test.Assert(t, "", 4, maxi, true)
	test.Assert(t, "", true, ok, true)
}

func TestMinEmpty(t *testing.T) {
	minv, mini, ok := Min(d[0])

	test.Assert(t, "", 0, minv, true)
	test.Assert(t, "", 0, mini, true)
	test.Assert(t, "", false, ok, true)
}

func TestMin(t *testing.T) {
	minv, mini, ok := Min(d[1])

	test.Assert(t, "", 0, minv, true)
	test.Assert(t, "", 5, mini, true)
	test.Assert(t, "", true, ok, true)
}

func TestSum(t *testing.T) {
	got := Sum(d[1])

	test.Assert(t, "", 45, got, true)
}

func TestCount(t *testing.T) {
	got := Count(d[0], 0)

	test.Assert(t, "", 0, got, true)

	got = Count(d[1], 1)

	test.Assert(t, "", 1, got, true)

	got = Count(d[2], 1)

	test.Assert(t, "", 4, got, true)

	got = Count(d[3], 0)

	test.Assert(t, "", 0, got, true)

	got = Count(d[3], 3)

	test.Assert(t, "", 1, got, true)
}

func TestCountsEmpty(t *testing.T) {
	classes := []int{1, 2, 3}
	exp := []int{0, 0, 0}

	got := Counts(d[0], classes)

	test.Assert(t, "", exp, got, true)
}

func TestCountsEmptyClasses(t *testing.T) {
	classes := []int{}
	var exp []int

	got := Counts(d[1], classes)

	test.Assert(t, "", exp, got, true)
}

func TestCounts(t *testing.T) {
	classes := []int{1, 2, 3}
	exp := []int{3, 3, 1}

	got := Counts(d[3], classes)

	test.Assert(t, "", exp, got, true)
}

func TestMaxCountOf(t *testing.T) {
	classes := []int{0, 1}
	exp := int(0)
	got, _ := MaxCountOf(d[2], classes)

	test.Assert(t, "", exp, got, true)

	// Swap the class values.
	classes = []int{1, 0}
	got, _ = MaxCountOf(d[2], classes)

	test.Assert(t, "", exp, got, true)
}

func TestSwapEmpty(t *testing.T) {
	exp := []int{}

	Swap(d[0], 1, 6)

	test.Assert(t, "", exp, d[0], true)
}

func TestSwapEqual(t *testing.T) {
	in := make([]int, len(d[1]))
	copy(in, d[1])

	exp := make([]int, len(in))
	copy(exp, in)

	Swap(in, 1, 1)

	test.Assert(t, "", exp, in, true)
}

func TestSwapOutOfRange(t *testing.T) {
	in := make([]int, len(d[1]))
	copy(in, d[1])

	exp := make([]int, len(in))
	copy(exp, in)

	Swap(in, 1, 100)

	test.Assert(t, "", exp, in, true)
}

func TestSwap(t *testing.T) {
	in := make([]int, len(d[1]))
	copy(in, d[1])

	exp := make([]int, len(in))
	copy(exp, in)

	Swap(in, 0, len(in)-1)

	test.Assert(t, "", exp, in, false)

	tmp := exp[0]
	exp[0] = exp[len(exp)-1]
	exp[len(exp)-1] = tmp

	test.Assert(t, "", exp, in, true)
}

func TestIsExist(t *testing.T) {
	var s bool

	// True positive.
	for _, d := range d {
		for _, v := range d {
			s = IsExist(d, v)

			test.Assert(t, "", true, s, true)
		}
	}

	// False positive.
	for _, d := range d {
		s = IsExist(d, -1)
		test.Assert(t, "", false, s, true)
		s = IsExist(d, 10)
		test.Assert(t, "", false, s, true)
	}
}

func TestInplaceInsertionSort(t *testing.T) {
	for x := range d {
		data := make([]int, len(d[x]))

		copy(data, d[x])

		ids := make([]int, len(data))
		for x := range ids {
			ids[x] = x
		}

		InplaceInsertionSort(data, ids, 0, len(ids), true)

		test.Assert(t, "", dSorted[x], data, true)
	}
}

func TestInplaceInsertionSortDesc(t *testing.T) {
	for x := range d {
		data := make([]int, len(d[x]))

		copy(data, d[x])

		ids := make([]int, len(data))
		for x := range ids {
			ids[x] = x
		}

		InplaceInsertionSort(data, ids, 0, len(ids), false)

		test.Assert(t, "", dSortedDesc[x], data, true)
	}
}

func TestSortByIndex(t *testing.T) {
	ids := [][]int{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 2, 4, 6, 8, 1, 3, 5, 7},
		{0, 1, 5, 6, 2, 3, 4},
	}

	for x := range d {
		data := make([]int, len(d[x]))

		copy(data, d[x])

		SortByIndex(&data, ids[x])

		test.Assert(t, "", dSorted[x], data, true)
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

func TestCompareSort(t *testing.T) {
	d1 := make([]int, n)
	generateRandomInts(d1)
	d2 := make([]int, len(d1))
	copy(d2, d1)

	sort.Ints(d1)
	IndirectSort(d2, true)

	test.Assert(t, "Compare sort", d1, d2, true)
}

func TestIndirectSort2(t *testing.T) {
	var res, exp string

	for i := range intsInSorts {
		IndirectSort(intsInSorts[i], true)

		res = fmt.Sprint(intsInSorts[i])
		exp = fmt.Sprint(intsExpSorts[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestIndirectSortDesc(t *testing.T) {
	var res, exp string

	for i := range intsInSorts {
		IndirectSort(intsInSorts[i], false)

		res = fmt.Sprint(intsInSorts[i])
		exp = fmt.Sprint(intsExpSortsDesc[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestIndirectSort_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IndirectSort(intsInSorts[5], true)

	test.Assert(t, "", exp, got, true)
}

func TestIndirectSortDesc_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IndirectSort(intsInSorts[5], false)

	test.Assert(t, "", exp, got, true)
}

func TestInplaceMergesort(t *testing.T) {
	size := len(intsInSorts[6])
	idx := make([]int, size)

	InplaceMergesort(intsInSorts[6], idx, 0, size, true)

	test.Assert(t, "", intsExpSorts[6], intsInSorts[6], true)
}

func TestIndirectSort_SortByIndex(t *testing.T) {
	expIds := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in1 := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in2 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	exp := fmt.Sprint(in1)

	sortedIds := IndirectSort(in1, true)

	test.Assert(t, "", expIds, sortedIds, true)

	// Reverse the sort.
	SortByIndex(&in2, sortedIds)

	got := fmt.Sprint(in2)

	test.Assert(t, "", exp, got, true)
}

func TestRemove(t *testing.T) {
	cases := []struct {
		d   []int
		v   int
		exp []int
	}{{
		d:   []int{},
		v:   1,
		exp: []int{},
	}, {
		d:   []int{1},
		v:   1,
		exp: []int{},
	}, {
		d:   []int{1, 2, 3, 4},
		v:   5,
		exp: []int{1, 2, 3, 4},
	}, {
		d:   []int{1, 2, 3, 4},
		v:   1,
		exp: []int{2, 3, 4},
	}}

	for _, c := range cases {
		got, _ := Remove(c.d, c.v)

		test.Assert(t, "Remove", c.exp, got, true)
	}
}
