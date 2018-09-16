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
	dInts64 = [][]int64{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 1, 0, 1, 0, 1, 0, 1, 0},
		{1, 1, 2, 2, 3, 1, 2},
	}
	dInts64Sorted = [][]int64{
		{},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		{0, 0, 0, 0, 0, 1, 1, 1, 1},
		{1, 1, 1, 2, 2, 2, 3},
	}
	dInts64SortedDesc = [][]int64{
		{},
		{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		{1, 1, 1, 1, 0, 0, 0, 0, 0},
		{3, 2, 2, 2, 1, 1, 1},
	}
)

func TestInts64FindMaxEmpty(t *testing.T) {
	gotv, goti, gotok := Ints64FindMax(dInts64[0])

	test.Assert(t, "", int64(-1), gotv, true)
	test.Assert(t, "", -1, goti, true)
	test.Assert(t, "", false, gotok, true)
}

func TestInts64FindMax(t *testing.T) {
	gotv, goti, gotok := Ints64FindMax(dInts64[1])

	test.Assert(t, "", int64(9), gotv, true)
	test.Assert(t, "", 4, goti, true)
	test.Assert(t, "", true, gotok, true)
}

func TestInts64FindMinEmpty(t *testing.T) {
	gotv, goti, gotok := Ints64FindMin(dInts64[0])

	test.Assert(t, "", int64(-1), gotv, true)
	test.Assert(t, "", -1, goti, true)
	test.Assert(t, "", false, gotok, true)
}

func TestInts64FindMin(t *testing.T) {
	gotv, goti, gotok := Ints64FindMin(dInts64[1])

	test.Assert(t, "", int64(0), gotv, true)
	test.Assert(t, "", 5, goti, true)
	test.Assert(t, "", true, gotok, true)
}

func TestInts64Sum(t *testing.T) {
	got := Ints64Sum(dInts64[1])

	test.Assert(t, "", int64(45), got, true)
}

func TestInts64Count(t *testing.T) {
	got := Ints64Count(dInts64[0], 0)

	test.Assert(t, "", 0, got, true)

	got = Ints64Count(dInts64[1], 1)

	test.Assert(t, "", 1, got, true)

	got = Ints64Count(dInts64[2], 1)

	test.Assert(t, "", 4, got, true)

	got = Ints64Count(dInts64[3], 0)

	test.Assert(t, "", 0, got, true)

	got = Ints64Count(dInts64[3], 3)

	test.Assert(t, "", 1, got, true)
}

func TestInts64CountsEmpty(t *testing.T) {
	classes := []int64{1, 2, 3}
	exp := []int{0, 0, 0}

	got := Ints64Counts(dInts64[0], classes)

	test.Assert(t, "", exp, got, true)
}

func TestInts64CountsEmptyClasses(t *testing.T) {
	classes := []int64{}
	var exp []int

	got := Ints64Counts(dInts64[1], classes)

	test.Assert(t, "", exp, got, true)
}

func TestInts64Counts(t *testing.T) {
	classes := []int64{1, 2, 3}
	exp := []int{3, 3, 1}

	got := Ints64Counts(dInts64[3], classes)

	test.Assert(t, "", exp, got, true)
}

func TestInts6464MaxCountOf(t *testing.T) {
	classes := []int64{0, 1}
	exp := int64(0)
	got, _ := Ints64MaxCountOf(dInts64[2], classes)

	test.Assert(t, "", exp, got, true)

	// Swap the class values.
	classes = []int64{1, 0}
	got, _ = Ints64MaxCountOf(dInts64[2], classes)

	test.Assert(t, "", exp, got, true)
}

func TestInts64SwapEmpty(t *testing.T) {
	exp := []int64{}

	Ints64Swap(dInts64[0], 1, 6)

	test.Assert(t, "", exp, dInts64[0], true)
}

func TestInts64SwapEqual(t *testing.T) {
	in := make([]int64, len(dInts64[1]))
	copy(in, dInts64[1])

	exp := make([]int64, len(in))
	copy(exp, in)

	Ints64Swap(in, 1, 1)

	test.Assert(t, "", exp, in, true)
}

