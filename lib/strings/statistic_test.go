// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestCountAlnum(t *testing.T) {
	cases := []struct {
		text string
		exp  int
	}{{
		// Empty
	}, {
		text: "// 123",
		exp:  3,
	}, {
		text: "// A B C",
		exp:  3,
	}, {
		text: "// A b c 1 2 3",
		exp:  6,
	}}

	for _, c := range cases {
		got := CountAlnum(c.text)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestCountAlnumDistribution(t *testing.T) {
	cases := []struct {
		text      string
		expChars  []rune
		expCounts []int
	}{{
		// Empty
	}, {
		text:      "// 123",
		expChars:  []rune{'1', '2', '3'},
		expCounts: []int{1, 1, 1},
	}, {
		text:      "// A B C",
		expChars:  []rune{'A', 'B', 'C'},
		expCounts: []int{1, 1, 1},
	}, {
		text:      "// A B C A B C",
		expChars:  []rune{'A', 'B', 'C'},
		expCounts: []int{2, 2, 2},
	}}

	for _, c := range cases {
		gotChars, gotCounts := CountAlnumDistribution(c.text)
		test.Assert(t, "chars", c.expChars, gotChars, true)
		test.Assert(t, "counts", c.expCounts, gotCounts, true)
	}
}

func TestCountCharSequence(t *testing.T) {
	cases := []struct {
		text      string
		expChars  []rune
		expCounts []int
	}{{
		text:      "// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.",
		expChars:  []rune{'/', 'l'},
		expCounts: []int{2, 2},
	}, {
		text: "Use of this source code is governed by a BSD-style",
	}, {
		text:      "aaa abcdee ffgf",
		expChars:  []rune{'a', 'e', 'f'},
		expCounts: []int{3, 2, 2},
	}, {
		text:      " |  image name          = {{legend|#0080FF|Areas affected by flooding}}{{legend|#002255|Death(s) affected by flooding}}{{legend|#C83737|Areas affected by flooding and strong winds}}{{legend|#550000|Death(s) affected by flooding and strong winds}}",
		expChars:  []rune{'{', '0', 'F', 'f', 'o', '}', '{', '0', '2', '5', 'f', 'o', '}', '{', 'f', 'o', '}', '{', '5', '0', 'f', 'o', '}'},
		expCounts: []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 4, 2, 2, 2},
	}}

	for _, c := range cases {
		gotChars, gotCounts := CountCharSequence(c.text)

		test.Assert(t, "", c.expChars, gotChars, true)
		test.Assert(t, "", c.expCounts, gotCounts, true)
	}
}

func TestCountDigit(t *testing.T) {
	cases := []struct {
		text string
		exp  int
	}{{
		// Empty.
	}, {
		text: "// Copyright 2018 Mhd Sulhan <ms@kilabit.info>. All rights reserved.",
		exp:  4,
	}}

	for _, c := range cases {
		got := CountDigit(c.text)

		test.Assert(t, "", c.exp, got, true)
	}
}

