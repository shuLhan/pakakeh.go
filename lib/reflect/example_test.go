// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package reflect_test

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"reflect"

	libreflect "git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
)

func ExampleDoEqual_struct() {
	type T struct {
		vstring  string
		vnoequal string `noequal:""` // This field will not checked for equality.
	}

	var (
		t1 = &T{
			vstring:  `a string`,
			vnoequal: `skip`,
		}
		t2 = &T{
			vstring:  `a string`,
			vnoequal: `skipped too`,
		}
	)

	fmt.Println(libreflect.DoEqual(t1, t2))
	// Output:
	// <nil>
}

func ExampleIsEqual_struct() {
	type T struct {
		vstring  string
		vnoequal string `noequal:""` // This field will not checked for equality.
	}

	var (
		t1 = &T{
			vstring:  `a string`,
			vnoequal: `skip`,
		}
		t2 = &T{
			vstring:  `a string`,
			vnoequal: `skipped too`,
		}
	)

	fmt.Println(libreflect.IsEqual(t1, t2))
	// Output:
	// true
}

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
		v any
	}{
		{}, // Uninitialized any.
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
		{v: fs},              // Uninitialized interface type to any.
	}

	for _, c := range cases {
		fmt.Printf("%19T: v == nil is %5t, IsNil() is %5t\n", c.v, c.v == nil, libreflect.IsNil(c.v))
	}

	// Output:
	// <nil>: v == nil is  true, IsNil() is  true
	//                bool: v == nil is false, IsNil() is false
	//            chan int: v == nil is false, IsNil() is  true
	//      reflect_test.F: v == nil is false, IsNil() is  true
	//         map[int]int: v == nil is false, IsNil() is  true
	//         map[int]int: v == nil is false, IsNil() is false
	//     *reflect_test.T: v == nil is false, IsNil() is  true
	//     *reflect_test.T: v == nil is false, IsNil() is false
	//               []int: v == nil is false, IsNil() is  true
	//               []int: v == nil is false, IsNil() is false
	//                 int: v == nil is false, IsNil() is false
	//               <nil>: v == nil is  true, IsNil() is  true
	// *errors.errorString: v == nil is false, IsNil() is false
	//               <nil>: v == nil is  true, IsNil() is  true
}

type InvalidMarshalText struct{}

func (imt *InvalidMarshalText) MarshalText() (string, error) {
	return "", nil
}

type ErrorMarshalJSON struct{}

func (emj *ErrorMarshalJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New(`ErrorMarshalJSON: test`)
}

func ExampleMarshal() {
	var (
		vint    = 1
		vURL, _ = url.Parse("https://example.org")
		bigRat  = big.NewRat(100, 2)
		bigInt  = big.NewInt(50)
		imt     = &InvalidMarshalText{}
		emj     = &ErrorMarshalJSON{}

		out []byte
		err error
	)

	out, err = libreflect.Marshal(vint)
	fmt.Println(out, err)

	out, err = libreflect.Marshal(&vint)
	fmt.Println(out, err)

	out, err = libreflect.Marshal(vURL)
	fmt.Println(string(out), err)

	out, err = libreflect.Marshal(bigRat)
	fmt.Println(string(out), err)

	out, err = libreflect.Marshal(bigInt)
	fmt.Println(string(out), err)

	out, err = libreflect.Marshal(imt)
	fmt.Println(string(out), err)

	out, err = libreflect.Marshal(emj)
	fmt.Println(string(out), err)

	// Output:
	// [] <nil>
	// [] <nil>
	// https://example.org <nil>
	// 50 <nil>
	// 50 <nil>
	//  Marshal: expecting first return as []byte got string
	//  Marshal: ErrorMarshalJSON: test
}

func ExampleSet_bool() {
	type Bool bool

	var (
		err    error
		vbool  bool
		mybool Bool
	)

	err = libreflect.Set(reflect.ValueOf(&vbool), "YES")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("YES:", vbool)
	}

	err = libreflect.Set(reflect.ValueOf(&vbool), "TRUE")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("TRUE:", vbool)
	}

	err = libreflect.Set(reflect.ValueOf(&vbool), "False")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("False:", vbool)
	}

	err = libreflect.Set(reflect.ValueOf(&vbool), "1")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("1:", vbool)
	}

	err = libreflect.Set(reflect.ValueOf(&mybool), "true")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("true:", mybool)
	}

	// Output:
	// YES: true
	// TRUE: true
	// False: false
	// 1: true
	// true: true
}

func ExampleSet_float() {
	type myFloat float32

	var (
		vf32    float32
		myfloat myFloat
		err     error
	)

	err = libreflect.Set(reflect.ValueOf(&vf32), "1.223")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vf32)
	}

	err = libreflect.Set(reflect.ValueOf(&myfloat), "999.999")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(myfloat)
	}

	// Output:
	// 1.223
	// 999.999
}

