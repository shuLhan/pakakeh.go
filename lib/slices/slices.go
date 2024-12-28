// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package slices complement the standard [slices] package for working with
// slices comparable and [cmp.Ordered] types.
//
// List of current features,
//
//   - sort slice of [cmp.Ordered] types using in-place mergesort algorithm.
//   - sort slice of [cmp.Ordered] types by predefined index.
//   - count number of value occurrence in slice of [comparable] types.
//   - find minimum or maximum value in slice of [cmp.Ordered] types.
//   - sum slice of [constraints.Integer] or [constraints.Float] types.
package slices

import (
	"cmp"

	"golang.org/x/exp/constraints"
)

// sortThreshold when the data less than sortThreshold, insertion sort
// will be used to replace mergesort.
const sortThreshold = 7

// Count the number of occurence of val in slice.
func Count[S []E, E comparable](slice S, val E) (count int) {
	if len(slice) == 0 {
		return 0
	}
	for _, v := range slice {
		if v == val {
			count++
		}
	}
	return count
}

// Counts number of each class in slice.
//
// For example, if slice is "[1,1,2]" and classes is "[1,2]", this function
// will return "[2,1]".
func Counts[S []E, E comparable](slice S, classes S) (counts []int) {
	if len(classes) == 0 {
		return nil
	}
	counts = make([]int, len(classes))
	for x, val := range classes {
		counts[x] = Count(slice, val)
	}
	return counts
}

// IndirectSort sort the data and return the sorted index.
func IndirectSort[S []E, E cmp.Ordered](slice S, isAsc bool) (
	sortedIdx []int,
) {
	sortedIdx = make([]int, len(slice))
	for x := range len(slice) {
		sortedIdx[x] = x
	}

	InplaceMergesort(slice, sortedIdx, 0, len(slice), isAsc)

	return sortedIdx
}

// InplaceInsertionSort sort the data and their index using insertion-sort
// algorithm.
//
// The parameter
// `slice` is the slice that will be sorted,
// `ids` is indices of slice,
// `l` is starting index of slice to be sorted, and
// `r` is the end index of slice to be sorted.
func InplaceInsertionSort[S []E, E cmp.Ordered](
	slice S, ids []int, l, r int, isAsc bool,
) {
	var x int
	var y int
	for x = l; x < r; x++ {
		for y = x + 1; y < r; y++ {
			if isAsc {
				if slice[x] > slice[y] {
					Swap(ids, x, y)
					Swap(slice, x, y)
				}
			} else {
				if slice[x] < slice[y] {
					Swap(ids, x, y)
					Swap(slice, x, y)
				}
			}
		}
	}
}

// InplaceMergesort sort the slice "slice" in-place, without memory
// allocation, using mergesort algorithm.
// The ids parameter is empty slice with length equal to the length of slice,
// which will used for storing sorted index.
func InplaceMergesort[S []E, E cmp.Ordered](
	slice S, ids []int, l, r int, isAsc bool,
) {
	if r-l <= sortThreshold {
		// If data length <= Threshold, then use insertion sort.
		InplaceInsertionSort(slice, ids, l, r, isAsc)
		return
	}

	// Divide into left and right.
	var c = ((r + l) / 2) + (r+l)%2

	// Sort left.
	InplaceMergesort(slice, ids, l, c, isAsc)

	// Sort right.
	InplaceMergesort(slice, ids, c, r, isAsc)

	// Merge sorted left and right.
	if isAsc {
		if slice[c-1] <= slice[c] {
			// If the last element of the left is lower then
			// the first element of the right, i.e. [1 2] [3 4];
			// no merging needed, return immediately.
			return
		}
	} else {
		if slice[c-1] >= slice[c] {
			return
		}
	}

	inplaceMerge(slice, ids, l, c, r, isAsc)
}

// Max2 find the maximum value in slice and return its value and index.
// If slice is empty, it will return (0, -1).
func Max2[S []E, E cmp.Ordered](slice S) (max E, idx int) {
	if len(slice) == 0 {
		return max, -1
	}
	max = slice[0]
	idx = 0
	for x := 1; x < len(slice); x++ {
		if slice[x] > max {
			max = slice[x]
			idx = x
		}
	}
	return max, idx
}

// MaxCountOf count number of occurrence of each element of classes
// in data and return the class with maximum count.
//
// If classes has the same count value, then the first max in the class will
// be returned.
//
// For example, given a data [5, 6, 5, 6, 5] and classes [5, 6, 7], the
// function will count 5 as 3, 6 as 2, and 7 as 0.
// Since frequency of 5 is greater than 6 and 7, then it will return `5` and
// true.
func MaxCountOf[S []E, E comparable](slice S, classes S) (maxClass E, state int) {
	if len(classes) == 0 {
		return maxClass, -1
	}
	if len(slice) == 0 {
		return maxClass, -2
	}

	var counts []int = Counts(slice, classes)

	_, idx := Max2(counts)

	return classes[idx], 0
}

// MaxRange find the (last) maximum value in slice between index "l" and "r".
//
// WARNING: This function does not check index out of range.
func MaxRange[S []E, E cmp.Ordered](slice S, l, r int) (v E, i int) {
	v = slice[l]
	i = l
	for l++; l < r; l++ {
		if slice[l] >= v {
			v = slice[l]
			i = l
		}
	}
	return
}

