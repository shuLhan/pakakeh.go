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
