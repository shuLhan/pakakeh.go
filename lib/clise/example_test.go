// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clise

import (
	"encoding/json"
	"fmt"
)

func ExampleClise_MarshalJSON() {
	type T struct {
		Int    int
		String string
	}

	var (
		c = New(3)

		bjson []byte
		err   error
	)

	c.Push(1, 2, 3, 4)
	bjson, err = json.Marshal(c)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(bjson))
	}

	c.Push("Hello", "Clise", "MarshalJSON")
	bjson, err = json.Marshal(c)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(bjson))
	}

	c.Push(&T{1, "Hello"}, &T{2, "world"})
	bjson, err = json.Marshal(c)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(bjson))
	}

	//Output:
	//[2,3,4]
	//["Hello","Clise","MarshalJSON"]
	//["MarshalJSON",{"Int":1,"String":"Hello"},{"Int":2,"String":"world"}]
}