func ExampleSet_int() {
	type myInt int

	var (
		vint   int
		vint8  int8
		vint16 int16
		vmyint myInt
		err    error
	)

	err = libreflect.Set(reflect.ValueOf(&vint), "")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vint)
	}

	err = libreflect.Set(reflect.ValueOf(&vint), "1")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vint)
	}

	err = libreflect.Set(reflect.ValueOf(&vint8), "-128")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vint8)
	}

	// Value of int16 is overflow.
	err = libreflect.Set(reflect.ValueOf(&vint16), "32768")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vint16)
	}

	err = libreflect.Set(reflect.ValueOf(&vmyint), "32768")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vmyint)
	}

	// Output:
	// 0
	// 1
	// -128
	// error: Set: int16 value is overflow: 32768
	// 32768
}

func ExampleSet_sliceByte() {
	type myBytes []byte

	var (
		vbytes   []byte
		vmyBytes myBytes
		err      error
	)

	err = libreflect.Set(reflect.ValueOf(vbytes), "Show me")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(string(vbytes))
	}

	err = libreflect.Set(reflect.ValueOf(&vbytes), "a hero")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(string(vbytes))
	}

	err = libreflect.Set(reflect.ValueOf(&vbytes), "")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(string(vbytes))
	}

	err = libreflect.Set(reflect.ValueOf(&vmyBytes), "and I will write you a tragedy")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(string(vmyBytes))
	}

	// Output:
	// error: Set: object []uint8 is not setable
	// a hero
	//
	// and I will write you a tragedy
}

func ExampleSet_sliceString() {
	var (
		vstring []string
		err     error
	)

	err = libreflect.Set(reflect.ValueOf(vstring), "Show me")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vstring)
	}

	err = libreflect.Set(reflect.ValueOf(&vstring), "a hero")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vstring)
	}

	err = libreflect.Set(reflect.ValueOf(&vstring), "and I will write you a tragedy")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(vstring)
	}

	// Output:
	// error: Set: object []string is not setable
	// [a hero]
	// [a hero and I will write you a tragedy]
}

func ExampleSet_unmarshal() {
	var (
		rat    = big.NewRat(0, 1)
		myURL  = &url.URL{}
		bigInt = big.NewInt(1)

		err error
	)

	// This Set will call UnmarshalText on big.Rat.
	err = libreflect.Set(reflect.ValueOf(rat), "1.234")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(rat.FloatString(4))
	}

	err = libreflect.Set(reflect.ValueOf(rat), "")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(rat.FloatString(4))
	}

	// This Set will call UnmarshalBinary on url.URL.
	err = libreflect.Set(reflect.ValueOf(myURL), "https://kilabit.info")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(myURL)
	}

	// This Set will call UnmarshalJSON.
	err = libreflect.Set(reflect.ValueOf(bigInt), "123_456")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(bigInt)
	}

	err = libreflect.Set(reflect.ValueOf(bigInt), "")
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(bigInt)
	}

	// Output:
	// 1.2340
	// 0.0000
	// https://kilabit.info
	// 123456
	// 0
}

func ExampleTag() {
	type T struct {
		F1 int `atag:" f1 , opt1 , opt2  ,"`
		F2 int `atag:", opt1"`
		F3 int
		F4 int `atag:" - ,opt1"`
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
		val, opts, hasTag = libreflect.Tag(field, "atag")
		fmt.Println(val, opts, hasTag)
	}
	// Output:
	// f1 [opt1 opt2 ] true
	// F2 [opt1] false
	// F3 [] false
	//  [] false
}

func ExampleUnmarshal_unmarshalBinary() {
	var (
		val = []byte("https://kilabit.info")

		err error
		ok  bool
	)

	// Passing variable will not work...
	var varB url.URL
	ok, err = libreflect.Unmarshal(reflect.ValueOf(varB), val)
	if err != nil {
		return
	}
	fmt.Println(varB.String(), ok)

	// Pass it like these.
	ok, err = libreflect.Unmarshal(reflect.ValueOf(&varB), val)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(varB.String(), ok)

	// Passing un-initialized pointer also not working...
	var varPtrB *url.URL
	ok, err = libreflect.Unmarshal(reflect.ValueOf(varPtrB), val)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(varPtrB, ok)

	// Pass it as **T.
	ok, err = libreflect.Unmarshal(reflect.ValueOf(&varPtrB), val)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(varPtrB, ok)

	var ptrB = &url.URL{}
	ok, err = libreflect.Unmarshal(reflect.ValueOf(&ptrB), val)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ptrB, ok)

	// Output:
	//   false
	// https://kilabit.info true
	// <nil> false
	// https://kilabit.info true
	// https://kilabit.info true
}

func ExampleUnmarshal_unmarshalText() {
	var (
		vals = [][]byte{
			[]byte(""),
			[]byte("123.456"),
			[]byte("123_456"),
			[]byte("123456"),
		}
		r = big.NewRat(0, 1)

		val []byte
		err error
	)

	for _, val = range vals {
		_, err = libreflect.Unmarshal(reflect.ValueOf(r), val)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(r)
		}
	}
	// Output:
	// 0/1
	// 15432/125
	// 123456/1
	// 123456/1
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
		_, err = libreflect.Unmarshal(reflect.ValueOf(bigInt), val)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(bigInt)
		}
	}
	// Output:
	// Unmarshal: math/big: cannot unmarshal "123.456" into a *big.Int
	// 123456
	// 123456
}
