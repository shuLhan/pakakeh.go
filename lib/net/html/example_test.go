// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import "fmt"

func ExampleNormalizeForID() {
	fmt.Println(NormalizeForID(""))
	fmt.Println(NormalizeForID(" id "))
	fmt.Println(NormalizeForID("_id.1"))
	fmt.Println(NormalizeForID("1-d"))
	fmt.Println(NormalizeForID("⌘"))
	//Output:
	//_
	//_id_
	//_id_1
	//_1-d
	//___
}
