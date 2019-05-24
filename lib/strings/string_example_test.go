// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"
)

func ExampleCleanURI() {
	text := `You can visit ftp://hostname or https://hostname/link%202 for more information`

	fmt.Printf("%s\n", CleanURI(text))
	// Output: You can visit  or  for more information
}

func ExampleCleanWikiMarkup() {
	text := `* Test image [[Image:fileto.png]].`

	fmt.Printf("%s\n", CleanWikiMarkup(text))
	// Output: * Test image .
}

func ExampleMergeSpaces() {
	line := "   a\n\nb c   d\n\n"
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
	line := `a b c [A] B C`
	fmt.Printf("%s\n", Split(line, false, false))
	fmt.Printf("%s\n", Split(line, true, false))
	fmt.Printf("%s\n", Split(line, false, true))
	fmt.Printf("%s\n", Split(line, true, true))
	// Output:
	// [a b c [A] B C]
	// [a b c A B C]
	// [a b c [A] B C]
	// [a b c]
}

func ExampleTrimNonAlnum() {
	inputs := []string{
		"[[alpha]]",
		"[[alpha",
		"alpha]]",
		"alpha",
		"alpha0",
		"1alpha",
		"1alpha0",
		"[][][]",
	}

	for _, in := range inputs {
		fmt.Printf("'%s'\n", TrimNonAlnum(in))
	}
	// Output:
	// 'alpha'
	// 'alpha'
	// 'alpha'
	// 'alpha'
	// 'alpha0'
	// '1alpha'
	// '1alpha0'
	// ''
}
