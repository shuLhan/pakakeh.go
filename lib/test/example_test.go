// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"fmt"
	"log"
)

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
