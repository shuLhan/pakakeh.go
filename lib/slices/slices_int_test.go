// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package slices_test

import (
	"crypto/rand"
	"log"
	"math"
	"math/big"
	"sort"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/slices"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestCount_int(t *testing.T) {
	var listCase = []struct {
		desc string
		list []int
		val  int
		exp  int
	}{{
		desc: `empty list`,
		list: []int{},
	}, {
		desc: `found`,
		list: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		val:  1,
		exp:  1,
	}, {
		desc: `not found`,
		list: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		val:  10,
	}}

	for _, tcase := range listCase {
		var got = slices.Count(tcase.list, tcase.val)
		test.Assert(t, tcase.desc, tcase.exp, got)
	}
}

func TestCounts_int(t *testing.T) {
	var listCase = []struct {
		desc    string
		list    []int
		classes []int
		exp     []int
	}{{
		desc:    `empty list`,
		classes: []int{1, 2, 3},
		exp:     []int{0, 0, 0},
	}, {
		desc: `empty classes`,
		list: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
	}, {
		desc:    `ok`,
		list:    []int{1, 1, 2, 2, 3, 1, 2},
		classes: []int{1, 2, 3},
		exp:     []int{3, 3, 1},
	}}

	for _, tcase := range listCase {
		var got = slices.Counts(tcase.list, tcase.classes)
		test.Assert(t, tcase.desc, tcase.exp, got)
	}
}

func TestIndirectSort_int(t *testing.T) {
	var listCase = []struct {
		desc         string
		slice        []int
		expSortedIdx []int
		exp          []int
		isAsc        bool
	}{{
		desc:         `case 1 asc`,
		slice:        []int{9, 8, 7, 6, 5, 4, 3},
		expSortedIdx: []int{6, 5, 4, 3, 2, 1, 0},
		exp:          []int{3, 4, 5, 6, 7, 8, 9},
		isAsc:        true,
	}, {
		desc:         `case 1 desc`,
		slice:        []int{9, 8, 7, 6, 5, 4, 3},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6},
		exp:          []int{9, 8, 7, 6, 5, 4, 3},
		isAsc:        false,
	}, {
		desc:         `case 2 asc`,
		slice:        []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		expSortedIdx: []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		exp:          []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc:        true,
	}, {
		desc:         `case 2 desc`,
		slice:        []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		exp:          []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		isAsc:        false,
	}, {
		desc:         `case 3 asc`,
		slice:        []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		exp:          []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc:        true,
	}, {
		desc:         `case 3 desc`,
		slice:        []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		expSortedIdx: []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		exp:          []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		isAsc:        false,
	}, {
		desc:         `case 4 asc`,
		slice:        []int{0, 6, 7, 8, 5, 1, 2, 3, 4, 9},
		expSortedIdx: []int{0, 5, 6, 7, 8, 4, 1, 2, 3, 9},
		exp:          []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc:        true,
	}, {
		desc:         `case 4 desc`,
		slice:        []int{0, 6, 7, 8, 5, 1, 2, 3, 4, 9},
		expSortedIdx: []int{9, 3, 2, 1, 4, 8, 7, 6, 5, 0},
		exp:          []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		isAsc:        false,
	}, {
		desc:         `stability asc`,
		slice:        []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		exp:          []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		isAsc:        true,
	}, {
		desc:         `stability desc`,
		slice:        []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		expSortedIdx: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		exp:          []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		isAsc:        false,
	}}

	for _, tcase := range listCase {
		var gotSortedIdx = slices.IndirectSort(tcase.slice,
			tcase.isAsc)
		test.Assert(t, tcase.desc+` sortedIdx`,
			tcase.expSortedIdx, gotSortedIdx)
		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)
	}
}

func TestIndirectSort_compareSortInts_int(t *testing.T) {
	const n = 10_000
	var slice1 = make([]int, n)
	generateRandomInts(slice1, n)

	var slice2 = make([]int, len(slice1))
	copy(slice2, slice1)

	sort.Ints(slice1)
	slices.IndirectSort(slice2, true)

	test.Assert(t, `vs sort.Ints`, slice1, slice2)
}

