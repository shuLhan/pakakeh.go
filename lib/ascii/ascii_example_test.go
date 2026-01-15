// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package ascii_test

import (
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

func ExampleIsAlnum() {
	chars := []byte(`0aZ!.`)

	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, ascii.IsAlnum(c))
	}
	// Output:
	// 0: true
	// a: true
	// Z: true
	// !: false
	// .: false
}

func ExampleIsAlpha() {
	chars := []byte(`0aZ!.`)
	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, ascii.IsAlpha(c))
	}
	// Output:
	// 0: false
	// a: true
	// Z: true
	// !: false
	// .: false
}

func ExampleIsDigit() {
	chars := []byte(`0aZ!.`)
	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, ascii.IsDigit(c))
	}
	// Output:
	// 0: true
	// a: false
	// Z: false
	// !: false
	// .: false
}

func ExampleIsDigits() {
	inputs := []string{
		`012`,
		`012 `,
		` 012 `,
		`0.`,
		`0.1`,
		`0.1a`,
	}

	for _, s := range inputs {
		fmt.Printf("%s: %t\n", s, ascii.IsDigits([]byte(s)))
	}
	// Output:
	// 012: true
	// 012 : false
	//  012 : false
	// 0.: false
	// 0.1: false
	// 0.1a: false
}

func ExampleIsHex() {
	chars := []byte(`09afgAFG`)
	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, ascii.IsHex(c))
	}
	// Output:
	// 0: true
	// 9: true
	// a: true
	// f: true
	// g: false
	// A: true
	// F: true
	// G: false
}

func ExampleIsSpace() {
	fmt.Printf("\\t: %t\n", ascii.IsSpace('\t'))
	fmt.Printf("\\n: %t\n", ascii.IsSpace('\n'))
	fmt.Printf("\\v: %t\n", ascii.IsSpace('\v'))
	fmt.Printf("\\f: %t\n", ascii.IsSpace('\f'))
	fmt.Printf("\\r: %t\n", ascii.IsSpace('\r'))
	fmt.Printf(" : %t\n", ascii.IsSpace(' '))
	fmt.Printf("	: %t\n", ascii.IsSpace('	'))
	fmt.Printf("\\: %t\n", ascii.IsSpace('\\'))
	fmt.Printf("0: %t\n", ascii.IsSpace('0'))
	// Output:
	// \t: true
	// \n: true
	// \v: true
	// \f: true
	// \r: true
	//  : true
	// 	: true
	// \: false
	// 0: false
}

func ExampleToLower() {
	in := []byte(`@ABCDEFGhijklmnoPQRSTUVWxyz{12345678`)
	fmt.Printf("%s\n", ascii.ToLower(in))
	// Output:
	// @abcdefghijklmnopqrstuvwxyz{12345678
}

func ExampleToUpper() {
	in := []byte(`@ABCDEFGhijklmnoPQRSTUVWxyz{12345678`)
	fmt.Printf("%s\n", ascii.ToUpper(in))
	// Output:
	// @ABCDEFGHIJKLMNOPQRSTUVWXYZ{12345678
}
