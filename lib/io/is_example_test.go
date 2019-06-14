// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"fmt"
)

func ExampleIsBinary() {
	fmt.Println(IsBinary("/bin/bash"))
	fmt.Println(IsBinary("io.go"))
	// Output:
	// true
	// false
}
