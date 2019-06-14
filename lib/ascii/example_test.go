// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ascii

import (
	"fmt"
)

func ExampleToLower() {
	in := []byte("@ABCDEFGhijklmnoPQRSTUVWxyz{12345678")

	ToLower(&in)

	fmt.Println(string(in))
	// Output:
	// @abcdefghijklmnopqrstuvwxyz{12345678
}

func ExampleToUpper() {
	in := []byte("@ABCDEFGhijklmnoPQRSTUVWxyz{12345678")

	ToUpper(&in)

	fmt.Println(string(in))
	// Output:
	// @ABCDEFGHIJKLMNOPQRSTUVWXYZ{12345678
}
