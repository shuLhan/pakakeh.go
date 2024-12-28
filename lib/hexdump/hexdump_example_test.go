// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package hexdump_test

import (
	"bytes"
	"fmt"
	"log"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/hexdump"
)

func ExampleParse() {
	var (
		in = []byte(`0000000 6548 6c6c 2c6f 7720 726f 646c 0021`)

		out []byte
		err error
	)

	out, err = hexdump.Parse(in, false)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, libbytes.TrimNull(out))

	// Output:
	// Hello, world!
}

func ExamplePrettyPrint() {
	var (
		data = []byte{1, 2, 3, 'H', 'e', 'l', 'l', 'o', 254, 255}
		bb   bytes.Buffer
	)

	hexdump.PrettyPrint(&bb, `PrettyPrint`, data)
	fmt.Println(bb.String())
	// Output:
	// PrettyPrint
	//           |  0  1  2  3  4  5  6  7 | 01234567 |   0   1   2   3   4   5   6   7 |
	//           |  8  9  A  B  C  D  E  F | 89ABCDEF |   8   9   A   B   C   D   E   F |
	// 0x00000000| 01 02 03 48 65 6c 6c 6f | ...Hello |   1   2   3  72 101 108 108 111 |0
	// 0x00000008| fe ff                   | ..       | 254 255                         |8
}

func ExamplePrint() {
	title := `Print`
	data := []byte(`Hello, world !`)
	hexdump.Print(title, data, 5)

	// Output:
	// Print
	//    0 - 48 65 6C 6C 6F || H e l l o
	//    5 - 2C 20 77 6F 72 || , . w o r
	//   10 - 6C 64 20 21    || l d . !
}
