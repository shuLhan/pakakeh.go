// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package strings

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestCountAlnum(t *testing.T) {
	type testCase struct {
		text string
		exp  int
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text: `// 123`,
		exp:  3,
	}, {
		text: `// A B C`,
		exp:  3,
	}, {
		text: `// A b c 1 2 3`,
		exp:  6,
	}}

	var (
		c   testCase
		got int
	)
	for _, c = range cases {
		got = CountAlnum(c.text)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestCountAlnumDistribution(t *testing.T) {
	type testCase struct {
		text      string
		expChars  []rune
		expCounts []int
	}

	var cases = []testCase{{
		// Empty input.
	}, {
		text:      `// 123`,
		expChars:  []rune{'1', '2', '3'},
		expCounts: []int{1, 1, 1},
	}, {
		text:      `// A B C`,
		expChars:  []rune{'A', 'B', 'C'},
		expCounts: []int{1, 1, 1},
	}, {
		text:      `// A B C A B C`,
		expChars:  []rune{'A', 'B', 'C'},
		expCounts: []int{2, 2, 2},
	}}

	var (
		c         testCase
		gotChars  []rune
		gotCounts []int
	)
	for _, c = range cases {
		gotChars, gotCounts = CountAlnumDistribution(c.text)
		test.Assert(t, `chars`, c.expChars, gotChars)
		test.Assert(t, `counts`, c.expCounts, gotCounts)
	}
}

func TestCountCharSequence(t *testing.T) {
	type testCase struct {
		text      string
		expChars  []rune
		expCounts []int
	}
	var cases = []testCase{{
		text:      `// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.`,
		expChars:  []rune{'/', 'l'},
		expCounts: []int{2, 2},
	}, {
		text: `Use of this source code is governed by a BSD-style`,
	}, {
		text:      `aaa abcdee ffgf`,
		expChars:  []rune{'a', 'e', 'f'},
		expCounts: []int{3, 2, 2},
	}, {
		text: ` |  image name          = {{legend|#0080FF|Areas affected by flooding}}{{legend|#002255|Death(s) affected by flooding}}{{legend|#C83737|Areas affected by flooding and strong winds}}{{legend|#550000|Death(s) affected by flooding and strong winds}}`,
		expChars: []rune{
			'{', '0', 'F', 'f', 'o',
			'}', '{', '0', '2', '5',
			'f', 'o', '}', '{', 'f',
			'o', '}', '{', '5', '0',
			'f', 'o', '}',
		},
		expCounts: []int{
			2, 2, 2, 2, 2,
			2, 2, 2, 2, 2,
			2, 2, 2, 2, 2,
			2, 2, 2,
			2, 4, 2, 2, 2,
		},
	}}

	var (
		c         testCase
		gotChars  []rune
		gotCounts []int
	)
	for _, c = range cases {
		gotChars, gotCounts = CountCharSequence(c.text)

		test.Assert(t, ``, c.expChars, gotChars)
		test.Assert(t, ``, c.expCounts, gotCounts)
	}
}

func TestCountDigit(t *testing.T) {
	type testCase struct {
		text string
		exp  int
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text: `// 2018 `,
		exp:  4,
	}}

	var (
		c   testCase
		got int
	)
	for _, c = range cases {
		got = CountDigit(c.text)

		test.Assert(t, ``, c.exp, got)
	}
}

