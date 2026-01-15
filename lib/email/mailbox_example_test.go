// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package email

import "fmt"

func ExampleParseMailbox() {
	fmt.Printf("%v\n", ParseMailbox([]byte("local")))
	fmt.Printf("%v\n", ParseMailbox([]byte("Name <domain>")))
	fmt.Printf("%v\n", ParseMailbox([]byte("local@domain")))
	fmt.Printf("%v\n", ParseMailbox([]byte("Name <local@domain>")))
	// Output:
	// <nil>
	// Name <@domain>
	// <local@domain>
	// Name <local@domain>
}
