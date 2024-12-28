// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package slices_test

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/numbers"
	"git.sr.ht/~shulhan/pakakeh.go/lib/slices"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestCount_float64(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []float64
		val   float64
		exp   int
	}{{
		desc: `empty`,
		exp:  0,
	}, {
		desc:  `case 1`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		val:   0.1,
		exp:   1,
	}, {
		slice: []float64{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
		val:   0.1,
		exp:   4,
	}, {
		slice: []float64{1, 1, 2, 2, 3, 1, 2},
		val:   0.1,
		exp:   0,
	}, {
		slice: []float64{1, 1, 2, 2, 3, 1, 2},
		val:   3,
		exp:   1,
	}}

	for _, tcase := range listCase {
		var got = slices.Count(tcase.slice, tcase.val)
		test.Assert(t, tcase.desc, tcase.exp, got)
	}
}

func TestCounts_float64(t *testing.T) {
	var listCase = []struct {
		desc    string
		slice   []float64
		classes []float64
		exp     []int
	}{{
		desc:    `empty slice`,
		classes: []float64{1, 2, 3},
		exp:     []int{0, 0, 0},
	}, {
		desc:  `empty classes`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
	}, {
		desc:    `ok`,
		slice:   []float64{1, 1, 2, 2, 3, 1, 2},
		classes: []float64{1, 2, 3},
		exp:     []int{3, 3, 1},
	}}

	for _, tcase := range listCase {
		var got = slices.Counts(tcase.slice, tcase.classes)
		test.Assert(t, tcase.desc, tcase.exp, got)
	}
}