func TestInplaceInsertionSort_int(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []int
		exp   []int
		isAsc bool
	}{{
		desc: `empty`,
	}, {
		desc:  `case 1 asc`,
		slice: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		isAsc: true,
	}, {
		desc:  `case 1 desc`,
		slice: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:   []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		isAsc: false,
	}, {
		desc:  `case 2 asc`,
		slice: []int{0, 1, 0, 1, 0, 1, 0, 1, 0},
		exp:   []int{0, 0, 0, 0, 0, 1, 1, 1, 1},
		isAsc: true,
	}, {
		desc:  `case 2 desc`,
		slice: []int{0, 1, 0, 1, 0, 1, 0, 1, 0},
		exp:   []int{1, 1, 1, 1, 0, 0, 0, 0, 0},
		isAsc: false,
	}, {
		desc:  `case 3 asc`,
		slice: []int{1, 1, 2, 2, 3, 1, 2},
		exp:   []int{1, 1, 1, 2, 2, 2, 3},
		isAsc: true,
	}, {
		desc:  `case 3 desc`,
		slice: []int{1, 1, 2, 2, 3, 1, 2},
		exp:   []int{3, 2, 2, 2, 1, 1, 1},
	}}

	for _, tcase := range listCase {
		var idx = make([]int, len(tcase.slice))

		slices.InplaceInsertionSort(tcase.slice, idx,
			0, len(tcase.slice), tcase.isAsc)

		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)
	}
}

func TestInplaceMergesort_int(t *testing.T) {
	var list = []int{
		51, 50, 56, 55, 55, 58, 55, 55, 58, 56,
		57, 50, 56, 59, 62, 60, 49, 63, 61, 56,
		58, 67, 61, 59, 60, 49, 56, 52, 61, 64,
		70, 57, 65, 69, 57, 64, 62, 66, 63, 62,
		54, 67, 61, 57, 55, 60, 30, 66, 57, 60,
		68, 60, 61, 63, 58, 58, 56, 57, 60, 69,
		69, 64, 63, 63, 67, 65, 58, 63, 64, 67,
		59, 72, 63, 63, 65, 71, 67, 76, 73, 64,
		67, 74, 60, 68, 65, 64, 67, 64, 65, 69,
		77, 67, 72, 77, 72, 77, 61, 79, 77, 68,
		62,
	}
	var exp = []int{
		30, 49, 49, 50, 50, 51, 52, 54, 55, 55,
		55, 55, 55, 56, 56, 56, 56, 56, 56, 57,
		57, 57, 57, 57, 57, 58, 58, 58, 58, 58,
		58, 59, 59, 59, 60, 60, 60, 60, 60, 60,
		60, 61, 61, 61, 61, 61, 61, 62, 62, 62,
		62, 63, 63, 63, 63, 63, 63, 63, 63, 64,
		64, 64, 64, 64, 64, 64, 65, 65, 65, 65,
		65, 66, 66, 67, 67, 67, 67, 67, 67, 67,
		67, 68, 68, 68, 69, 69, 69, 69, 70, 71,
		72, 72, 72, 73, 74, 76, 77, 77, 77, 77,
		79,
	}

	var size = len(list)
	var idx = make([]int, size)

	slices.InplaceMergesort(list, idx, 0, size, true)

	test.Assert(t, `InplaceMergesort`, exp, list)
}

func TestMax_int(t *testing.T) {
	var listCase = []struct {
		desc   string
		slice  []int
		exp    int
		expIdx int
	}{{
		desc:   `empty slice`,
		expIdx: -1,
	}, {
		desc:   `case 1`,
		slice:  []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:    9,
		expIdx: 4,
	}}

	for _, tcase := range listCase {
		gotMax, gotIdx := slices.Max2(tcase.slice)
		test.Assert(t, tcase.desc+` max`, tcase.exp, gotMax)
		test.Assert(t, tcase.desc+` idx`, tcase.expIdx, gotIdx)
	}
}

func TestMaxCountOf_int(t *testing.T) {
	var listCase = []struct {
		desc     string
		slice    []int
		classes  []int
		exp      int
		expState int
	}{{
		desc:     `empty classes`,
		slice:    []int{1, 2, 1},
		expState: -1,
	}, {
		desc:     `empty slice`,
		classes:  []int{1, 2},
		expState: -2,
	}, {
		desc:     `case 1`,
		slice:    []int{0, 1, 0, 1, 0, 1, 0, 1, 0},
		classes:  []int{0, 1},
		exp:      0,
		expState: 0,
	}, {
		desc:     `case 2`,
		slice:    []int{0, 1, 0, 1, 0, 1, 0, 1, 0},
		classes:  []int{1, 0},
		exp:      0,
		expState: 0,
	}}

	for _, tcase := range listCase {
		got, gotState := slices.MaxCountOf(tcase.slice, tcase.classes)
		test.Assert(t, tcase.desc+` maxCount`, tcase.exp, got)
		test.Assert(t, tcase.desc+` state`, tcase.expState, gotState)
	}
}

