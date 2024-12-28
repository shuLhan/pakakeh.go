// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package slices_test

import (
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/slices"
)

func ExampleMax2() {
	ints := []int{5, 6, 7, 8, 9, 0, 1, 2, 3, 4}

	fmt.Println(slices.Max2(ints))

	// Output:
	// 9 4
}

func ExampleMergeByDistance() {
	a := []int{1, 5, 9}
	b := []int{4, 11, 15}

	ab := slices.MergeByDistance(a, b, 3)
	ba := slices.MergeByDistance(b, a, 3)
	fmt.Println(ab)
	fmt.Println(ba)

	// Output:
	// [1 5 9 15]
	// [1 5 9 15]
}