func TestIndirectSort_float64(t *testing.T) {
	var listCase = []struct {
		desc         string
		slice        []float64
		expSortedIdx []int
		exp          []float64
		isAsc        bool
	}{{
		desc:         `case 1 asc`,
		slice:        []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0},
		exp:          []float64{3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
		expSortedIdx: []int{6, 5, 4, 3, 2, 1, 0},
		isAsc:        true,
	}, {
		desc:         `case 1 desc`,
		slice:        []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0},
		exp:          []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6},
		isAsc:        false,
	}, {
		desc:         `case 2 asc`,
		slice:        []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
		exp:          []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
		expSortedIdx: []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		isAsc:        true,
	}, {
		desc:         `case 2 desc`,
		slice:        []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
		exp:          []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc:        false,
	}, {
		desc:         `case 3 asc`,
		slice:        []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
		exp:          []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc:        true,
	}, {
		desc:         `case 3 desc`,
		slice:        []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
		exp:          []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0, 0.0},
		expSortedIdx: []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		isAsc:        false,
	}, {
		desc:         `case 4 asc`,
		slice:        []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0},
		exp:          []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
		expSortedIdx: []int{8, 7, 6, 5, 4, 3, 2, 1, 0},
		isAsc:        true,
	}, {
		desc:         `case 4 desc`,
		slice:        []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0},
		exp:          []float64{9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
		isAsc:        false,
	}, {
		desc:         `case 5 stable asc`,
		slice:        []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		exp:          []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc:        true,
	}, {
		desc:         `case 5 stable desc`,
		slice:        []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		exp:          []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc:        false,
	}, {
		desc: `case 6 asc`,
		slice: []float64{
			5.1, 5, 5.6, 5.5, 5.5, 5.8, 5.5, 5.5, 5.8, 5.6,
			5.7, 5, 5.6, 5.9, 6.2, 6, 4.9, 6.3, 6.1, 5.6,
			5.8, 6.7, 6.1, 5.9, 6, 4.9, 5.6, 5.2, 6.1, 6.4,
			7, 5.7, 6.5, 6.9, 5.7, 6.4, 6.2, 6.6, 6.3, 6.2,
			5.4, 6.7, 6.1, 5.7, 5.5, 6, 3, 6.6, 5.7, 6,
			6.8, 6, 6.1, 6.3, 5.8, 5.8, 5.6, 5.7, 6, 6.9,
			6.9, 6.4, 6.3, 6.3, 6.7, 6.5, 5.8, 6.3, 6.4, 6.7,
			5.9, 7.2, 6.3, 6.3, 6.5, 7.1, 6.7, 7.6, 7.3, 6.4,
			6.7, 7.4, 6, 6.8, 6.5, 6.4, 6.7, 6.4, 6.5, 6.9,
			7.7, 6.7, 7.2, 7.7, 7.2, 7.7, 6.1, 7.9, 7.7, 6.8,
			6.2,
		},
		exp: []float64{
			3, 4.9, 4.9, 5, 5, 5.1, 5.2, 5.4, 5.5, 5.5,
			5.5, 5.5, 5.5, 5.6, 5.6, 5.6, 5.6, 5.6, 5.6, 5.7,
			5.7, 5.7, 5.7, 5.7, 5.7, 5.8, 5.8, 5.8, 5.8, 5.8,
			5.8, 5.9, 5.9, 5.9, 6, 6, 6, 6, 6, 6,
			6, 6.1, 6.1, 6.1, 6.1, 6.1, 6.1, 6.2, 6.2, 6.2,
			6.2, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.4,
			6.4, 6.4, 6.4, 6.4, 6.4, 6.4, 6.5, 6.5, 6.5, 6.5,
			6.5, 6.6, 6.6, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7,
			6.7, 6.8, 6.8, 6.8, 6.9, 6.9, 6.9, 6.9, 7, 7.1,
			7.2, 7.2, 7.2, 7.3, 7.4, 7.6, 7.7, 7.7, 7.7, 7.7,
			7.9,
		},
		isAsc: true,
	}, {
		desc: `case 6 desc`,
		slice: []float64{
			5.1, 5, 5.6, 5.5, 5.5, 5.8, 5.5, 5.5, 5.8, 5.6,
			5.7, 5, 5.6, 5.9, 6.2, 6, 4.9, 6.3, 6.1, 5.6,
			5.8, 6.7, 6.1, 5.9, 6, 4.9, 5.6, 5.2, 6.1, 6.4,
			7, 5.7, 6.5, 6.9, 5.7, 6.4, 6.2, 6.6, 6.3, 6.2,
			5.4, 6.7, 6.1, 5.7, 5.5, 6, 3, 6.6, 5.7, 6,
			6.8, 6, 6.1, 6.3, 5.8, 5.8, 5.6, 5.7, 6, 6.9,
			6.9, 6.4, 6.3, 6.3, 6.7, 6.5, 5.8, 6.3, 6.4, 6.7,
			5.9, 7.2, 6.3, 6.3, 6.5, 7.1, 6.7, 7.6, 7.3, 6.4,
			6.7, 7.4, 6, 6.8, 6.5, 6.4, 6.7, 6.4, 6.5, 6.9,
			7.7, 6.7, 7.2, 7.7, 7.2, 7.7, 6.1, 7.9, 7.7, 6.8,
			6.2,
		},
		exp: []float64{
			7.9, 7.7, 7.7, 7.7, 7.7, 7.6, 7.4, 7.3, 7.2, 7.2,
			7.2, 7.1, 7, 6.9, 6.9, 6.9, 6.9, 6.8, 6.8, 6.8,
			6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.6, 6.6,
			6.5, 6.5, 6.5, 6.5, 6.5, 6.4, 6.4, 6.4, 6.4, 6.4,
			6.4, 6.4, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3,
			6.2, 6.2, 6.2, 6.2, 6.1, 6.1, 6.1, 6.1, 6.1, 6.1,
			6, 6, 6, 6, 6, 6, 6, 5.9, 5.9, 5.9,
			5.8, 5.8, 5.8, 5.8, 5.8, 5.8, 5.7, 5.7, 5.7, 5.7,
			5.7, 5.7, 5.6, 5.6, 5.6, 5.6, 5.6, 5.6, 5.5, 5.5,
			5.5, 5.5, 5.5, 5.4, 5.2, 5.1, 5, 5, 4.9, 4.9,
			3,
		},
		isAsc: false,
	}}

	for _, tcase := range listCase {
		var gotSortedIdx = slices.IndirectSort(tcase.slice, tcase.isAsc)
		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)

		if len(tcase.expSortedIdx) == 0 {
			continue
		}

		var expSortedIdxStr = fmt.Sprint(tcase.expSortedIdx)
		var gotSortedIdxStr = fmt.Sprint(gotSortedIdx)
		test.Assert(t, tcase.desc+` sortedIdx`, expSortedIdxStr,
			gotSortedIdxStr)
	}
}

