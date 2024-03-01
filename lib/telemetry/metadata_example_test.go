// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry_test

import (
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/telemetry"
)

func ExampleMetadata() {
	var md = telemetry.NewMetadata()

	// The new Metadata has version=0.
	fmt.Println(md.Version(), md.String(), ".")

	// Setting a key increase the version to 1
	md.Set(`host`, `localhost`)
	fmt.Println(md.Version(), md.String(), ".")

	// ... even if the key already exist.
	md.Set(`host`, `my.localhost`)
	fmt.Println(md.Version(), md.String(), ".")

	// Deleting a key increase the version too.
	md.Delete(`host`)
	fmt.Println(md.Version(), md.String(), ".")

	// But if the key is not exist, it will not increase the version.
	md.Delete(`host`)
	fmt.Println(md.Version(), md.String(), ".")

	// Output:
	// 0  .
	// 1 host=localhost .
	// 2 host=my.localhost .
	// 3  .
	// 3  .
}
