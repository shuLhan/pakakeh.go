// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"fmt"

	libos "github.com/shuLhan/share/lib/os"
)

func ExampleIsBinary() {
	fmt.Println(libos.IsBinary("/bin/bash"))
	fmt.Println(libos.IsBinary("io.go"))
	// Output:
	// true
	// false
}
