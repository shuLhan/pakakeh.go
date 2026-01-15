// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package time

import (
	"fmt"
	"time"
)

func ExampleMicrosecond() {
	nano := time.Unix(1612331000, 123456789)
	fmt.Printf("%d", Microsecond(&nano))
	// Output:
	// 123456
}

func ExampleUnixMilliString() {
	nano := time.Unix(1612331000, 123456789)
	fmt.Printf("%s", UnixMilliString(nano))
	// Output:
	// 1612331000123
}
