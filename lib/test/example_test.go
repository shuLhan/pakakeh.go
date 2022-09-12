// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"fmt"
	"log"
	"math/big"
)

func ExampleAssert_struct() {
	type ADT struct {
		BigRat *big.Rat
		Int    int
		Bytes  []byte
	}

	var cases = []struct {
		desc string
		exp  ADT
		got  ADT
	}{{
		desc: `On field struct`,
		exp: ADT{
			BigRat: big.NewRat(123, 456),
		},
		got: ADT{
			BigRat: big.NewRat(124, 456),
		},
	}, {
		desc: `On field int`,
		exp: ADT{
			BigRat: big.NewRat(1, 2),
			Int:    1,
		},
		got: ADT{
			BigRat: big.NewRat(1, 2),
			Int:    2,
		},
	}, {
		desc: `On field []byte`,
		exp: ADT{
			Bytes: []byte(`hello, world`),
		},
		got: ADT{
			Bytes: []byte(`hello, world!`),
		},
	}, {
		desc: `On field []byte, same length`,
		exp: ADT{
			Bytes: []byte(`heelo, world!`),
		},
		got: ADT{
			Bytes: []byte(`hello, world!`),
		},
	}}

	var (
		tw = testWriter{}
	)

	for _, c := range cases {
		Assert(&tw, c.desc, c.exp, c.got)
		fmt.Println(tw.String())
		tw.Reset()
	}
	// Output:
	// !!! Assert: On field struct: ADT.BigRat: Rat.a: Int.abs: nat[0]: expecting Word(41), got Word(31)
	// !!! Assert: On field int: ADT.Int: expecting int(1), got int(2)
	// !!! Assert: On field []byte: ADT.Bytes: len(): expecting 12, got 13
	// !!! Assert: On field []byte, same length: ADT.Bytes: [2]: expecting uint8(101), got uint8(108)
}

func ExampleAssert_string() {
	var (
		tw  = testWriter{}
		exp string
		got string
	)

	exp = `a string`
	got = `b string`
	Assert(&tw, ``, exp, got)
	fmt.Println(tw.String())

	exp = `func (tw *testWriter) Fatal(args ...any)                 { fmt.Fprint(tw, args...) }`
	got = `func (tw *testWriter) Fatalf(format string, args ...any) { fmt.Fprintf(tw, format, args...) }`
	tw.Reset()
	Assert(&tw, ``, exp, got)
	fmt.Println(tw.String())
	// Output:
	// !!! Assert: expecting string(a string), got string(b string)
	// !!! strings not matched:
	// --++
	// 0 - "func (tw *testWriter) Fatal(args ...any)                 { fmt.Fprint(tw, args...) }"
	// 0 + "func (tw *testWriter) Fatalf(format string, args ...any) { fmt.Fprintf(tw, format, args...) }"
	// ^27 - "("
	// ^41 - "                "
	// ^27 + "f(format string, "
	// ^69 + "f"
	// ^56 + ", format"
}

func ExampleLoadDataDir() {
	var (
		listData []*Data
		data     *Data
		err      error
		name     string
		content  []byte
	)

	listData, err = LoadDataDir("testdata/")
	if err != nil {
		log.Fatal(err)
	}

	for _, data = range listData {
		fmt.Printf("%s\n", data.Name)
		fmt.Printf("  Flags=%v\n", data.Flag)
		fmt.Printf("  Desc=%s\n", data.Desc)
		fmt.Println("  Input")
		for name, content = range data.Input {
			fmt.Printf("    %s=%s\n", name, content)
		}
		fmt.Println("  Output")
		for name, content = range data.Output {
			fmt.Printf("    %s=%s\n", name, content)
		}
	}

	// Output:
	// data1_test.txt
	//   Flags=map[key:value]
	//   Desc=Description of test1.
	//   Input
	//     default=input.
	//   Output
	//     default=output.
	// data2_test.txt
	//   Flags=map[]
	//   Desc=
	//   Input
	//     default=another test input.
	//   Output
	//     default=another test output.
}

func ExampleLoadData() {
	var (
		data    *Data
		name    string
		content []byte
		err     error
	)

	// Content of data1_test.txt,
	//
	//	key: value
	//	Description of test1.
	//	>>>
	//	input.
	//
	//	<<<
	//	output.

	data, err = LoadData("testdata/data1_test.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", data.Name)
	fmt.Printf("  Flags=%v\n", data.Flag)
	fmt.Printf("  Desc=%s\n", data.Desc)
	fmt.Println("  Input")
	for name, content = range data.Input {
		fmt.Printf("    %s=%s\n", name, content)
	}
	fmt.Println("  Output")
	for name, content = range data.Output {
		fmt.Printf("    %s=%s\n", name, content)
	}

	// Output:
	// data1_test.txt
	//   Flags=map[key:value]
	//   Desc=Description of test1.
	//   Input
	//     default=input.
	//   Output
	//     default=output.
}