func TestMin_int(t *testing.T) {
	var listCase = []struct {
		desc   string
		slice  []int
		exp    int
		expIdx int
	}{{
		desc:   `empty`,
		expIdx: -1,
	}, {
		desc:   `case 1`,
		slice:  []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:    0,
		expIdx: 5,
	}}

	for _, tcase := range listCase {
		gotMin, gotIdx := slices.Min2(tcase.slice)
		test.Assert(t, tcase.desc, tcase.exp, gotMin)
		test.Assert(t, tcase.desc, tcase.expIdx, gotIdx)
	}
}

func TestRemove_int(t *testing.T) {
	var listCase = []struct {
		slice []int
		exp   []int
		val   int
		expOK bool
	}{{
		slice: []int{},
		val:   1,
		exp:   []int{},
	}, {
		slice: []int{1},
		val:   1,
		exp:   []int{},
		expOK: true,
	}, {
		slice: []int{1, 2, 3, 4},
		val:   5,
		exp:   []int{1, 2, 3, 4},
	}, {
		slice: []int{1, 2, 3, 4},
		val:   1,
		exp:   []int{2, 3, 4},
		expOK: true,
	}}

	var got []int
	var ok bool
	for _, tcase := range listCase {
		got, ok = slices.Remove(tcase.slice, tcase.val)
		test.Assert(t, `ok`, tcase.expOK, ok)
		test.Assert(t, `result`, tcase.exp, got)
	}
}

func TestSortByIndex_int(t *testing.T) {
	var listCase = []struct {
		desc      string
		slice     []int
		sortedIdx []int
		exp       []int
	}{{
		desc: `empty slice`,
	}, {
		desc:      `case 1`,
		slice:     []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		sortedIdx: []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		exp:       []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}, {
		desc:      `case 2`,
		slice:     []int{0, 1, 0, 1, 0, 1, 0, 1, 0},
		sortedIdx: []int{0, 2, 4, 6, 8, 1, 3, 5, 7},
		exp:       []int{0, 0, 0, 0, 0, 1, 1, 1, 1},
	}, {
		desc:      `case 3`,
		slice:     []int{1, 1, 2, 2, 3, 1, 2},
		sortedIdx: []int{0, 1, 5, 6, 2, 3, 4},
		exp:       []int{1, 1, 1, 2, 2, 2, 3},
	}}

	for _, tcase := range listCase {
		slices.SortByIndex(&tcase.slice, tcase.sortedIdx)
		test.Assert(t, tcase.desc, tcase.exp, tcase.slice)
	}
}

func TestSum_int(t *testing.T) {
	var listCase = []struct {
		desc  string
		slice []int
		exp   int
	}{{
		desc:  `case 1`,
		slice: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:   45,
	}}

	for _, tcase := range listCase {
		var got = slices.Sum(tcase.slice)
		test.Assert(t, tcase.desc, tcase.exp, got)
	}
}

func TestSwap_int(t *testing.T) {
	var listCase = []struct {
		desc string
		list []int
		exp  []int
		idx1 int
		idx2 int
	}{{
		desc: `empty list`,
		idx1: 1,
		idx2: 6,
	}, {
		desc: `equal index`,
		list: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:  []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		idx1: 1,
		idx2: 1,
	}, {
		desc: `index 2 out of range`,
		list: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:  []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		idx1: 1,
		idx2: 100,
	}, {
		desc: `ok`,
		list: []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		exp:  []int{4, 6, 7, 8, 9, 0, 1, 2, 3, 5},
		idx1: 0,
		idx2: 9,
	}}

	for _, tcase := range listCase {
		slices.Swap(tcase.list, tcase.idx1, tcase.idx2)
		test.Assert(t, tcase.desc, tcase.exp, tcase.list)
	}
}

func generateRandomInts(data []int, n int) {
	var (
		max   = big.NewInt(math.MaxInt)
		randv *big.Int
		err   error
	)
	for x := 0; x < n; x++ {
		randv, err = rand.Int(rand.Reader, max)
		if err != nil {
			log.Fatalf(`generateRandomInts: %s`, err)
		}
		data[x] = int(randv.Int64())
	}
}
