// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import "fmt"

func ExampleRat_RoundingUpToZero() {
	values := []string{
		"2.555", "2.5", "2.1", "2.05", "2.01", "2.004", "2.0012",
		"-2.0012", "-2.004", "-2.01", "-2.05", "-2.1", "-2.555",
	}
	for _, val := range values {
		r := NewRat(val)
		r.RoundingUpToZero()
		fmt.Printf("%s: %s\n", val, r)
	}
	//Output:
	// 2.555: 3
	// 2.5: 3
	// 2.1: 3
	// 2.05: 2.1
	// 2.01: 2.1
	// 2.004: 2.01
	// 2.0012: 2.01
	// -2.0012: -2
	// -2.004: -2
	// -2.01: -2
	// -2.05: -2
	// -2.1: -2
	// -2.555: -2
}
