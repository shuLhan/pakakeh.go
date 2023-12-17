// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"fmt"
	"time"
)

func ExampleMicrosecond() {
	nano := time.Unix(1612331000, 123456789)
	fmt.Printf("%d", Microsecond(&nano))
	//Output:
	//123456
}

func ExampleUnixMilliString() {
	nano := time.Unix(1612331000, 123456789)
	fmt.Printf("%s", UnixMilliString(nano))
	//Output:
	//1612331000123
}
