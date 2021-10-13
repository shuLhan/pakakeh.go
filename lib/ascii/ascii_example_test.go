// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ascii

import (
	"fmt"
)

func ExampleIsAlnum() {
	chars := []byte("0aZ!.")

	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, IsAlnum(c))
	}
	// Output:
	// 0: true
	// a: true
	// Z: true
	// !: false
	// .: false
}

func ExampleIsAlpha() {
	chars := []byte("0aZ!.")
	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, IsAlpha(c))
	}
	// Output:
	// 0: false
	// a: true
	// Z: true
	// !: false
	// .: false
}

func ExampleIsDigit() {
	chars := []byte("0aZ!.")
	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, IsDigit(c))
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
		"012",
		"012 ",
		" 012 ",
		"0.",
		"0.1",
		"0.1a",
	}

	for _, s := range inputs {
		fmt.Printf("%s: %t\n", s, IsDigits([]byte(s)))
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
	chars := []byte("09afgAFG")
	for _, c := range chars {
		fmt.Printf("%c: %t\n", c, IsHex(c))
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
	fmt.Printf("\\t: %t\n", IsSpace('\t'))
	fmt.Printf("\\n: %t\n", IsSpace('\n'))
	fmt.Printf("\\v: %t\n", IsSpace('\v'))
	fmt.Printf("\\f: %t\n", IsSpace('\f'))
	fmt.Printf("\\r: %t\n", IsSpace('\r'))
	fmt.Printf(" : %t\n", IsSpace(' '))
	fmt.Printf("	: %t\n", IsSpace('	'))
	fmt.Printf("\\: %t\n", IsSpace('\\'))
	fmt.Printf("0: %t\n", IsSpace('0'))
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

func ExampleRandom() {
	fmt.Printf("Random 5 Letters: %s\n", Random([]byte(Letters), 5))
	fmt.Printf("Random 5 LettersNumber: %s\n", Random([]byte(LettersNumber), 5))
	fmt.Printf("Random 5 HexaLETTERS: %s\n", Random([]byte(HexaLETTERS), 5))
	fmt.Printf("Random 5 HexaLetters: %s\n", Random([]byte(HexaLetters), 5))
	fmt.Printf("Random 5 Hexaletters: %s\n", Random([]byte(Hexaletters), 5))
	fmt.Printf("Random 5 binary: %s\n", Random([]byte("01"), 5))
	// Output:
	// Random 5 Letters: XVlBz
	// Random 5 LettersNumber: 80Aep
	// Random 5 HexaLETTERS: 6F218
	// Random 5 HexaLetters: 675DA
	// Random 5 Hexaletters: fa82f
	// Random 5 binary: 11001
}

func ExampleToLower() {
	in := []byte("@ABCDEFGhijklmnoPQRSTUVWxyz{12345678")
	fmt.Printf("%s\n", ToLower(in))
	// Output:
	// @abcdefghijklmnopqrstuvwxyz{12345678
}

func ExampleToUpper() {
	in := []byte("@ABCDEFGhijklmnoPQRSTUVWxyz{12345678")
	fmt.Printf("%s\n", ToUpper(in))
	// Output:
	// @ABCDEFGHIJKLMNOPQRSTUVWXYZ{12345678
}