func TestInplaceInsertionSort_float64(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []float64
		exp   []float64
		isAsc bool
	}{{
		desc: `case empty`,
	}, {
		desc:  `case 1 asc`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		exp:   []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9},
		isAsc: true,
	}, {
		desc:  `case 1 desc`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		exp:   []float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1, 0.0},
		isAsc: false,
	}, {
		desc:  `case 2 asc`,
		slice: []float64{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
		exp:   []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.1, 0.1, 0.1, 0.1},
		isAsc: true,
	}, {
		desc:  `case 2 desc`,
		slice: []float64{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
		exp:   []float64{0.1, 0.1, 0.1, 0.1, 0.0, 0.0, 0.0, 0.0, 0.0},
		isAsc: false,
	}, {
		desc:  `case 3 asc`,
		slice: []float64{1, 1, 2, 2, 3, 1, 2},
		exp:   []float64{1, 1, 1, 2, 2, 2, 3},
		isAsc: true,
	}, {
		desc:  `case 3 desc`,
		slice: []float64{1, 1, 2, 2, 3, 1, 2},
		exp:   []float64{3, 2, 2, 2, 1, 1, 1},
		isAsc: false,
	}}
	for _, tcase := range listCase {
		var ids = make([]int, len(tcase.slice))
		slices.InplaceInsertionSort(tcase.slice, ids, 0, len(ids),
			tcase.isAsc)
		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)
	}
}

func TestInplaceMergesort_float64(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []float64
		exp   []float64
	}{{
		desc: `case 1`,
		slice: []float64{
			5.1, 5, 5.6, 5.5, 5.5, 5.8, 5.5, 5.5, 5.8, 5.6,
			5.7, 5, 5.6, 5.9, 6.2, 6, 4.9, 6.3, 6.1, 5.6,
			5.8, 6.7, 6.1, 5.9, 6, 4.9, 5.6, 5.2, 6.1, 6.4,
			7, 5.7, 6.5, 6.9, 5.7, 6.4, 6.2, 6.6, 6.3, 6.2,
			5.4, 6.7, 6.1, 5.7, 5.5, 6, 3, 6.6, 5.7, 6,
			6.8, 6, 6.1, 6.3, 5.8, 5.8, 5.6, 5.7, 6, 6.9,
			6.9, 6.4, 6.3, 6.3, 6.7, 6.5, 5.8, 6.3, 6.4, 6.7,
			5.9, 7.2, 6.3, 6.3, 6.5, 7.1, 6.7, 7.6, 7.3, 6.4,
			6.7, 7.4, 6, 6.8, 6.5, 6.4, 6.7, 6.4, 6.5, 6.9,
			7.7, 6.7, 7.2, 7.7, 7.2, 7.7, 6.1, 7.9, 7.7, 6.8,
			6.2,
		},
		exp: []float64{
			3, 4.9, 4.9, 5, 5, 5.1, 5.2, 5.4, 5.5, 5.5,
			5.5, 5.5, 5.5, 5.6, 5.6, 5.6, 5.6, 5.6, 5.6, 5.7,
			5.7, 5.7, 5.7, 5.7, 5.7, 5.8, 5.8, 5.8, 5.8, 5.8,
			5.8, 5.9, 5.9, 5.9, 6, 6, 6, 6, 6, 6,
			6, 6.1, 6.1, 6.1, 6.1, 6.1, 6.1, 6.2, 6.2, 6.2,
			6.2, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.3, 6.4,
			6.4, 6.4, 6.4, 6.4, 6.4, 6.4, 6.5, 6.5, 6.5, 6.5,
			6.5, 6.6, 6.6, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7,
			6.7, 6.8, 6.8, 6.8, 6.9, 6.9, 6.9, 6.9, 7, 7.1,
			7.2, 7.2, 7.2, 7.3, 7.4, 7.6, 7.7, 7.7, 7.7, 7.7,
			7.9,
		},
	}}

	for _, tcase := range listCase {
		var size = len(tcase.slice)
		var idx = make([]int, size)

		slices.InplaceMergesort(tcase.slice, idx, 0, size, true)
		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)
	}
}

