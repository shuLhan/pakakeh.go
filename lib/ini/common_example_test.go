// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>

package ini

import "fmt"

func ExampleIsValidVarName() {
	fmt.Println(IsValidVarName(""))
	fmt.Println(IsValidVarName("1abcd"))
	fmt.Println(IsValidVarName("-abcd"))
	fmt.Println(IsValidVarName("_abcd"))
	fmt.Println(IsValidVarName(".abcd"))
	fmt.Println(IsValidVarName("a@bcd"))
	fmt.Println(IsValidVarName("a-b_c.d"))
	// Output:
	// false
	// false
	// false
	// false
	// false
	// false
	// true
}
