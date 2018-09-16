// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

//
// IntsFindMax given slice of integer, return the maximum value in slice and
// its index.
//
// If data is empty, it will return `-1` in value and index, and false in ok.
//
func IntsFindMax(d []int) (maxv int, maxi int, ok bool) {
	l := len(d)
	if l <= 0 {
		return -1, -1, false
	}

	x := 0
	maxv = d[x]
	maxi = x

	for x = 1; x < l; x++ {
		if d[x] > maxv {
			maxv = d[x]
			maxi = x
		}
	}

	return maxv, maxi, true
}

//
// IntsFindMin given slice of integer, return the minimum value in slice and
// its index.
//
// If data is empty, return -1 in value and index, and false in ok.
//
// Example, given a slice of data: [0 1 2 3 4], it will return 0 as min and 0
// as minimum index.
//
func IntsFindMin(d []int) (minv int, mini int, ok bool) {
	l := len(d)
	if l <= 0 {
		return -1, -1, false
	}

	x := 0
	minv = d[x]
	mini = x

	for x = 1; x < l; x++ {
		if d[x] < minv {
			minv = d[x]
			mini = x
		}
	}

	return minv, mini, true
}

//
// IntsSum return sum of all value in slice.
//
func IntsSum(d []int) (sum int) {
	for _, v := range d {
		sum += v
	}
	return
}

//
// IntsCount will count number of class in data.
//
func IntsCount(d []int, class int) (count int) {
	if len(d) <= 0 {
		return
	}

	for _, v := range d {
		if v == class {
			count++
		}
	}
	return count
}

//
// IntsCounts will count class in data and return each of the counter.
//
// For example, if data is "[1,1,2]" and class is "[1,2]", this function will
// return "[2,1]".
//
//	idx class  count
//	0 : 1   -> 2
//	1 : 2   -> 1
//
func IntsCounts(d, classes []int) (counts []int) {
	if len(classes) <= 0 {
		return
	}

	counts = make([]int, len(classes))

	for x, c := range classes {
		counts[x] = IntsCount(d, c)
	}
	return
}

//
// IntsMaxCountOf will count number of occurence of each element of classes
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
func IntsMaxCountOf(d, classes []int) (int, bool) {
	if len(classes) == 0 {
		return -1, false
	}
	if len(d) == 0 {
		return -2, false
	}

	counts := IntsCounts(d, classes)

	_, maxi, _ := IntsFindMax(counts)
	if maxi < 0 {
		return -1, false
	}

	return classes[maxi], true
}

//
// IntsSwap swap two indices value of integer.
//
func IntsSwap(d []int, x, y int) {
	if x == y {
		return
	}
	if len(d) <= 1 || x > len(d) || y > len(d) {
		return
	}

	tmp := d[x]
	d[x] = d[y]
	d[y] = tmp
}

//
// IntsIsExist will return true if value `v` exist in slice of `d`,
// otherwise it will return false.
//
func IntsIsExist(d []int, i int) bool {
	for _, v := range d {
		if i == v {
			return true
		}
	}
	return false
}

//
// IntsTo64 convert slice of integer to 64bit values.
//
func IntsTo64(ints []int) []int64 {
	i64 := make([]int64, len(ints))
	for x, v := range ints {
		i64[x] = int64(v)
	}
	return i64
}

//
// IntsInsertionSort will sort the data using insertion-sort algorithm.
//
// Parameters:
// - `data` is slice that will be sorted.
// - `idx` is indices of data.
// - `l` is starting index of slice to be sorted.
// - `r` is end index of slice to be sorted.
//
func IntsInsertionSort(d, ids []int, l, r int, asc bool) {
	for x := l; x < r; x++ {
		for y := x + 1; y < r; y++ {
			if asc {
				if d[x] > d[y] {
					IntsSwap(ids, x, y)
					IntsSwap(d, x, y)
				}
			} else {
				if d[x] < d[y] {
					IntsSwap(ids, x, y)
					IntsSwap(d, x, y)
				}
			}
		}
	}
}

//
// IntsSortByIndex will sort the slice `d` using sorted index `sortedIds`.
//
func IntsSortByIndex(d *[]int, sortedIds []int) {
	newd := make([]int, len(*d))

	for i := range sortedIds {
		newd[i] = (*d)[sortedIds[i]]
	}

	(*d) = newd
}

//
// IntsInplaceMergesort in-place merge-sort without memory allocation.
//
func IntsInplaceMergesort(d []int, idx []int, l, r int, asc bool) {
	// (0) If data length == Threshold, then
	if l+SortThreshold >= r {
		// (0.1) use insertion sort.
		IntsInsertionSort(d, idx, l, r, asc)
		return
	}

	// (1) Divide into left and right.
	res := (r + l) % 2
	c := (r + l) / 2
	if res == 1 {
		c++
	}

	// (2) Sort left.
	IntsInplaceMergesort(d, idx, l, c, asc)

	// (3) Sort right.
	IntsInplaceMergesort(d, idx, c, r, asc)

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

	intsInplaceMerge(d, idx, l, c, r, asc)
}

//
// Let `x` be the first index of left-side, `y` be the first index of
// the right-side, and `r` as length of slice `d`
//
func intsInplaceMerge(d []int, idx []int, x, y, r int, asc bool) {
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
		ylast = intsMoveY(d, x, y, r, asc)

		// (4.3.3) SWAP DATA, X, Y, YLAST
		intsMultiswap(d, idx, x, y, ylast)

	next:
		// (4.3.4) LET Y := the minimum value between x and r on `d`
		intsMinY(d, &x, &y, r, asc)
	}
}

// (4.3.4) LET Y := the minimum value between x and r on `d`
func intsMinY(d []int, x, y *int, r int, asc bool) {
	for *x < r {
		if asc {
			*y = intsMin(d, *x, r)
		} else {
			*y = intsMax(d, *x, r)
		}

		if *y != *x {
			break
		}
		(*x)++
	}
}

func intsMoveY(d []int, x, y, r int, asc bool) int {
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

func intsMultiswap(d []int, idx []int, x, y, ylast int) int {
	for y < ylast {
		IntsSwap(idx, x, y)
		IntsSwap(d, x, y)
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

func intsMin(d []int, l, r int) (m int) {
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

func intsMax(d []int, l, r int) (m int) {
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

//
// IntsIndirectSort will sort the data and return the sorted index.
//
func IntsIndirectSort(d []int, asc bool) (sortedIdx []int) {
	dlen := len(d)

	sortedIdx = make([]int, dlen)
	for i := 0; i < dlen; i++ {
		sortedIdx[i] = i
	}

	IntsInplaceMergesort(d, sortedIdx, 0, dlen, asc)

	return
}