func TestCountNonAlnum(t *testing.T) {
	type testCase struct {
		text      string
		exp       int
		withspace bool
	}
	var cases = []testCase{{
		// Empty.
	}, {
		text: `// 123`,
		exp:  2,
	}, {
		text:      `// 123`,
		withspace: true,
		exp:       3,
	}, {
		text: `// A B C`,
		exp:  2,
	}, {
		text:      `// A B C`,
		withspace: true,
		exp:       5,
	}, {
		text: `// A b c 1 2 3`,
		exp:  2,
	}, {
		text:      `// A b c 1 2 3`,
		withspace: true,
		exp:       8,
	}}

	var (
		c   testCase
		got int
	)
	for _, c = range cases {
		got = CountNonAlnum(c.text, c.withspace)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestCountUniqChar(t *testing.T) {
	type testCase struct {
		text string
		exp  int
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text: `abc abc`,
		exp:  4,
	}, {
		text: `abc ABC`,
		exp:  7,
	}}

	var (
		c   testCase
		got int
	)
	for _, c = range cases {
		got = CountUniqChar(c.text)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestCountUpperLower(t *testing.T) {
	type testCase struct {
		text     string
		expUpper int
		expLower int
	}

	var cases = []testCase{{
		text:     `// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.`,
		expUpper: 4,
		expLower: 44,
	}}

	var (
		c     testCase
		gotup int
		gotlo int
	)
	for _, c = range cases {
		gotup, gotlo = CountUpperLower(c.text)

		test.Assert(t, ``, c.expUpper, gotup)
		test.Assert(t, ``, c.expLower, gotlo)
	}
}

func TestMaxCharSequence(t *testing.T) {
	type testCase struct {
		text  string
		char  rune
		count int
	}

	var cases = []testCase{{
		text:  `// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.`,
		char:  '/',
		count: 2,
	}, {
		text: `Use of this source code is governed by a BSD-style`,
	}, {
		text:  `aaa abcdee ffgf`,
		char:  'a',
		count: 3,
	}, {
		text:  ` |  image name          = {{legend|#0080FF|Areas affected by flooding}}{{legend|#002255|Death(s) affected by flooding}}{{legend|#C83737|Areas affected by flooding and strong winds}}{{legend|#550000|Death(s) affected by flooding and strong winds}}`,
		char:  '0',
		count: 4,
	}}

	var (
		c    testCase
		gotv rune
		gotc int
	)
	for _, c = range cases {
		gotv, gotc = MaxCharSequence(c.text)

		test.Assert(t, ``, c.char, gotv)
		test.Assert(t, ``, c.count, gotc)
	}
}

func TestRatioAlnum(t *testing.T) {
	type testCase struct {
		text string
		exp  float64
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text: `// A b c d`,
		exp:  0.4,
	}, {
		text: `// A123b`,
		exp:  0.625,
	}}

	var (
		c   testCase
		got float64
	)
	for _, c = range cases {
		got = RatioAlnum(c.text)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestRatioDigit(t *testing.T) {
	type testCase struct {
		text string
		exp  float64
	}
	var cases = []testCase{{
		// Empty.
	}, {
		text: `// A b c d`,
		exp:  0,
	}, {
		text: `// A123b`,
		exp:  0.375,
	}}

	var (
		c   testCase
		got float64
	)
	for _, c = range cases {
		got = RatioDigit(c.text)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestRatioNonAlnum(t *testing.T) {
	type testCase struct {
		text      string
		exp       float64
		withspace bool
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text: `// A b c d`,
		exp:  0.2,
	}, {
		text:      `// A b c d`,
		withspace: true,
		exp:       0.6,
	}, {
		text: `// A123b`,
		exp:  0.25,
	}, {
		text:      `// A123b`,
		withspace: true,
		exp:       0.375,
	}}

	var (
		c   testCase
		got float64
	)
	for _, c = range cases {
		got = RatioNonAlnum(c.text, c.withspace)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestRatioUpper(t *testing.T) {
	type testCase struct {
		text string
		exp  float64
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text: `// A b c d`,
		exp:  0.25,
	}}

	var (
		c   testCase
		got float64
	)
	for _, c = range cases {
		got = RatioUpper(c.text)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestRatioUpperLower(t *testing.T) {
	type testCase struct {
		text string
		exp  float64
	}
	var cases = []testCase{{
		// Empty.
	}, {
		text: `// 134234`,
	}, {
		text: `// A B C`,
		exp:  3,
	}, {
		text: `// A b c d e`,
		exp:  0.25,
	}}

	var (
		c   testCase
		got float64
	)
	for _, c = range cases {
		got = RatioUpperLower(c.text)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestTextSumCountTokens(t *testing.T) {
	type testCase struct {
		text      string
		tokens    []string
		exp       int
		sensitive bool
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text:   `[[aa]] [[AA]]`,
		tokens: []string{`[[`},
		exp:    2,
	}, {
		text:   `[[aa]] [[AA]]`,
		tokens: []string{`]]`},
		exp:    2,
	}, {
		text:   `[[aa]] [[AA]]`,
		tokens: []string{`[[`, `]]`},
		exp:    4,
	}, {
		text:   `[[aa]] [[AA]]`,
		tokens: []string{`aa`},
		exp:    2,
	}, {
		text:      `[[aa]] [[AA]]`,
		tokens:    []string{`aa`},
		sensitive: true,
		exp:       1,
	}}

	var (
		c   testCase
		got int
	)
	for _, c = range cases {
		got = TextSumCountTokens(c.text, c.tokens, c.sensitive)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestTextFrequencyOfTokens(t *testing.T) {
	type testCase struct {
		text      string
		tokens    []string
		sensitive bool
		exp       float64
	}

	var cases = []testCase{{
		// Empty.
	}, {
		text:   `a b c d A B C D`,
		tokens: []string{`a`},
		exp:    0.25,
	}, {
		text:      `a b c d A B C D`,
		tokens:    []string{`a`},
		sensitive: true,
		exp:       0.125,
	}}

	var (
		c   testCase
		got float64
	)
	for _, c = range cases {
		got = TextFrequencyOfTokens(c.text, c.tokens, c.sensitive)
		test.Assert(t, ``, c.exp, got)
	}
}
