// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package floats64 provide a library for working with slice of 64 bit float.
package floats64

import (
	"github.com/shuLhan/share/lib/ints"
)

// Count number of class in data.
func Count(d []float64, class float64) (count int) {
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

// Counts count class in data and return each of the counter.
//
// For example, if data is "[1,1,2]" and class is "[1,2]", this function will
// return "[2,1]".
func Counts(d, classes []float64) (counts []int) {
	if len(classes) == 0 {
		return
	}

	counts = make([]int, len(classes))

	for x, c := range classes {
		counts[x] = Count(d, c)
	}
	return
}

// IndirectSort sort the data and return the sorted index.
func IndirectSort(d []float64, asc bool) (sortedIdx []int) {
	dlen := len(d)

	sortedIdx = make([]int, dlen)
	for i := 0; i < dlen; i++ {
		sortedIdx[i] = i
	}

	InplaceMergesort(d, sortedIdx, 0, dlen, asc)

	return
}

// InplaceMergesort in-place merge-sort without memory allocation.
func InplaceMergesort(d []float64, idx []int, l, r int, asc bool) {
	// (0) If data length == Threshold, then
	if l+7 >= r {
		// (0.1) use insertion sort.
		InplaceInsertionSort(d, idx, l, r, asc)
		return
	}

	// (1) Divide into left and right.
	res := (r + l) % 2
	c := (r + l) / 2
	if res == 1 {
		c++
	}

	// (2) Sort left.
	InplaceMergesort(d, idx, l, c, asc)

	// (3) Sort right.
	InplaceMergesort(d, idx, c, r, asc)

	// (4) Merge sorted left and right.
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

// InplaceInsertionSort will sort the data using insertion-sort algorithm.
//
// Parameters:
// `data` is slice that will be sorted,
// `idx` is indices of data,
// `l` is starting index of slice to be sorted, and
// `r` is end index of slice to be sorted.
func InplaceInsertionSort(d []float64, ids []int, l, r int, asc bool) {
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

// IsExist will return true if value `v` exist in slice of `d`,
// otherwise it will return false.
func IsExist(d []float64, v float64) bool {
	for _, x := range d {
		if v == x {
			return true
		}
	}
	return false
}

// Max find the maximum value in slice and and return its value and index.
//
// If data is empty, it will return false in ok.
//
// Example, given data: [0.0 0.1 0.2 0.2 0.4], it will return 0.4 as max and 4
// as index of maximum value.
func Max(d []float64) (v float64, i int, ok bool) {
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

// MaxRange find the (last) maximum value in slice between index "l" and "r".
//
// WARNING: This function does not check index out of range.
func MaxRange(d []float64, l, r int) (m int) {
	maxv := d[l]
	m = l
	for l++; l < r; l++ {
		if d[l] >= maxv {
			maxv = d[l]
			m = l
		}
	}
	return m
}

// MaxCountOf will count number of occurrence of each element of
// classes in data and return the class with maximum count.
//
// If `classes` is empty, it will return -1 and false.
// If `data` is empty, it will return -2 and false.
// If classes has the same count value, then the first max in the class will be
// returned.
//
// For example, given a data [5, 6, 5, 6, 5] and classes [5, 6, 7], the
// function will count 5 as 3, 6 as 2, and 7 as 0.
// Since frequency of 5 is greater than 6 and 7, then it will return `5` and
// `true`.
func MaxCountOf(d, classes []float64) (float64, bool) {
	if len(classes) == 0 {
		return -1, false
	}
	if len(d) == 0 {
		return -2, false
	}

	counts := Counts(d, classes)

	_, maxi, _ := ints.Max(counts)

	return classes[maxi], true
}

// Min find the minimum value in slice and and return it with their index.
//
// If data is empty, return 0 in value and index, and false in ok.
//
// Example, given data: [0.0 0.1 0.2 0.2 0.4], it will return 0 as min and 0
// as index of minimum value.
func Min(d []float64) (v float64, i int, ok bool) {
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

// MinRange find the (last) minimum value in slice between index "l" and "r".
//
// WARNING: This function does not check index out of range.
func MinRange(d []float64, l, r int) (m int) {
	min := d[l]
	m = l
	for l++; l < r; l++ {
		if d[l] <= min {
			min = d[l]
			m = l
		}
	}
	return
}

// SortByIndex sort the slice of float using sorted index.
func SortByIndex(d *[]float64, sortedListID []int) {
	newd := make([]float64, len(*d))

	for i := range sortedListID {
		newd[i] = (*d)[sortedListID[i]]
	}

	(*d) = newd
}

// Sum value of slice.
func Sum(d []float64) (sum float64) {
	for x := 0; x < len(d); x++ {
		sum += d[x]
	}
	return
}

// Swap swap two indices value of 64bit float.
func Swap(d []float64, x, y int) {
	if x == y || len(d) <= 1 || x > len(d) || y > len(d) {
		return
	}
	tmp := d[x]
	d[x] = d[y]
	d[y] = tmp
}

// Let `x` be the first index of left-side, `y` be the first index of
// the right-side, and `r` as length of slice `d`
func inplaceMerge(d []float64, idx []int, x, y, r int, asc bool) {
	var ylast int

	// (4.3) Loop until either x or y reached the maximum slice.
	for x < r && y < r {
		// (4.3.1) IF DATA[x] <= DATA[y]
		if asc {
			if d[x] <= d[y] {
				x++

				// (4.3.1.2) IF x >= y THEN GOTO 4.3.4
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
func minY(d []float64, x, y *int, r int, asc bool) {
	for *x < r {
		if asc {
			*y = MinRange(d, *x, r)
		} else {
			*y = MaxRange(d, *x, r)
		}

		if *y != *x {
			break
		}
		(*x)++
	}
}

func moveY(d []float64, x, y, r int, asc bool) int {
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

func multiswap(d []float64, idx []int, x, y, ylast int) int {
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
