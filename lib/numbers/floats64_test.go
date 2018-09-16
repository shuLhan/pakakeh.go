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
	dFloats64 = [][]float64{
		{},
		{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
		{1, 1, 2, 2, 3, 1, 2},
	}
	dFloats64Sorted = [][]float64{
		{},
		{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9},
		{0.0, 0.0, 0.0, 0.0, 0.0, 0.1, 0.1, 0.1, 0.1},
		{1, 1, 1, 2, 2, 2, 3},
	}
	dFloats64SortedDesc = [][]float64{
		{},
		{0.9, 0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1, 0.0},
		{0.1, 0.1, 0.1, 0.1, 0.0, 0.0, 0.0, 0.0, 0.0},
		{3, 2, 2, 2, 1, 1, 1},
	}
)

func TestFloats64FindMaxEmpty(t *testing.T) {
	gotv, goti, gotok := Floats64FindMax(dFloats64[0])

	test.Assert(t, "", float64(-1), gotv, true)
	test.Assert(t, "", -1, goti, true)
	test.Assert(t, "", false, gotok, true)
}

func TestFloats64FindMax(t *testing.T) {
	gotv, goti, gotok := Floats64FindMax(dFloats64[1])

	test.Assert(t, "", float64(0.9), gotv, true)
	test.Assert(t, "", 4, goti, true)
	test.Assert(t, "", true, gotok, true)
}

func TestFloats64FindMinEmpty(t *testing.T) {
	gotv, goti, gotok := Floats64FindMin(dFloats64[0])

	test.Assert(t, "", gotv, float64(-1), true)
	test.Assert(t, "", goti, -1, true)
	test.Assert(t, "", gotok, false, true)
}

func TestFloats64FindMin(t *testing.T) {
	gotv, goti, gotok := Floats64FindMin(dFloats64[1])

	test.Assert(t, "", gotv, float64(0.0), true)
	test.Assert(t, "", goti, 5, true)
	test.Assert(t, "", gotok, true, true)
}

func TestFloats64Sum(t *testing.T) {
	got := Floats64Sum(dFloats64[1])

	test.Assert(t, "", float64(4.5), Float64Round(got, 1), true)
}

func TestFloats64Count(t *testing.T) {
	got := Floats64Count(dFloats64[0], 0)

	test.Assert(t, "", 0, got, true)

	got = Floats64Count(dFloats64[1], 0.1)

	test.Assert(t, "", 1, got, true)

	got = Floats64Count(dFloats64[2], 0.1)

	test.Assert(t, "", 4, got, true)

	got = Floats64Count(dFloats64[3], 0.1)

	test.Assert(t, "", 0, got, true)

	got = Floats64Count(dFloats64[3], 3)

	test.Assert(t, "", 1, got, true)
}

func TestFloats64CountsEmpty(t *testing.T) {
	classes := []float64{1, 2, 3}
	exp := []int{0, 0, 0}

	got := Floats64Counts(dFloats64[0], classes)

	test.Assert(t, "", exp, got, true)
}

func TestFloats64CountsEmptyClasses(t *testing.T) {
	classes := []float64{}
	var exp []int

	got := Floats64Counts(dFloats64[1], classes)

	test.Assert(t, "", exp, got, true)
}

func TestFloats64Counts(t *testing.T) {
	classes := []float64{1, 2, 3}
	exp := []int{3, 3, 1}

	got := Floats64Counts(dFloats64[3], classes)

	test.Assert(t, "", exp, got, true)
}

func TestFloats64MaxCountOf(t *testing.T) {
	classes := []float64{0, 1}
	exp := float64(0)
	got, _ := Floats64MaxCountOf(dFloats64[2], classes)

	test.Assert(t, "", exp, got, true)

	// Swap the class values.
	classes = []float64{1, 0}
	got, _ = Floats64MaxCountOf(dFloats64[2], classes)

	test.Assert(t, "", exp, got, true)
}

func TestFloats64SwapEmpty(t *testing.T) {
	exp := []float64{}

	Floats64Swap(dFloats64[0], 1, 6)

	test.Assert(t, "", exp, dFloats64[0], true)
}

func TestFloats64SwapEqual(t *testing.T) {
	in := make([]float64, len(dFloats64[1]))
	copy(in, dFloats64[1])

	exp := make([]float64, len(in))
	copy(exp, in)

	Floats64Swap(in, 1, 1)

	test.Assert(t, "", exp, in, true)
}

