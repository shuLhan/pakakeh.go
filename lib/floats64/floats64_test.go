// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package floats64

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/numbers"
	"github.com/shuLhan/share/lib/test"
)

var d = [][]float64{ //nolint: gochecknoglobals
	{},
	{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
	{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
	{1, 1, 2, 2, 3, 1, 2},
}
var dSorted = [][]float64{ //nolint: gochecknoglobals
	{},
	{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9},
	{0.0, 0.0, 0.0, 0.0, 0.0, 0.1, 0.1, 0.1, 0.1},
	{1, 1, 1, 2, 2, 2, 3},
}
var dSortedDesc = [][]float64{ //nolint: gochecknoglobals
	{},
	{0.9, 0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1, 0.0},
	{0.1, 0.1, 0.1, 0.1, 0.0, 0.0, 0.0, 0.0, 0.0},
	{3, 2, 2, 2, 1, 1, 1},
}

func TestMaxEmpty(t *testing.T) {
	gotv, goti, gotok := Max(d[0])

	test.Assert(t, "", float64(0), gotv, true)
	test.Assert(t, "", 0, goti, true)
	test.Assert(t, "", false, gotok, true)
}

func TestMax(t *testing.T) {
	gotv, goti, gotok := Max(d[1])

	test.Assert(t, "", float64(0.9), gotv, true)
	test.Assert(t, "", 4, goti, true)
	test.Assert(t, "", true, gotok, true)
}

func TestMinEmpty(t *testing.T) {
	gotv, goti, gotok := Min(d[0])

	test.Assert(t, "", gotv, float64(0), true)
	test.Assert(t, "", goti, 0, true)
	test.Assert(t, "", gotok, false, true)
}

func TestMin(t *testing.T) {
	gotv, goti, gotok := Min(d[1])

	test.Assert(t, "", gotv, float64(0.0), true)
	test.Assert(t, "", goti, 5, true)
	test.Assert(t, "", gotok, true, true)
}

func TestSum(t *testing.T) {
	got := Sum(d[1])

	test.Assert(t, "", float64(4.5), numbers.Float64Round(got, 1), true)
}

func TestCount(t *testing.T) {
	got := Count(d[0], 0)

	test.Assert(t, "", 0, got, true)

	got = Count(d[1], 0.1)

	test.Assert(t, "", 1, got, true)

	got = Count(d[2], 0.1)

	test.Assert(t, "", 4, got, true)

	got = Count(d[3], 0.1)

	test.Assert(t, "", 0, got, true)

	got = Count(d[3], 3)

	test.Assert(t, "", 1, got, true)
}

func TestCountsEmpty(t *testing.T) {
	classes := []float64{1, 2, 3}
	exp := []int{0, 0, 0}

	got := Counts(d[0], classes)

	test.Assert(t, "", exp, got, true)
}

func TestCountsEmptyClasses(t *testing.T) {
	classes := []float64{}
	var exp []int

	got := Counts(d[1], classes)

	test.Assert(t, "", exp, got, true)
}

func TestCounts(t *testing.T) {
	classes := []float64{1, 2, 3}
	exp := []int{3, 3, 1}

	got := Counts(d[3], classes)

	test.Assert(t, "", exp, got, true)
}

func TestMaxCountOf(t *testing.T) {
	classes := []float64{0, 1}
	exp := float64(0)
	got, _ := MaxCountOf(d[2], classes)

	test.Assert(t, "", exp, got, true)

	// Swap the class values.
	classes = []float64{1, 0}
	got, _ = MaxCountOf(d[2], classes)

	test.Assert(t, "", exp, got, true)
}

func TestSwapEmpty(t *testing.T) {
	exp := []float64{}

	Swap(d[0], 1, 6)

	test.Assert(t, "", exp, d[0], true)
}

func TestSwapEqual(t *testing.T) {
	in := make([]float64, len(d[1]))
	copy(in, d[1])

	exp := make([]float64, len(in))
	copy(exp, in)

	Swap(in, 1, 1)

	test.Assert(t, "", exp, in, true)
}

func TestSwapOutOfRange(t *testing.T) {
	in := make([]float64, len(d[1]))
	copy(in, d[1])

	exp := make([]float64, len(in))
	copy(exp, in)

	Swap(in, 1, 100)

	test.Assert(t, "", exp, in, true)
}

func TestSwap(t *testing.T) {
	in := make([]float64, len(d[1]))
	copy(in, d[1])

	exp := make([]float64, len(in))
	copy(exp, in)

	Swap(in, 0, len(in)-1)

	test.Assert(t, "", exp, in, false)

	tmp := exp[0]
	exp[0] = exp[len(exp)-1]
	exp[len(exp)-1] = tmp

	test.Assert(t, "", exp, in, true)
}

func TestIsExist(t *testing.T) {
	got := IsExist(d[0], 0)

	test.Assert(t, "", false, got, true)

	got = IsExist(d[1], float64(0))

	test.Assert(t, "", true, got, true)

	got = IsExist(d[1], float64(0.01))

	test.Assert(t, "", false, got, true)
}

func TestInplaceInsertionSort(t *testing.T) {
	for x := range d {
		data := make([]float64, len(d[x]))

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
		data := make([]float64, len(d[x]))

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
		data := make([]float64, len(d[x]))

		copy(data, d[x])

		SortByIndex(&data, ids[x])

		test.Assert(t, "", dSorted[x], data, true)
	}
}

var inSorts = [][]float64{ // nolint: gochecknoglobals,dupl
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0},
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
	{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
	{0.0, 6.0, 7.0, 8.0, 5.0, 1.0, 2.0, 3.0, 4.0, 9.0},
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{5.1, 5, 5.6, 5.5, 5.5, 5.8, 5.5, 5.5, 5.8, 5.6,
		5.7, 5, 5.6, 5.9, 6.2, 6, 4.9, 6.3, 6.1, 5.6,
		5.8, 6.7, 6.1, 5.9, 6, 4.9, 5.6, 5.2, 6.1, 6.4,
		7, 5.7, 6.5, 6.9, 5.7, 6.4, 6.2, 6.6, 6.3, 6.2,
		5.4, 6.7, 6.1, 5.7, 5.5, 6, 3, 6.6, 5.7, 6,
		6.8, 6, 6.1, 6.3, 5.8, 5.8, 5.6, 5.7, 6, 6.9,
		6.9, 6.4, 6.3, 6.3, 6.7, 6.5, 5.8, 6.3, 6.4, 6.7,
		5.9, 7.2, 6.3, 6.3, 6.5, 7.1, 6.7, 7.6, 7.3, 6.4,
		6.7, 7.4, 6, 6.8, 6.5, 6.4, 6.7, 6.4, 6.5, 6.9,
		7.7, 6.7, 7.2, 7.7, 7.2, 7.7, 6.1, 7.9, 7.7, 6.8,
		6.2},
}

var expSorts = [][]float64{ // nolint: gochecknoglobals,dupl
	{3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
	{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
	{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
	{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
	{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{3, 4.9, 4.9, 5, 5, 5.1, 5.2, 5.4, 5.5, 5.5,
		5.5, 5.5, 5.5, 5.6, 5.6, 5.6, 5.6, 5.6, 5.6, 5.7,
		5.7, 5.7, 5.7, 5.7, 5.7, 5.8, 5.8, 5.8, 5.8, 5.8,
		5.8, 5.9, 5.9, 5.9, 6, 6, 6, 6, 6, 6,
		6, 6.1, 6.1, 6.1, 6.1, 6.1, 6.1, 6.2, 6.2, 6.2,
		6.2, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.4,
		6.4, 6.4, 6.4, 6.4, 6.4, 6.4, 6.5, 6.5, 6.5, 6.5,
		6.5, 6.6, 6.6, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7,
		6.7, 6.8, 6.8, 6.8, 6.9, 6.9, 6.9, 6.9, 7, 7.1,
		7.2, 7.2, 7.2, 7.3, 7.4, 7.6, 7.7, 7.7, 7.7, 7.7,
		7.9},
}

var expSortsDesc = [][]float64{ // nolint: gochecknoglobals,dupl
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0},
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
	{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{7.9, 7.7, 7.7, 7.7, 7.7, 7.6, 7.4, 7.3, 7.2, 7.2,
		7.2, 7.1, 7, 6.9, 6.9, 6.9, 6.9, 6.8, 6.8, 6.8,
		6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.6, 6.6,
		6.5, 6.5, 6.5, 6.5, 6.5, 6.4, 6.4, 6.4, 6.4, 6.4,
		6.4, 6.4, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3,
		6.2, 6.2, 6.2, 6.2, 6.1, 6.1, 6.1, 6.1, 6.1, 6.1,
		6, 6, 6, 6, 6, 6, 6, 5.9, 5.9, 5.9,
		5.8, 5.8, 5.8, 5.8, 5.8, 5.8, 5.7, 5.7, 5.7, 5.7,
		5.7, 5.7, 5.6, 5.6, 5.6, 5.6, 5.6, 5.6, 5.5, 5.5,
		5.5, 5.5, 5.5, 5.4, 5.2, 5.1, 5, 5, 4.9, 4.9, 3},
}

func TestIndirectSort(t *testing.T) {
	var res, exp string

	for i := range inSorts {
		IndirectSort(inSorts[i], true)

		res = fmt.Sprint(inSorts[i])
		exp = fmt.Sprint(expSorts[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestIndirectSortDesc(t *testing.T) {
	var res, exp string

	for i := range inSorts {
		IndirectSort(inSorts[i], false)

		res = fmt.Sprint(inSorts[i])
		exp = fmt.Sprint(expSortsDesc[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestIndirectSort_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IndirectSort(inSorts[5], true)

	test.Assert(t, "", exp, got, true)
}

func TestIndirectSortDesc_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IndirectSort(inSorts[5], false)

	test.Assert(t, "", exp, got, true)
}

func TestInplaceMergesort(t *testing.T) {
	size := len(inSorts[6])
	idx := make([]int, size)

	InplaceMergesort(inSorts[6], idx, 0, size, true)

	test.Assert(t, "", expSorts[6], inSorts[6], true)
}

func TestIndirectSort_SortByIndex(t *testing.T) {
	expIds := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in1 := []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0}
	in2 := []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}

	exp := fmt.Sprint(in1)

	sortedIds := IndirectSort(in1, true)

	test.Assert(t, "", expIds, sortedIds, true)

	// Reverse the sort.
	SortByIndex(&in2, sortedIds)

	got := fmt.Sprint(in2)

	test.Assert(t, "", exp, got, true)
}
