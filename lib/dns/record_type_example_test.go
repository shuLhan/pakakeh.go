// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package dns

import "fmt"

func ExampleRecordType() {
	fmt.Println(RecordType(1))  // Known record type.
	fmt.Println(RecordType(17)) // Unregistered record type.
	// Output:
	// 1
	// 17
}

func ExampleRecordTypeFromAddress() {
	fmt.Println(RecordTypeFromAddress([]byte("127.0.0.1")))
	fmt.Println(RecordTypeFromAddress([]byte("fc00::")))
	fmt.Println(RecordTypeFromAddress([]byte("127")))
	// Output:
	// 1
	// 28
	// 0
}