func TestInts64SwapOutOfRange(t *testing.T) {
	in := make([]int64, len(dInts64[1]))
	copy(in, dInts64[1])

	exp := make([]int64, len(in))
	copy(exp, in)

	Ints64Swap(in, 1, 100)

	test.Assert(t, "", exp, in, true)
}

func TestInts64Swap(t *testing.T) {
	in := make([]int64, len(dInts64[1]))
	copy(in, dInts64[1])

	exp := make([]int64, len(in))
	copy(exp, in)

	Ints64Swap(in, 0, len(in)-1)

	test.Assert(t, "", exp, in, false)

	tmp := exp[0]
	exp[0] = exp[len(exp)-1]
	exp[len(exp)-1] = tmp

	test.Assert(t, "", exp, in, true)
}

func TestInts64IsExist(t *testing.T) {
	var s bool

	// True positive.
	for _, d := range dInts64 {
		for _, v := range d {
			s = Ints64IsExist(d, v)

			test.Assert(t, "", true, s, true)
		}
	}

	// False positive.
	for _, d := range dInts64 {
		s = Ints64IsExist(d, -1)
		test.Assert(t, "", false, s, true)
		s = Ints64IsExist(d, 10)
		test.Assert(t, "", false, s, true)
	}
}

func TestInts64InsertionSort(t *testing.T) {
	for x := range dInts64 {
		d := make([]int64, len(dInts64[x]))

		copy(d, dInts64[x])

		ids := make([]int, len(d))
		for x := range ids {
			ids[x] = x
		}

		Ints64InsertionSort(d, ids, 0, len(ids), true)

		test.Assert(t, "", dInts64Sorted[x], d, true)
	}
}

func TestInts64InsertionSortDesc(t *testing.T) {
	for x := range dInts64 {
		d := make([]int64, len(dInts64[x]))

		copy(d, dInts64[x])

		ids := make([]int, len(d))
		for x := range ids {
			ids[x] = x
		}

		Ints64InsertionSort(d, ids, 0, len(ids), false)

		test.Assert(t, "", dInts64SortedDesc[x], d, true)
	}
}

func TestInts64SortByIndex(t *testing.T) {
	ids := [][]int{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 2, 4, 6, 8, 1, 3, 5, 7},
		{0, 1, 5, 6, 2, 3, 4},
	}

	for x := range dInts64 {
		d := make([]int64, len(dInts64[x]))

		copy(d, dInts64[x])

		Ints64SortByIndex(&d, ids[x])

		test.Assert(t, "", dInts64Sorted[x], d, true)
	}
}

var ints64InSorts = [][]int64{
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

var ints64ExpSorts = [][]int64{
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

var ints64ExpSortsDesc = [][]int64{
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

func TestInts64IndirectSort(t *testing.T) {
	var res, exp string

	for i := range ints64InSorts {
		Ints64IndirectSort(ints64InSorts[i], true)

		res = fmt.Sprint(ints64InSorts[i])
		exp = fmt.Sprint(ints64ExpSorts[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestInts64IndirectSortDesc(t *testing.T) {
	var res, exp string

	for i := range ints64InSorts {
		Ints64IndirectSort(ints64InSorts[i], false)

		res = fmt.Sprint(ints64InSorts[i])
		exp = fmt.Sprint(ints64ExpSortsDesc[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestInts64IndirectSort_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := Ints64IndirectSort(ints64InSorts[5], true)

	test.Assert(t, "", exp, got, true)
}

func TestInts64IndirectSortDesc_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := Ints64IndirectSort(ints64InSorts[5], false)

	test.Assert(t, "", exp, got, true)
}

func TestInts64InplaceMergesort(t *testing.T) {
	size := len(ints64InSorts[6])
	idx := make([]int, size)

	Ints64InplaceMergesort(ints64InSorts[6], idx, 0, size, true)

	test.Assert(t, "", ints64ExpSorts[6], ints64InSorts[6], true)
}

func TestInts64IndirectSort_SortByIndex(t *testing.T) {
	expIds := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in1 := []int64{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in2 := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	exp := fmt.Sprint(in1)

	sortedIds := Ints64IndirectSort(in1, true)

	test.Assert(t, "", expIds, sortedIds, true)

	// Reverse the sort.
	Ints64SortByIndex(&in2, sortedIds)

	got := fmt.Sprint(in2)

	test.Assert(t, "", exp, got, true)
}
