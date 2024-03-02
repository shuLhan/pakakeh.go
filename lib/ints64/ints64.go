// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ints64 provide a library for working with slice of 64 bit integer.
package ints64

import (
	"git.sr.ht/~shulhan/pakakeh.go/lib/ints"
)

// Count number of class in slice.
func Count(d []int64, class int64) (count int) {
	if len(d) == 0 {
		return
	}
	for x := 0; x < len(d); x++ {
		if d[x] == class {
			count++
		}
	}
	return
}

// Counts number of each class in slice.
//
// For example, if data is "[3,3,4]" and classes is "[3,4,5]", this function
// will return "[2,1,0]".
func Counts(d, classes []int64) (counts []int) {
	if len(classes) == 0 {
		return
	}
	counts = make([]int, len(classes))
	for x, c := range classes {
		counts[x] = Count(d, c)
	}
	return
}

// IndirectSort will sort the data and return the sorted index.
func IndirectSort(d []int64, asc bool) (sortedIdx []int) {
	if len(d) == 0 {
		return
	}

	sortedIdx = make([]int, len(d))
	for x := 0; x < len(d); x++ {
		sortedIdx[x] = x
	}

	InplaceMergesort(d, sortedIdx, 0, len(d), asc)

	return
}

// InplaceMergesort in-place merge-sort without memory allocation.
func InplaceMergesort(d []int64, idx []int, l, r int, asc bool) {
	// If length of data == Threshold, then use insertion sort.
	if l+7 >= r {
		InsertionSortWithIndices(d, idx, l, r, asc)
		return
	}

	// Divide into left and right.
	res := (r + l) % 2
	c := (r + l) / 2
	if res == 1 {
		c++
	}

	// Sort left.
	InplaceMergesort(d, idx, l, c, asc)

	// Sort right.
	InplaceMergesort(d, idx, c, r, asc)

	// Merge sorted left and right.
	if asc {
		if d[c-1] <= d[c] {
			// (4.1) If the last element of the left is lower then
			// the first element of the right, i.e. [1 2] [3 4];
			// no merging needed, return immediately.
			return
		}
	} else {
		if d[c-1] >= d[c] {
			return
		}
	}

	inplaceMerge(d, idx, l, c, r, asc)
}

// InsertionSortWithIndices will sort the data using insertion-sort algorithm.
//
// Parameters:
// `d` is slice that will be sorted,
// `ids` is indices of data "d",
// `l` is starting index of slice to be sorted, and
// `r` is end index of slice to be sorted.
func InsertionSortWithIndices(d []int64, ids []int, l, r int, asc bool) {
	for x := l; x < r; x++ {
		for y := x + 1; y < r; y++ {
			if asc {
				if d[x] > d[y] {
					ints.Swap(ids, x, y)
					Swap(d, x, y)
				}
			} else {
				if d[x] < d[y] {
					ints.Swap(ids, x, y)
					Swap(d, x, y)
				}
			}
		}
	}
}

// Max find the maximum value in slice and return its value and index.
//
// If slice is empty, it will return false in ok.
func Max(d []int64) (v int64, i int, ok bool) {
	if len(d) == 0 {
		return 0, 0, false
	}
	v = d[0]
	i = 0
	for x := 1; x < len(d); x++ {
		if d[x] > v {
			v = d[x]
			i = x
		}
	}
	return v, i, true
}

// IsExist will return true if value `v` exist in slice of `d`,
// otherwise it will return false.
func IsExist(d []int64, v int64) bool {
	for x := 0; x < len(d); x++ {
		if d[x] == v {
			return true
		}
	}
	return false
}

// MaxCountOf count number of occurrence of each element of classes
// in data and return the class with maximum count.
//
// If `classes` is empty, it will return -1 and false.
// If `data` is empty, it will return -2 and false.
// If two or more class has the same count value, then the first max in the
// class will be returned.
//
// For example, given a data [0, 1, 0, 1, 0] and classes [0, 1], the function
// will count 0 as 3, 1 as 2; and return (0, true).
func MaxCountOf(d, classes []int64) (int64, bool) {
	if len(classes) == 0 {
		return -1, false
	}
	if len(d) == 0 {
		return -2, false
	}

	counts := Counts(d, classes)

	_, i, _ := ints.Max(counts)

	return classes[i], true
}

