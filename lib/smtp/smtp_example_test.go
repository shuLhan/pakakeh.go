// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"fmt"
)

func ExampleParsePath() {
	mb, _ := ParsePath([]byte(`<@domain.com,@domain.net:local.part@domain.com>`))
	fmt.Printf("%s\n", mb)
	mb, _ = ParsePath([]byte(`<local.part@domain.com>`))
	fmt.Printf("%s\n", mb)
	mb, _ = ParsePath([]byte(`<local>`))
	fmt.Printf("%s\n", mb)
	// Output:
	// local.part@domain.com
	// local.part@domain.com
	// local
}
