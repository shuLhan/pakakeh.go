// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

//
// Ints64FindMax given a slice of integer, return the maximum value in slice
// and its index.
//
// If data is empty, it will return `-1` in value and index, and false in ok.
//
// Example, given a slice of data: [0 1 2 3 4], it will return 4 as max and 4
// as index of maximum value.
//
func Ints64FindMax(d []int64) (maxv int64, maxi int, ok bool) {
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
// Ints64FindMin given a slice of integer, return the minimum value in slice
// and its index.
//
// If data is empty, return -1 in value and index; and false in ok.
//
// Example, given a slice of data: [0 1 2 3 4], it will return 0 as min and 0
// as minimum index.
//
func Ints64FindMin(d []int64) (minv int64, mini int, ok bool) {
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
// Ints64Sum return sum of all value in slice.
//
func Ints64Sum(d []int64) (sum int64) {
	for _, v := range d {
		sum += v
	}
	return sum
}

//
// Ints64Count will count number of class in data.
//
func Ints64Count(d []int64, class int64) (count int) {
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
// Ints64Counts will count class in data and return each of the counter.
//
// For example, if data is "[1,1,2]" and class is "[1,2]", this function will
// return "[2,1]".
//
//	idx class  count
//	0 : 1   -> 2
//	1 : 2   -> 1
//
func Ints64Counts(d, classes []int64) (counts []int) {
	if len(classes) <= 0 {
		return
	}

	counts = make([]int, len(classes))

	for x, c := range classes {
		counts[x] = Ints64Count(d, c)
	}
	return
}

//
// Ints64MaxCountOf will count number of occurrence of each element of classes
// in data and return the class with maximum count.
//
// If `classes` is empty, it will return -1 and false.
// If `data` is empty, it will return -2 and false.
// If classes has the same count value, then the first max in the class will be
// returned.
//
// For example, given a data [0, 1, 0, 1, 0] and classes [0, 1], the function
// will count 0 as 3, 1 as 2; and return 0.
//
func Ints64MaxCountOf(d, classes []int64) (int64, bool) {
	if len(classes) == 0 {
		return -1, false
	}
	if len(d) == 0 {
		return -2, false
	}

	counts := Ints64Counts(d, classes)

	_, maxi, _ := IntsFindMax(counts)
	if maxi < 0 {
		return -1, false
	}

	return classes[maxi], true
}

//
// Ints64Swap swap two indices value of integer.
//
func Ints64Swap(d []int64, x, y int) {
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
// Ints64IsExist will return true if value `v` exist in slice of `d`,
// otherwise it will return false.
//
func Ints64IsExist(d []int64, i int64) bool {
	for _, v := range d {
		if i == v {
			return true
		}
	}
	return false
}

//
// Ints64InsertionSort will sort the data using insertion-sort algorithm.
//
// Parameters:
// - `data` is slice that will be sorted.
// - `idx` is indices of data.
// - `l` is starting index of slice to be sorted.
// - `r` is end index of slice to be sorted.
//
func Ints64InsertionSort(d []int64, ids []int, l, r int, asc bool) {
	for x := l; x < r; x++ {
		for y := x + 1; y < r; y++ {
			if asc {
				if d[x] > d[y] {
					IntsSwap(ids, x, y)
					Ints64Swap(d, x, y)
				}
			} else {
				if d[x] < d[y] {
					IntsSwap(ids, x, y)
					Ints64Swap(d, x, y)
				}
			}
		}
	}
}

//
// Ints64SortByIndex will sort the slice `d` using sorted index `sortedIds`.
//
func Ints64SortByIndex(d *[]int64, sortedIds []int) {
	newd := make([]int64, len(*d))

	for i := range sortedIds {
		newd[i] = (*d)[sortedIds[i]]
	}

	(*d) = newd
}

//
// Ints64InplaceMergesort in-place merge-sort without memory allocation.
//
func Ints64InplaceMergesort(d []int64, idx []int, l, r int, asc bool) {
	// (0) If data length == Threshold, then
	if l+SortThreshold >= r {
		// (0.1) use insertion sort.
		Ints64InsertionSort(d, idx, l, r, asc)
		return
	}

	// (1) Divide into left and right.
	res := (r + l) % 2
	c := (r + l) / 2
	if res == 1 {
		c++
	}

	// (2) Sort left.
	Ints64InplaceMergesort(d, idx, l, c, asc)

	// (3) Sort right.
	Ints64InplaceMergesort(d, idx, c, r, asc)

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

	ints64InplaceMerge(d, idx, l, c, r, asc)
}

//
// Let `x` be the first index of left-side, `y` be the first index of
// the right-side, and `r` as length of slice `d`
//
func ints64InplaceMerge(d []int64, idx []int, x, y, r int, asc bool) {
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
		ylast = ints64MoveY(d, x, y, r, asc)

		// (4.3.3) SWAP DATA, X, Y, YLAST
		ints64Multiswap(d, idx, x, y, ylast)

	next:
		// (4.3.4) LET Y := the minimum value between x and r on `d`
		ints64MinY(d, &x, &y, r, asc)
	}
}

// (4.3.4) LET Y := the minimum value between x and r on `d`
func ints64MinY(d []int64, x, y *int, r int, asc bool) {
	for *x < r {
		if asc {
			*y = ints64Min(d, *x, r)
		} else {
			*y = ints64Max(d, *x, r)
		}

		if *y != *x {
			break
		}
		(*x)++
	}
}

func ints64MoveY(d []int64, x, y, r int, asc bool) int {
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

func ints64Multiswap(d []int64, idx []int, x, y, ylast int) int {
	for y < ylast {
		IntsSwap(idx, x, y)
		Ints64Swap(d, x, y)
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

func ints64Min(d []int64, l, r int) (m int) {
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

func ints64Max(d []int64, l, r int) (m int) {
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
// Ints64IndirectSort will sort the data and return the sorted index.
//
func Ints64IndirectSort(d []int64, asc bool) (sortedIdx []int) {
	dlen := len(d)

	sortedIdx = make([]int, dlen)
	for i := 0; i < dlen; i++ {
		sortedIdx[i] = i
	}

	Ints64InplaceMergesort(d, sortedIdx, 0, dlen, asc)

	return
}