// MaxRange find the (last) maximum value in range of slice between index "l"
// and "r".
//
// WARNING: Caller must check for out of range index of "l" or "r" before
// calling this function or it will be panic.
func MaxRange(d []int64, l, r int) (v int64, i int) {
	v = d[l]
	i = l
	for l++; l < r; l++ {
		if d[l] >= v {
			v = d[l]
			i = l
		}
	}
	return
}

// Min find the minimum value in slice and return its value and index.
//
// If slice is empty, it will return false in ok.
func Min(d []int64) (v int64, i int, ok bool) {
	if len(d) == 0 {
		return 0, 0, false
	}
	v = d[0]
	i = 0
	for x := 1; x < len(d); x++ {
		if d[x] < v {
			v = d[x]
			i = x
		}
	}
	return v, i, true
}

// MinRange find the (last) minimum value in range of slice between "l" to "r"
// index.
//
// WARNING: this function does not check if slice is empty or index value of
// "l" or "r" out of range, it other words you must check manually before
// calling this function of it will become panic.
func MinRange(d []int64, l, r int) (v int64, i int) {
	v = d[l]
	i = l
	for l++; l < r; l++ {
		if d[l] <= v {
			v = d[l]
			i = l
		}
	}
	return
}

// SortByIndex will sort the slice `d` using sorted index `sortedListID`.
func SortByIndex(d *[]int64, sortedListID []int) {
	newd := make([]int64, len(*d))

	for x := range sortedListID {
		newd[x] = (*d)[sortedListID[x]]
	}

	(*d) = newd
}

// Sum all value in slice.
func Sum(d []int64) (sum int64) {
	for x := 0; x < len(d); x++ {
		sum += d[x]
	}
	return
}

// Swap two indices value of slice.
func Swap(d []int64, x, y int) {
	if x == y || len(d) <= 1 || x > len(d) || y > len(d) {
		return
	}
	d[x], d[y] = d[y], d[x]
}

// Let `x` be the first index of left-side, `y` be the first index of
// the right-side, and `r` as length of slice `d`
func inplaceMerge(d []int64, idx []int, x, y, r int, asc bool) {
	var ylast int

	// (4.3) Loop until either x or y reached the maximum slice.
	for x < r && y < r {
		// (4.3.1) IF DATA[x] <= DATA[y]
		if asc {
			if d[x] <= d[y] {
				x++

				// (4.3.1.2) IF x > y THEN GOTO 4.3
				if x >= y {
					goto next
				}

				// (4.3.1.3) GOTO 4.3
				continue
			}
		} else {
			if d[x] >= d[y] {
				x++

				if x >= y {
					goto next
				}

				continue
			}
		}

		// (4.3.2) LET YLAST := the next DATA[y] that is less DATA[x]
		ylast = moveY(d, x, y, r, asc)

		// (4.3.3) SWAP DATA, X, Y, YLAST
		multiswap(d, idx, x, y, ylast)

	next:
		// (4.3.4) LET Y := the minimum value between x and r on `d`
		minY(d, &x, &y, r, asc)
	}
}

// (4.3.4) LET Y := the minimum value between x and r on `d`.
func minY(d []int64, x, y *int, r int, asc bool) {
	for *x < r {
		if asc {
			_, *y = MinRange(d, *x, r)
		} else {
			_, *y = MaxRange(d, *x, r)
		}

		if *y != *x {
			break
		}
		(*x)++
	}
}

func moveY(d []int64, x, y, r int, asc bool) int {
	yorg := y
	y++
	for y < r {
		if asc {
			if d[y] >= d[x] {
				break
			}
			if d[y] < d[yorg] {
				break
			}
		} else {
			if d[y] <= d[x] {
				break
			}
			if d[y] > d[yorg] {
				break
			}
		}
		y++
	}
	return y
}

func multiswap(d []int64, idx []int, x, y, ylast int) int {
	for y < ylast {
		ints.Swap(idx, x, y)
		Swap(d, x, y)
		x++
		y++
		if y >= ylast {
			return y
		}
		if d[x] <= d[y] {
			return y
		}
	}

	return y
}
