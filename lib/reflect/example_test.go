// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
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

func ExampleTag() {
	type T struct {
		F1 int `atag:" f1 , opt1 , opt2  ,"`
		F2 int `atag:", opt1"`
		F3 int
		f4 int
	}

	var (
		t      T
		vtype  reflect.Type
		field  reflect.StructField
		val    string
		opts   []string
		x      int
		hasTag bool
	)

	vtype = reflect.TypeOf(t)

	for x = 0; x < vtype.NumField(); x++ {
		field = vtype.Field(x)
		val, opts, hasTag = Tag(field, "atag")
		fmt.Printf("%q %v %v\n", val, opts, hasTag)
	}
	//Output:
	//"f1" [opt1 opt2 ] true
	//"F2" [opt1] false
	//"F3" [] false
	//"" [] false
}

func ExampleUnmarshal_unmarshalBinary() {
	var (
		val = []byte("https://kilabit.info")

		err error
		ok  bool
	)

	// Passing variable will not work...
	var varB url.URL
	ok, err = Unmarshal(reflect.ValueOf(varB), val)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v: %q\n", ok, varB.String())

	// Pass it like these.
	ok, err = Unmarshal(reflect.ValueOf(&varB), val)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v: %q\n", ok, varB.String())

	// Passing un-initialized pointer also not working...
	var varPtrB *url.URL
	ok, err = Unmarshal(reflect.ValueOf(varPtrB), val)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v: %q\n", ok, varPtrB)

	// Pass it as **T.
	ok, err = Unmarshal(reflect.ValueOf(&varPtrB), val)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v: %q\n", ok, varPtrB)

	var ptrB = &url.URL{}
	ok, err = Unmarshal(reflect.ValueOf(&ptrB), val)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v: %q\n", ok, ptrB)

	//Output:
	//false: ""
	//true: "https://kilabit.info"
	//false: <nil>
	//true: "https://kilabit.info"
	//true: "https://kilabit.info"
}

func ExampleUnmarshal_unmarshalText() {
	var (
		vals = [][]byte{
			[]byte("123.456"),
			[]byte("123_456"),
			[]byte("123456"),
		}
		r = big.NewRat(0, 1)

		val []byte
		err error
	)

	for _, val = range vals {
		_, err = Unmarshal(reflect.ValueOf(r), val)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s\n", r)
		}
	}
	//Output:
	//15432/125
	//123456/1
	//123456/1
}

func ExampleUnmarshal_unmarshalJSON() {
	var (
		vals = [][]byte{
			[]byte("123.456"),
			[]byte("123_456"),
			[]byte("123456"),
		}
		bigInt = big.NewInt(1)

		val []byte
		err error
	)

	for _, val = range vals {
		_, err = Unmarshal(reflect.ValueOf(bigInt), val)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s\n", bigInt)
		}
	}
	//Output:
	//Unmarshal: math/big: cannot unmarshal "123.456" into a *big.Int
	//123456
	//123456
}
