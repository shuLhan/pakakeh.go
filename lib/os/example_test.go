// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"fmt"
	"os"

	libos "git.sr.ht/~shulhan/pakakeh.go/lib/os"
)

func ExampleEnvironments() {
	os.Clearenv()
	os.Setenv(`USER`, `gopher`)
	os.Setenv(`HOME`, `/home/underground`)
	var osEnvs = libos.Environments()
	fmt.Println(osEnvs)
	// Output:
	// map[HOME:/home/underground USER:gopher]
}

func ExampleIsBinary() {
	fmt.Println(libos.IsBinary("/bin/bash"))
	fmt.Println(libos.IsBinary("io.go"))
	// Output:
	// true
	// false
}
