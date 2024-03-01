// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp_test

import (
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/smtp"
)

func ExampleParsePath() {
	var mb []byte

	mb, _ = smtp.ParsePath([]byte(`<@domain.com,@domain.net:local.part@domain.com>`))
	fmt.Printf("%s\n", mb)
	mb, _ = smtp.ParsePath([]byte(`<local.part@domain.com>`))
	fmt.Printf("%s\n", mb)
	mb, _ = smtp.ParsePath([]byte(`<local>`))
	fmt.Printf("%s\n", mb)
	// Output:
	// local.part@domain.com
	// local.part@domain.com
	// local
}
