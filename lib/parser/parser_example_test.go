// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"strings"
)

func ExampleNew() {
	content := "[test]\nkey = value"
	p := New(content, "=[]")

	for {
		token, del := p.Token()
		token = strings.TrimSpace(token)
		fmt.Printf("%q %q\n", token, del)
		if del == 0 {
			break
		}
	}
	// Output:
	// "" '['
	// "test" ']'
	// "key" '='
	// "value" '\x00'
}
