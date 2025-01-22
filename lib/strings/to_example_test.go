// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package strings

import (
	"fmt"
)

func ExampleToBytes() {
	var (
		ss     = []string{`This`, `is`, `a`, `string`}
		sbytes = ToBytes(ss)
	)

	fmt.Printf(`%s`, sbytes)
	// Output: [This is a string]
}

func ExampleToFloat64() {
	var (
		in   = []string{`0`, `1.1`, `e`, `3`}
		sf64 = ToFloat64(in)
	)

	fmt.Println(sf64)
	// Output: [0 1.1 0 3]
}

func ExampleToInt64() {
	var (
		in   = []string{`0`, `1`, `e`, `3.3`}
		si64 = ToInt64(in)
	)

	fmt.Println(si64)
	// Output: [0 1 0 3]
}

func ExampleToStrings() {
	var (
		i64 = []any{0, 1.99, 2, 3}
		ss  = ToStrings(i64)
	)

	fmt.Println(ss)
	// Output: [0 1.99 2 3]
}