func TestMax_float64(t *testing.T) {
	var listCase = []struct {
		desc   string
		slice  []float64
		exp    float64
		expIdx int
	}{{
		expIdx: -1,
	}, {
		slice:  []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		exp:    0.9,
		expIdx: 4,
	}}

	for _, tcase := range listCase {
		gotMax, gotIdx := slices.Max2(tcase.slice)

		test.Assert(t, tcase.desc+` max`, tcase.exp, gotMax)
		test.Assert(t, tcase.desc+` idx`, tcase.expIdx, gotIdx)
	}
}

func TestMin_float64(t *testing.T) {
	var listCase = []struct {
		desc   string
		slice  []float64
		exp    float64
		expIdx int
	}{{
		expIdx: -1,
	}, {
		slice:  []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		exp:    0.0,
		expIdx: 5,
	}}

	for _, tcase := range listCase {
		gotMin, gotIdx := slices.Min2(tcase.slice)

		test.Assert(t, tcase.desc+` min`, tcase.exp, gotMin)
		test.Assert(t, tcase.desc+` idx`, tcase.expIdx, gotIdx)
	}
}

func TestMaxCountOf_float64(t *testing.T) {
	var listCase = []struct {
		desc    string
		slice   []float64
		classes []float64
		exp     float64
		state   int
	}{{
		desc:    `case 1`,
		slice:   []float64{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
		classes: []float64{0, 1},
		exp:     0,
		state:   0,
	}, {
		desc:    `case 2`,
		slice:   []float64{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
		classes: []float64{1, 0},
		exp:     0,
	}}

	for _, tcase := range listCase {
		got, gotState := slices.MaxCountOf(tcase.slice, tcase.classes)
		test.Assert(t, tcase.desc+` got`, tcase.exp, got)
		test.Assert(t, tcase.desc+` state`, tcase.state, gotState)
	}
}

func TestSortByIndex_float64(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []float64
		ids   []int
		exp   []float64
	}{{
		desc: `case empty`,
	}, {
		desc:  `case 1`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		ids:   []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:   []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9},
	}, {
		desc:  `case 2`,
		slice: []float64{0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0, 0.1, 0.0},
		ids:   []int{0, 2, 4, 6, 8, 1, 3, 5, 7},
		exp:   []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.1, 0.1, 0.1, 0.1},
	}, {
		desc:  `case 3`,
		slice: []float64{1, 1, 2, 2, 3, 1, 2},
		ids:   []int{0, 1, 5, 6, 2, 3, 4},
		exp:   []float64{1, 1, 1, 2, 2, 2, 3},
	}}

	for _, tcase := range listCase {
		slices.SortByIndex(&tcase.slice, tcase.ids)
		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)
	}
}

func TestSum_float64(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []float64
		exp   float64
	}{{
		desc:  `case 1`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		exp:   4.5,
	}}

	for _, tcase := range listCase {
		var got = slices.Sum(tcase.slice)
		got = numbers.Float64Round(got, 1)
		test.Assert(t, tcase.desc, tcase.exp, got)
	}
}

func TestSwap_float64(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []float64
		exp   []float64
		idx1  int
		idx2  int
	}{{
		desc: `empty`,
		idx1: 1,
		idx2: 6,
	}, {
		desc:  `equal indices`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		idx1:  1,
		idx2:  1,
		exp:   []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
	}, {
		desc:  `out of range indices`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		idx1:  1,
		idx2:  100,
		exp:   []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
	}, {
		desc:  `case 1`,
		slice: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.4},
		idx1:  0,
		idx2:  9,
		exp:   []float64{0.4, 0.6, 0.7, 0.8, 0.9, 0.0, 0.1, 0.2, 0.3, 0.5},
	}}

	for _, tcase := range listCase {
		slices.Swap(tcase.slice, tcase.idx1, tcase.idx2)
		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)
	}
}
