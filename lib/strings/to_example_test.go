// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"
)

func ExampleToBytes() {
	ss := []string{"This", "is", "a", "string"}
	fmt.Printf("%s\n", ToBytes(ss))
	// Output:
	// [This is a string]
}

func ExampleToFloat64() {
	in := []string{"0", "1.1", "e", "3"}

	fmt.Println(ToFloat64(in))
	// Output: [0 1.1 0 3]
}

func ExampleToInt64() {
	in := []string{"0", "1", "e", "3.3"}

	fmt.Println(ToInt64(in))
	// Output: [0 1 0 3]
}

func ExampleToStrings() {
	i64 := []interface{}{0, 1.99, 2, 3}

	fmt.Println(ToStrings(i64))
	// Output: [0 1.99 2 3]
}
