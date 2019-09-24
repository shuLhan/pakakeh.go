// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ints

import (
	"fmt"
)

func ExampleMax() {
	ints := []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4}

	fmt.Println(Max(ints))
	// Output:
	// 9 4 true
}

func ExampleMergeByDistance() {
	a := []int{1, 5, 9}
	b := []int{4, 11, 15}

	ab := MergeByDistance(a, b, 3)
	ba := MergeByDistance(b, a, 3)
	fmt.Println(ab)
	fmt.Println(ba)
	// Output:
	// [1 5 9 15]
	// [1 5 9 15]
}
