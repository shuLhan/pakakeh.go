// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import "fmt"

func ExampleParseMailbox() {
	fmt.Printf("%v\n", ParseMailbox([]byte("local")))
	fmt.Printf("%v\n", ParseMailbox([]byte("Name <domain>")))
	fmt.Printf("%v\n", ParseMailbox([]byte("local@domain")))
	fmt.Printf("%v\n", ParseMailbox([]byte("Name <local@domain>")))
	//Output:
	//<nil>
	//Name <@domain>
	//<local@domain>
	//Name <local@domain>
}
