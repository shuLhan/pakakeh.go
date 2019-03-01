// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Packages ints provide a library for working with slice of integer.
package ints

//
// Count number of class in data.
//
func Count(d []int, class int) (count int) {
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

//
// Counts number of each class in slice.
//
// For example, if data is "[1,1,2]" and classes is "[1,2]", this function
// will return "[2,1]".
//
func Counts(d, classes []int) (counts []int) {
	if len(classes) == 0 {
		return
	}
	counts = make([]int, len(classes))
	for x, c := range classes {
		counts[x] = Count(d, c)
	}
	return
}

//
// IndirectSort sort the data and return the sorted index.
//
func IndirectSort(d []int, asc bool) (sortedIdx []int) {
	sortedIdx = make([]int, len(d))
	for i := 0; i < len(d); i++ {
		sortedIdx[i] = i
	}

	InplaceMergesort(d, sortedIdx, 0, len(d), asc)

	return
}

//
// InplaceInsertionSort will sort the data and their index using
// insertion-sort algorithm.
//
// Parameters:
// `d` is slice that will be sorted,
// `ids` is indices of data,
// `l` is starting index of slice to be sorted, and
// `r` is end index of slice to be sorted.
//
func InplaceInsertionSort(d, ids []int, l, r int, asc bool) {
	for x := l; x < r; x++ {
		for y := x + 1; y < r; y++ {
			if asc {
				if d[x] > d[y] {
					Swap(ids, x, y)
					Swap(d, x, y)
				}
			} else {
				if d[x] < d[y] {
					Swap(ids, x, y)
					Swap(d, x, y)
				}
			}
		}
	}
}

//
// InplaceMergesort sort the slice "d" in-place, without memory allocation,
// using mergesort algorithm.
//
func InplaceMergesort(d []int, idx []int, l, r int, asc bool) {
	if l+7 >= r {
		// If data length <= Threshold, then use insertion sort.
		InplaceInsertionSort(d, idx, l, r, asc)
		return
	}

	// Divide into left and right.
	c := ((r + l) / 2) + (r+l)%2

	// Sort left.
	InplaceMergesort(d, idx, l, c, asc)

	// Sort right.
	InplaceMergesort(d, idx, c, r, asc)

	// Merge sorted left and right.
	if asc {
		if d[c-1] <= d[c] {
			// If the last element of the left is lower then
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

//
// IsExist will return true if value `v` exist in slice of `d`,
// otherwise it will return false.
//
func IsExist(d []int, v int) bool {
	for x := 0; x < len(d); x++ {
		if d[x] == v {
			return true
		}
	}
	return false
}

//
// Max find the maximum value in slice and return its value and index.
//
// If slice is empty, it will return false in ok.
//
func Max(d []int) (v int, i int, ok bool) {
	if len(d) == 0 {
		return
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

//
// MaxCountOf count number of occurrence of each element of classes
// in data and return the class with maximum count.
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
//
func MaxCountOf(d, classes []int) (int, bool) {
	if len(classes) == 0 {
		return -1, false
	}
	if len(d) == 0 {
		return -2, false
	}

	counts := Counts(d, classes)

	_, i, _ := Max(counts)

	return classes[i], true
}

//
// MaxRange find the (last) maximum value in slice between index "l" and "r".
//
// WARNING: This function does not check index out of range.
//
func MaxRange(d []int, l, r int) (v, i int) {
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

//
// Min find the minimum value in slice and return its value and index.
//
// If slice is empty, it will return false in ok.
//
func Min(d []int) (v int, i int, ok bool) {
	if len(d) == 0 {
		return
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

//
// MinRange find the (last) minimum value in slice between index "l" and "r".
//
// WARNING: This function does not check index out of range.
//
func MinRange(d []int, l, r int) (v, i int) {
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

//
// Remove value "v" from slice if its exist and return new slice and true;
// otherwise, if not found, return unmodified slice and false.
//
func Remove(d []int, v int) ([]int, bool) {
	for x := 0; x < len(d); x++ {
		if d[x] == v {
			d = append(d[:x], d[x+1:]...)
			return d, true
		}
	}
	return d, false
}

//
// SortByIndex will sort the slice `d` using sorted index `sortedIds`.
//
func SortByIndex(d *[]int, sortedIds []int) {
	newd := make([]int, len(*d))

	for x := 0; x < len(sortedIds); x++ {
		newd[x] = (*d)[sortedIds[x]]
	}

	(*d) = newd
}

//
// Sum all value in slice.
//
func Sum(d []int) (sum int) {
	for x := 0; x < len(d); x++ {
		sum += d[x]
	}
	return
}

//
// Swap two indices value of slice.
//
func Swap(d []int, x, y int) {
	if x == y || len(d) <= 1 || x > len(d) || y > len(d) {
		return
	}
	tmp := d[x]
	d[x] = d[y]
	d[y] = tmp
}

//
// To64 convert slice of integer to 64 bit values.
//
func To64(ints []int) []int64 {
	i64 := make([]int64, len(ints))
	for x, v := range ints {
		i64[x] = int64(v)
	}
	return i64
}

//
// Let `x` be the first index of left-side, `y` be the first index of
// the right-side, and `r` as length of slice `d`
//
func inplaceMerge(d []int, idx []int, x, y, r int, asc bool) {
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
		inplaceMultiswap(d, idx, x, y, ylast)

	next:
		// (4.3.4) LET Y := the minimum value between x and r on `d`
		minY(d, &x, &y, r, asc)
	}
}

func inplaceMultiswap(d []int, idx []int, x, y, ylast int) int {
	for y < ylast {
		Swap(idx, x, y)
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

// (4.3.4) LET Y := the minimum value between x and r on `d`
func minY(d []int, x, y *int, r int, asc bool) {
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

func moveY(d []int, x, y, r int, asc bool) int {
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
