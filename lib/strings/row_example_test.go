// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"
)

func ExampleRow_IsEqual() {
	var row = Row{{`a`}, {`b`, `c`}}
	fmt.Println(row.IsEqual(Row{{`a`}, {`b`, `c`}}))
	fmt.Println(row.IsEqual(Row{{`a`}, {`c`, `b`}}))
	fmt.Println(row.IsEqual(Row{{`c`, `b`}, {`a`}}))
	fmt.Println(row.IsEqual(Row{{`b`, `c`}, {`a`}}))
	fmt.Println(row.IsEqual(Row{{`a`}, {`b`}}))
	// Output:
	// true
	// true
	// true
	// true
	// false
}

func ExampleRow_Join() {
	var row = Row{{`a`}, {`b`, `c`}}
	fmt.Println(row.Join(`;`, `,`))

	row = Row{{`a`}, {}}
	fmt.Println(row.Join(`;`, `,`))
	// Output:
	// a;b,c
	// a;
}
