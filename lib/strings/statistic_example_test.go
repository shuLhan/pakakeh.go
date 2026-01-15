// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package strings

import (
	"fmt"
)

func ExampleCountAlnum() {
	fmt.Println(CountAlnum(`// A b c 1 2 3`))
	// Output: 6
}

func ExampleCountAlnumDistribution() {
	var (
		chars  []rune
		counts []int
	)

	chars, counts = CountAlnumDistribution(`// A b c A b`)

	fmt.Printf(`%c %v`, chars, counts)
	// Output: [A b c] [2 2 1]
}

func ExampleCountCharSequence() {
	var (
		text   = `aaa abcdee ffgf`
		chars  []rune
		counts []int
	)

	chars, counts = CountCharSequence(text)

	// 'a' is not counted as 4 because its breaked by another character,
	// space ' '.
	fmt.Printf(`%c %v`, chars, counts)

	// Output:
	// [a e f] [3 2 2]
}

func ExampleCountDigit() {
	var text = `// Copyright 2018 Mhd Sulhan <ms@kilabit.info>. All rights reserved.`
	fmt.Println(CountDigit(text))
	// Output: 4
}

func ExampleCountUniqChar() {
	fmt.Println(CountUniqChar(`abc abc`))
	fmt.Println(CountUniqChar(`abc ABC`))
	// Output:
	// 4
	// 7
}

func ExampleCountUpperLower() {
	fmt.Println(CountUpperLower(`// A B C d e f g h I J K`))
	// Output: 6 5
}

func ExampleMaxCharSequence() {
	var (
		c rune
		n int
	)

	c, n = MaxCharSequence(`aaa abcdee ffgf`)

	fmt.Printf(`%c %d`, c, n)
	// Output: a 3
}

func ExampleRatioAlnum() {
	fmt.Println(RatioAlnum(`//A1`))
	// Output: 0.5
}

func ExampleRatioDigit() {
	fmt.Println(RatioDigit(`// A b 0 1`))
	// Output: 0.2
}

func ExampleRatioNonAlnum() {
	fmt.Println(RatioNonAlnum(`// A1`, false))
	fmt.Println(RatioNonAlnum(`// A1`, true))
	// Output:
	// 0.4
	// 0.6
}

func ExampleRatioUpper() {
	fmt.Println(RatioUpper(`// A b c d`))
	// Output: 0.25
}

func ExampleRatioUpperLower() {
	fmt.Println(RatioUpperLower(`// A b c d e`))
	// Output: 0.25
}

func ExampleTextSumCountTokens() {
	var (
		text   = `[[aa]] [[AA]]`
		tokens = []string{`[[`}
	)

	fmt.Println(TextSumCountTokens(text, tokens, false))

	tokens = []string{`aa`}
	fmt.Println(TextSumCountTokens(text, tokens, false))

	fmt.Println(TextSumCountTokens(text, tokens, true))

	// Output:
	// 2
	// 2
	// 1
}

func ExampleTextFrequencyOfTokens() {
	var text = `a b c d A B C D 1 2`

	fmt.Println(TextFrequencyOfTokens(text, []string{`a`}, false))
	fmt.Println(TextFrequencyOfTokens(text, []string{`a`}, true))
	// Output:
	// 0.2
	// 0.1
}
