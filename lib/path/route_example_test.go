// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path_test

import (
	"fmt"
	"log"

	libpath "github.com/shuLhan/share/lib/path"
)

func ExampleRoute_Set() {
	var (
		rute *libpath.Route
		err  error
	)
	rute, err = libpath.NewRoute(`/:user/:repo`)
	if err != nil {
		log.Fatal(err)
	}

	rute.Set(`user`, `shuLhan`)
	fmt.Println(rute)

	rute.Set(`repo`, `share`)
	fmt.Println(rute)
	// Output:
	// /shuLhan/:repo
	// /shuLhan/share
}