func TestCountNonAlnum(t *testing.T) {
	cases := []struct {
		text      string
		withspace bool
		exp       int
	}{{
		// Empty
	}, {
		text: "// 123",
		exp:  2,
	}, {
		text:      "// 123",
		withspace: true,
		exp:       3,
	}, {
		text: "// A B C",
		exp:  2,
	}, {
		text:      "// A B C",
		withspace: true,
		exp:       5,
	}, {
		text: "// A b c 1 2 3",
		exp:  2,
	}, {
		text:      "// A b c 1 2 3",
		withspace: true,
		exp:       8,
	}}

	for _, c := range cases {
		got := CountNonAlnum(c.text, c.withspace)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestCountUniqChar(t *testing.T) {
	cases := []struct {
		text string
		exp  int
	}{{
		// Empty.
	}, {
		text: "abc abc",
		exp:  4,
	}, {
		text: "abc ABC",
		exp:  7,
	}}

	for _, c := range cases {
		got := CountUniqChar(c.text)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestCountUpperLower(t *testing.T) {
	cases := []struct {
		text     string
		expUpper int
		expLower int
	}{{
		text:     "// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.",
		expUpper: 4,
		expLower: 44,
	}}

	for _, c := range cases {
		gotup, gotlo := CountUpperLower(c.text)

		test.Assert(t, "", c.expUpper, gotup, true)
		test.Assert(t, "", c.expLower, gotlo, true)
	}
}

func TestMaxCharSequence(t *testing.T) {
	cases := []struct {
		text  string
		char  rune
		count int
	}{{
		text:  "// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.",
		char:  '/',
		count: 2,
	}, {
		text: "Use of this source code is governed by a BSD-style",
	}, {
		text:  "aaa abcdee ffgf",
		char:  'a',
		count: 3,
	}, {
		text:  " |  image name          = {{legend|#0080FF|Areas affected by flooding}}{{legend|#002255|Death(s) affected by flooding}}{{legend|#C83737|Areas affected by flooding and strong winds}}{{legend|#550000|Death(s) affected by flooding and strong winds}}",
		char:  '0',
		count: 4,
	}}

	for _, c := range cases {
		gotv, gotc := MaxCharSequence(c.text)

		test.Assert(t, "", c.char, gotv, true)
		test.Assert(t, "", c.count, gotc, true)
	}
}

func TestRatioAlnum(t *testing.T) {
	cases := []struct {
		text string
		exp  float64
	}{{
		// Empty.
	}, {
		text: "// A b c d",
		exp:  0.4,
	}, {
		text: "// A123b",
		exp:  0.625,
	}}

	for _, c := range cases {
		got := RatioAlnum(c.text)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestRatioDigit(t *testing.T) {
	cases := []struct {
		text string
		exp  float64
	}{{
		// Empty.
	}, {
		text: "// A b c d",
		exp:  0,
	}, {
		text: "// A123b",
		exp:  0.375,
	}}

	for _, c := range cases {
		got := RatioDigit(c.text)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestRatioNonAlnum(t *testing.T) {
	cases := []struct {
		text      string
		withspace bool
		exp       float64
	}{{
		// Empty.
	}, {
		text: "// A b c d",
		exp:  0.2,
	}, {
		text:      "// A b c d",
		withspace: true,
		exp:       0.6,
	}, {
		text: "// A123b",
		exp:  0.25,
	}, {
		text:      "// A123b",
		withspace: true,
		exp:       0.375,
	}}

	for _, c := range cases {
		got := RatioNonAlnum(c.text, c.withspace)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestRatioUpper(t *testing.T) {
	cases := []struct {
		text string
		exp  float64
	}{{
		// Empty.
	}, {
		text: "// A b c d",
		exp:  0.25,
	}}

	for _, c := range cases {
		got := RatioUpper(c.text)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestRatioUpperLower(t *testing.T) {
	cases := []struct {
		text string
		exp  float64
	}{{
		// Empty
	}, {
		text: "// 134234",
	}, {
		text: "// A B C",
		exp:  3,
	}, {
		text: "// A b c d e",
		exp:  0.25,
	}}

	for _, c := range cases {
		got := RatioUpperLower(c.text)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestTextSumCountTokens(t *testing.T) {
	cases := []struct {
		text      string
		tokens    []string
		sensitive bool
		exp       int
	}{{
		// Empty.
	}, {
		text:   "[[aa]] [[AA]]",
		tokens: []string{"[["},
		exp:    2,
	}, {
		text:   "[[aa]] [[AA]]",
		tokens: []string{"]]"},
		exp:    2,
	}, {
		text:   "[[aa]] [[AA]]",
		tokens: []string{"[[", "]]"},
		exp:    4,
	}, {
		text:   "[[aa]] [[AA]]",
		tokens: []string{"aa"},
		exp:    2,
	}, {
		text:      "[[aa]] [[AA]]",
		tokens:    []string{"aa"},
		sensitive: true,
		exp:       1,
	}}

	for _, c := range cases {
		got := TextSumCountTokens(c.text, c.tokens, c.sensitive)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestTextFrequencyOfTokens(t *testing.T) {
	cases := []struct {
		text      string
		tokens    []string
		sensitive bool
		exp       float64
	}{{
		// Empty.
	}, {
		text:   "a b c d A B C D",
		tokens: []string{"a"},
		exp:    0.25,
	}, {
		text:      "a b c d A B C D",
		tokens:    []string{"a"},
		sensitive: true,
		exp:       0.125,
	}}

	for _, c := range cases {
		got := TextFrequencyOfTokens(c.text, c.tokens, c.sensitive)
		test.Assert(t, "", c.exp, got, true)
	}

}
