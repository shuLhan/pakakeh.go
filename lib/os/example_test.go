// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

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
	fmt.Println(libos.IsBinary(`testdata/exp.bz2`))
	fmt.Println(libos.IsBinary(`os.go`))
	// Output:
	// true
	// false
}
