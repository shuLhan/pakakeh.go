// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"
)

func ExampleAppendUniq() {
	var (
		in   = []string{`a`, ``, `b`, `c`}
		vals = []string{`b`, ``, `C`, `d`}
	)

	in = AppendUniq(in, vals...)

	fmt.Println(in)
	// Output:
	// [a  b c C d]
}

func ExampleCountMissRate() {
	var (
		src = []string{`A`, `B`, `C`, `D`}
		tgt = []string{`A`, `B`, `C`, `D`}
	)

	fmt.Println(CountMissRate(src, tgt))

	src = []string{`A`, `B`, `C`, `D`}
	tgt = []string{`B`, `B`, `C`, `D`}
	fmt.Println(CountMissRate(src, tgt))

	src = []string{`A`, `B`, `C`, `D`}
	tgt = []string{`B`, `C`, `C`, `D`}
	fmt.Println(CountMissRate(src, tgt))

	src = []string{`A`, `B`, `C`, `D`}
	tgt = []string{`B`, `C`, `D`, `D`}
	fmt.Println(CountMissRate(src, tgt))

	src = []string{`A`, `B`, `C`, `D`}
	tgt = []string{`C`, `D`, `D`, `E`}
	fmt.Println(CountMissRate(src, tgt))

	// Output:
	// 0 0 4
	// 0.25 1 4
	// 0.5 2 4
	// 0.75 3 4
	// 1 4 4
}

func ExampleCountToken() {
	var (
		words = []string{`A`, `B`, `C`, `a`, `b`, `c`}
	)

	fmt.Println(CountToken(words, `C`, false))
	fmt.Println(CountToken(words, `C`, true))
	// Output:
	// 2
	// 1
}

func ExampleCountTokens() {
	var (
		words  = []string{`A`, `B`, `C`, `a`, `b`, `c`}
		tokens = []string{`A`, `B`}
	)

	fmt.Println(CountTokens(words, tokens, false))
	fmt.Println(CountTokens(words, tokens, true))
	// Output:
	// [2 2]
	// [1 1]
}

func ExampleFrequencyOfToken() {
	var (
		words = []string{`A`, `B`, `C`, `a`, `b`, `c`}
	)

	fmt.Println(FrequencyOfToken(words, `C`, false))
	fmt.Println(FrequencyOfToken(words, `C`, true))
	// Output:
	// 0.3333333333333333
	// 0.16666666666666666
}

func ExampleFrequencyOfTokens() {
	var (
		words  = []string{`A`, `B`, `C`, `a`, `b`, `c`}
		tokens = []string{`A`, `B`}
	)

	fmt.Println(FrequencyOfTokens(words, tokens, false))
	fmt.Println(FrequencyOfTokens(words, tokens, true))
	// Output:
	// [0.3333333333333333 0.3333333333333333]
	// [0.16666666666666666 0.16666666666666666]
}

func ExampleIsEqual() {
	fmt.Println(IsEqual([]string{`a`, `b`}, []string{`a`, `b`}))
	fmt.Println(IsEqual([]string{`a`, `b`}, []string{`b`, `a`}))
	fmt.Println(IsEqual([]string{`a`, `b`}, []string{`a`}))
	fmt.Println(IsEqual([]string{`a`, `b`}, []string{`b`, `b`}))
	// Output:
	// true
	// true
	// false
	// false
}

func ExampleLongest() {
	var (
		words = []string{`a`, `bb`, `ccc`, `d`, `eee`}
	)

	fmt.Println(Longest(words))
	// Output: ccc 2
}

func ExampleMostFrequentTokens() {
	var (
		words  = []string{`a`, `b`, `B`, `B`, `a`}
		tokens = []string{`a`, `b`}
	)

	fmt.Println(MostFrequentTokens(words, tokens, false))
	fmt.Println(MostFrequentTokens(words, tokens, true))
	// Output:
	// b
	// a
}

func ExampleSortByIndex() {
	var (
		dat = []string{`Z`, `X`, `C`, `V`, `B`, `N`, `M`}
		ids = []int{4, 2, 6, 5, 3, 1, 0}
	)

	fmt.Println(dat)
	SortByIndex(&dat, ids)
	fmt.Println(dat)
	// Output:
	// [Z X C V B N M]
	// [B C M N V X Z]
}

func ExampleSwap() {
	var (
		ss = []string{`a`, `b`, `c`}
	)

	Swap(ss, -1, 1)
	fmt.Println(ss)
	Swap(ss, 1, -1)
	fmt.Println(ss)
	Swap(ss, 4, 1)
	fmt.Println(ss)
	Swap(ss, 1, 4)
	fmt.Println(ss)
	Swap(ss, 1, 2)
	fmt.Println(ss)
	// Output:
	// [a b c]
	// [a b c]
	// [a b c]
	// [a b c]
	// [a c b]
}

func ExampleTotalFrequencyOfTokens() {
	var (
		words  = []string{`A`, `B`, `C`, `a`, `b`, `c`}
		tokens = []string{`A`, `B`}
	)

	fmt.Println(TotalFrequencyOfTokens(words, tokens, false))
	fmt.Println(TotalFrequencyOfTokens(words, tokens, true))
	// Output:
	// 0.6666666666666666
	// 0.3333333333333333
}

func ExampleUniq() {
	var (
		words = []string{`a`, ``, `A`}
	)

	fmt.Printf("%s %s\n", Uniq(words, false), words)
	words = []string{`a`, ``, `A`}
	fmt.Printf("%s %s\n", Uniq(words, true), words)
	// Output:
	// [a] [a  ]
	// [a A] [a  A]
}
