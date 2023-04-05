// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"
)

func ExampleAlnum() {
	type testCase struct {
		text      string
		withSpace bool
	}

	var cases = []testCase{
		{`A, b.c`, false},
		{`A, b.c`, true},
		{`A1 b`, false},
		{`A1 b`, true},
	}

	var (
		c testCase
	)
	for _, c = range cases {
		fmt.Println(Alnum(c.text, c.withSpace))
	}
	// Output:
	// Abc
	// A bc
	// A1b
	// A1 b
}

func ExampleCleanURI() {
	var text = `You can visit ftp://hostname or https://hostname/link%202 for more information`

	fmt.Println(CleanURI(text))
	// Output: You can visit  or  for more information
}

func ExampleCleanWikiMarkup() {
	var text = `* Test image [[Image:fileto.png]].`

	fmt.Println(CleanWikiMarkup(text))
	// Output: * Test image .
}

func ExampleMergeSpaces() {
	var line = "   a\n\nb c   d\n\n"
	fmt.Printf("Without merging newline: '%s'\n", MergeSpaces(line, false))
	fmt.Printf("With merging newline: '%s'\n", MergeSpaces(line, true))
	// Output:
	// Without merging newline: ' a
	//
	// b c d
	//
	// '
	// With merging newline: ' a
	// b c d
	// '
}

func ExampleSplit() {
	var line = `a b   c [A] B C`
	fmt.Println(Split(line, false, false))
	fmt.Println(Split(line, true, false))
	fmt.Println(Split(line, false, true))
	fmt.Println(Split(line, true, true))
	// Output:
	// [a b c [A] B C]
	// [a b c A B C]
	// [a b c [A]]
	// [a b c]
}

func ExampleTrimNonAlnum() {
	var (
		inputs = []string{
			`[[alpha]]`,
			`[[alpha`,
			`alpha]]`,
			`alpha`,
			`alpha0`,
			`1alpha`,
			`1alpha0`,
			`[a][b][c]`,
			`[][][]`,
		}
		in string
	)

	for _, in = range inputs {
		fmt.Println(TrimNonAlnum(in))
	}
	// Output:
	// alpha
	// alpha
	// alpha
	// alpha
	// alpha0
	// 1alpha
	// 1alpha0
	// a][b][c
	//
}