// MergeByDistance merge two slice of integers by their distance between each
// others.
//
// For example, if slice a contains "{1, 5, 9}" and b contains
// "{4, 11, 15}" and the distance is 3, the output of merged is
// "{1, 5, 9, 15}".  The 4 and 11 are not included because 4 is in
// range between 1 and (1+3), and 11 is in range between 9 and 9+3.
func MergeByDistance[S []E, E ~int | ~int8 | ~int16 | ~int32 | ~int64](
	a, b S, distance E,
) (out S) {
	var lenab = len(a) + len(b)
	if lenab == 0 {
		return nil
	}

	var ab = make(S, 0, lenab)
	ab = append(ab, a...)
	ab = append(ab, b...)

	var idx = make([]int, lenab)
	InplaceMergesort(ab, idx, 0, lenab, true)

	out = append(out, ab[0])
	var last E = ab[0]
	for x := 1; x < len(ab); x++ {
		if ab[x] > last+distance {
			out = append(out, ab[x])
			last = ab[x]
			continue
		}
	}

	return out
}

// Min2 find the minimum value in slice and return its value and index.
//
// If slice is empty, it will return (0, -1).
func Min2[S []E, E cmp.Ordered](slice S) (min E, idx int) {
	if len(slice) == 0 {
		return min, -1
	}
	min = slice[0]
	idx = 0
	for x := 1; x < len(slice); x++ {
		if slice[x] < min {
			min = slice[x]
			idx = x
		}
	}
	return min, idx
}

// MinRange find the (last) minimum value in slice between index "l" and "r".
//
// WARNING: This function does not check index out of range.
func MinRange[S []E, E cmp.Ordered](slice S, l, r int) (v E, i int) {
	v = slice[l]
	i = l
	for l++; l < r; l++ {
		if slice[l] <= v {
			v = slice[l]
			i = l
		}
	}
	return
}

// Remove val from slice if its exist and return new slice and true.
// Otherwise, if val not found, return unmodified slice and false.
func Remove[S []E, E comparable](slice S, v E) (S, bool) {
	for x := 0; x < len(slice); x++ {
		if slice[x] == v {
			slice = append(slice[:x], slice[x+1:]...)
			return slice, true
		}
	}
	return slice, false
}

// SortByIndex sort the slice using sorted index sortedIdx.
func SortByIndex[S []E, E cmp.Ordered](slice *S, sortedIdx []int) {
	var newSlice = make([]E, len(*slice))

	for x := range len(sortedIdx) {
		newSlice[x] = (*slice)[sortedIdx[x]]
	}

	(*slice) = newSlice
}

// Sum all value in slice.
func Sum[S []E, E constraints.Integer | constraints.Float](slice S) (sum E) {
	for x := 0; x < len(slice); x++ {
		sum += slice[x]
	}
	return sum
}

// Swap two indices value of slice.
func Swap[S []E, E any](slice S, x, y int) {
	if x == y || len(slice) <= 1 || x > len(slice) || y > len(slice) {
		return
	}
	slice[x], slice[y] = slice[y], slice[x]
}

// ToInt64 convert slice of integer to its 64 bit values.
func ToInt64[S []E, E constraints.Integer](ints S) []int64 {
	i64 := make([]int64, len(ints))
	for x, v := range ints {
		i64[x] = int64(v)
	}
	return i64
}

// Let `x` be the first index of left-side, `y` be the first index of
// the right-side, and `r` as length of slice `slice`
func inplaceMerge[S []E, E cmp.Ordered](
	slice S, ids []int, x, y, r int, isAsc bool,
) {
	var ylast int

	// (4.3) Loop until either x or y reached the maximum slice.
	for x < r && y < r {
		// (4.3.1) IF DATA[x] <= DATA[y]
		if isAsc {
			if slice[x] <= slice[y] {
				x++

				// (4.3.1.2) IF x > y THEN GOTO 4.3
				if x >= y {
					goto next
				}

				// (4.3.1.3) GOTO 4.3
				continue
			}
		} else {
			if slice[x] >= slice[y] {
				x++

				if x >= y {
					goto next
				}

				continue
			}
		}

		// (4.3.2) LET YLAST := the next DATA[y] that is less DATA[x]
		ylast = moveY(slice, x, y, r, isAsc)

		// (4.3.3) SWAP DATA, X, Y, YLAST
		inplaceMultiswap(slice, ids, x, y, ylast)

	next:
		// (4.3.4) LET Y := the minimum value between x and r on
		// `slice`.
		minY(slice, &x, &y, r, isAsc)
	}
}

func inplaceMultiswap[S []E, E cmp.Ordered](
	slice S, ids []int, x, y, ylast int,
) int {
	for y < ylast {
		Swap(ids, x, y)
		Swap(slice, x, y)
		x++
		y++
		if y >= ylast {
			return y
		}
		if slice[x] <= slice[y] {
			return y
		}
	}

	return y
}

// (4.3.4) LET Y := the minimum value between x and r on slice.
func minY[S []E, E cmp.Ordered](slice S, x, y *int, r int, isAsc bool) {
	for *x < r {
		if isAsc {
			_, *y = MinRange(slice, *x, r)
		} else {
			_, *y = MaxRange(slice, *x, r)
		}

		if *y != *x {
			break
		}
		(*x)++
	}
}

func moveY[S []E, E cmp.Ordered](slice S, x, y, r int, isAsc bool) int {
	yorg := y
	y++
	for y < r {
		if isAsc {
			if slice[y] >= slice[x] {
				break
			}
			if slice[y] < slice[yorg] {
				break
			}
		} else {
			if slice[y] <= slice[x] {
				break
			}
			if slice[y] > slice[yorg] {
				break
			}
		}
		y++
	}
	return y
}
