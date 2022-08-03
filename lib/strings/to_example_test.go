// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"
)

func ExampleToBytes() {
	var (
		ss              = []string{`This`, `is`, `a`, `string`}
		sbytes [][]byte = ToBytes(ss)
	)

	fmt.Printf(`%s`, sbytes)
	// Output: [This is a string]
}

func ExampleToFloat64() {
	var (
		in             = []string{`0`, `1.1`, `e`, `3`}
		sf64 []float64 = ToFloat64(in)
	)

	fmt.Println(sf64)
	// Output: [0 1.1 0 3]
}

func ExampleToInt64() {
	var (
		in           = []string{`0`, `1`, `e`, `3.3`}
		si64 []int64 = ToInt64(in)
	)

	fmt.Println(si64)
	// Output: [0 1 0 3]
}

func ExampleToStrings() {
	var (
		i64          = []interface{}{0, 1.99, 2, 3}
		ss  []string = ToStrings(i64)
	)

	fmt.Println(ss)
	// Output: [0 1.99 2 3]
}