func TestFloats64SwapOutOfRange(t *testing.T) {
	in := make([]float64, len(dFloats64[1]))
	copy(in, dFloats64[1])

	exp := make([]float64, len(in))
	copy(exp, in)

	Floats64Swap(in, 1, 100)

	test.Assert(t, "", exp, in, true)
}

func TestFloats64Swap(t *testing.T) {
	in := make([]float64, len(dFloats64[1]))
	copy(in, dFloats64[1])

	exp := make([]float64, len(in))
	copy(exp, in)

	Floats64Swap(in, 0, len(in)-1)

	test.Assert(t, "", exp, in, false)

	tmp := exp[0]
	exp[0] = exp[len(exp)-1]
	exp[len(exp)-1] = tmp

	test.Assert(t, "", exp, in, true)
}

func TestFloats64IsExist(t *testing.T) {
	got := Floats64IsExist(dFloats64[0], 0)

	test.Assert(t, "", false, got, true)

	got = Floats64IsExist(dFloats64[1], float64(0))

	test.Assert(t, "", true, got, true)

	got = Floats64IsExist(dFloats64[1], float64(0.01))

	test.Assert(t, "", false, got, true)
}

func TestFloats64InsertionSort(t *testing.T) {
	for x := range dFloats64 {
		d := make([]float64, len(dFloats64[x]))

		copy(d, dFloats64[x])

		ids := make([]int, len(d))
		for x := range ids {
			ids[x] = x
		}

		Floats64InsertionSort(d, ids, 0, len(ids), true)

		test.Assert(t, "", dFloats64Sorted[x], d, true)
	}
}

func TestFloats64InsertionSortDesc(t *testing.T) {
	for x := range dFloats64 {
		d := make([]float64, len(dFloats64[x]))

		copy(d, dFloats64[x])

		ids := make([]int, len(d))
		for x := range ids {
			ids[x] = x
		}

		Floats64InsertionSort(d, ids, 0, len(ids), false)

		test.Assert(t, "", dFloats64SortedDesc[x], d, true)
	}
}

func TestFloats64SortByIndex(t *testing.T) {
	ids := [][]int{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 2, 4, 6, 8, 1, 3, 5, 7},
		{0, 1, 5, 6, 2, 3, 4},
	}

	for x := range dFloats64 {
		d := make([]float64, len(dFloats64[x]))

		copy(d, dFloats64[x])

		Floats64SortByIndex(&d, ids[x])

		test.Assert(t, "", dFloats64Sorted[x], d, true)
	}
}

var inSorts = [][]float64{
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

var expSorts = [][]float64{
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

var expSortsDesc = [][]float64{
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

func TestFloats64IndirectSort(t *testing.T) {
	var res, exp string

	for i := range inSorts {
		Floats64IndirectSort(inSorts[i], true)

		res = fmt.Sprint(inSorts[i])
		exp = fmt.Sprint(expSorts[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestFloats64IndirectSortDesc(t *testing.T) {
	var res, exp string

	for i := range inSorts {
		Floats64IndirectSort(inSorts[i], false)

		res = fmt.Sprint(inSorts[i])
		exp = fmt.Sprint(expSortsDesc[i])

		test.Assert(t, "", exp, res, true)
	}
}

func TestFloats64IndirectSort_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := Floats64IndirectSort(inSorts[5], true)

	test.Assert(t, "", exp, got, true)
}

func TestFloats64IndirectSortDesc_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := Floats64IndirectSort(inSorts[5], false)

	test.Assert(t, "", exp, got, true)
}

func TestFloats64InplaceMergesort(t *testing.T) {
	size := len(inSorts[6])
	idx := make([]int, size)

	Floats64InplaceMergesort(inSorts[6], idx, 0, size, true)

	test.Assert(t, "", expSorts[6], inSorts[6], true)
}

func TestFloats64IndirectSort_SortByIndex(t *testing.T) {
	expIds := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in1 := []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0}
	in2 := []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}

	exp := fmt.Sprint(in1)

	sortedIds := Floats64IndirectSort(in1, true)

	test.Assert(t, "", expIds, sortedIds, true)

	// Reverse the sort.
	Floats64SortByIndex(&in2, sortedIds)

	got := fmt.Sprint(in2)

	test.Assert(t, "", exp, got, true)
}
