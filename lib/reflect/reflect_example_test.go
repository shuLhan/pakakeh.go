// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import (
	"errors"
	"fmt"
	"net/http"
)

type F func()

type T struct{}

func (t *T) J() bool {
	return true
}

func ExampleIsNil() {
	var (
		aBoolean   bool
		aChannel   chan int
		aFunction  F
		aMap       map[int]int
		aPtr       *T
		aSlice     []int
		anInt      int
		emptyError error
		fs         http.FileSystem
	)

	cases := []struct {
		v interface{}
	}{
		{}, // Uninitialized interface{}.
		{v: aBoolean},
		{v: aChannel},          // Uninitialized channel.
		{v: aFunction},         // Empty func type.
		{v: aMap},              // Uninitialized map.
		{v: make(map[int]int)}, // Initialized map.
		{v: aPtr},              // Uninitialized pointer to struct.
		{v: &T{}},              // Initialized pointer to struct.
		{v: aSlice},            // Uninitialized slice.
		{v: make([]int, 0)},    // Initialized slice.
		{v: anInt},
		{v: emptyError},
		{v: errors.New("e")}, // Initialized error.
		{v: fs},              // Uninitialized interface type to interface{}.
	}

	for _, c := range cases {
		fmt.Printf("%19T: v == nil is %5t, IsNil() is %5t\n", c.v, c.v == nil, IsNil(c.v))
	}

	//Output:
	// <nil>: v == nil is  true, IsNil() is  true
	//                bool: v == nil is false, IsNil() is false
	//            chan int: v == nil is false, IsNil() is  true
	//           reflect.F: v == nil is false, IsNil() is  true
	//         map[int]int: v == nil is false, IsNil() is  true
	//         map[int]int: v == nil is false, IsNil() is false
	//          *reflect.T: v == nil is false, IsNil() is  true
	//          *reflect.T: v == nil is false, IsNil() is false
	//               []int: v == nil is false, IsNil() is  true
	//               []int: v == nil is false, IsNil() is false
	//                 int: v == nil is false, IsNil() is false
	//               <nil>: v == nil is  true, IsNil() is  true
	// *errors.errorString: v == nil is false, IsNil() is false
	//               <nil>: v == nil is  true, IsNil() is  true
}
