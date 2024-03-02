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
		String string
		Int    int
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

	c.Push(&T{Int: 1, String: `Hello`}, &T{Int: 2, String: `world`})
	bjson, err = json.Marshal(c)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(bjson))
	}

	// Output:
	// [2,3,4]
	// ["Hello","Clise","MarshalJSON"]
	// ["MarshalJSON",{"String":"Hello","Int":1},{"String":"world","Int":2}]
}

func ExampleClise_UnmarshalJSON() {
	var (
		clise = New(3)

		cases = []string{
			`{"k":1}`, // Non array.
			`[2,3,4]`,
			`[2,3,4,5]`, // Array elements greater that maximum size.
			`["Hello","Clise","MarshalJSON"]`,
			`["MarshalJSON",{"Int":1,"String":"Hello"},{"Int":2,"String":"world"}]`,
		}

		rawJSON string
		err     error
	)

	for _, rawJSON = range cases {
		err = clise.UnmarshalJSON([]byte(rawJSON))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(clise.Slice())
		}
	}

	// Output:
	// UnmarshalJSON: json: cannot unmarshal object into Go value of type []interface {}
	// [2 3 4]
	// [3 4 5]
	// [Hello Clise MarshalJSON]
	// [MarshalJSON map[Int:1 String:Hello] map[Int:2 String:world]]
}
