// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"

	"github.com/shuLhan/share/lib/numbers"
	"github.com/shuLhan/share/lib/test"
)

func TestCountMissRate(t *testing.T) {
	cases := []struct {
		src    []string
		target []string
		exp    float64
	}{{
		// Empty.
	}, {
		src:    []string{"A", "B", "C", "D"},
		target: []string{"A", "B", "C"},
		exp:    0,
	}, {
		src:    []string{"A", "B", "C", "D"},
		target: []string{"A", "B", "C", "D"},
		exp:    0,
	}, {
		src:    []string{"A", "B", "C", "D"},
		target: []string{"B", "B", "C", "D"},
		exp:    0.25,
	}, {
		src:    []string{"A", "B", "C", "D"},
		target: []string{"B", "C", "C", "D"},
		exp:    0.5,
	}, {
		src:    []string{"A", "B", "C", "D"},
		target: []string{"B", "C", "D", "D"},
		exp:    0.75,
	}, {
		src:    []string{"A", "B", "C", "D"},
		target: []string{"C", "D", "D", "E"},
		exp:    1.0,
	}}

	for _, c := range cases {
		got, _, _ := CountMissRate(c.src, c.target)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestCountTokens(t *testing.T) {
	cases := []struct {
		words     []string
		tokens    []string
		sensitive bool
		exp       []int
	}{{
		// Empty.
	}, {
		words:  []string{"A", "B", "C", "a", "b", "c"},
		tokens: []string{"A", "B"},
		exp:    []int{2, 2},
	}, {
		words:     []string{"A", "B", "C", "a", "b", "c"},
		tokens:    []string{"A", "B"},
		sensitive: true,
		exp:       []int{1, 1},
	}}

	for _, c := range cases {
		got := CountTokens(c.words, c.tokens, c.sensitive)

		test.Assert(t, "", c.exp, got, true)
	}
}

func TestDelete(t *testing.T) {
	value := "b"
	exp := []string{"a", "c"}

	cases := []struct {
		in []string
	}{{
		in: []string{"b", "a", "c"},
	}, {
		in: []string{"a", "b", "c"},
	}, {
		in: []string{"a", "c", "b"},
	}}

	for _, c := range cases {
		var ok bool
		c.in, ok = Delete(c.in, value)
		test.Assert(t, "Delete OK?", true, ok, true)
		test.Assert(t, "Delete", exp, c.in, true)
	}
}

func TestFrequencyOfTokens(t *testing.T) {
	cases := []struct {
		words, tokens []string
		exp           []float64
		sensitive     bool
	}{{
		// Empty.
	}, {
		words:  []string{"a", "b", "a", "b", "a", "c"},
		tokens: []string{"a", "b"},
		exp:    []float64{0.5, 0.3333333333333333},
	}}

	for _, c := range cases {
		got := FrequencyOfTokens(c.words, c.tokens, c.sensitive)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestIsContain(t *testing.T) {
	ss := []string{"a", "b", "c", "d"}

	got := IsContain(ss, "a")
	test.Assert(t, "", true, got, true)

	got = IsContain(ss, "e")
	test.Assert(t, "", false, got, true)
}

func TestIsEqual(t *testing.T) {
	cases := []struct {
		a, b []string
		exp  bool
	}{{
		a:   []string{"a", "b"},
		b:   []string{"a", "b"},
		exp: true,
	}, {
		a:   []string{"a", "b"},
		b:   []string{"b", "a"},
		exp: true,
	}, {
		a: []string{"a", "b"},
		b: []string{"a"},
	}, {
		a: []string{"a"},
		b: []string{"b", "a"},
	}, {
		a: []string{"a", "b"},
		b: []string{"a", "c"},
	}}

	for _, c := range cases {
		test.Assert(t, "", c.exp, IsEqual(c.a, c.b), true)
	}
}

func TestLongest(t *testing.T) {
	cases := []struct {
		words  []string
		exp    string
		expIdx int
	}{{
		// Empty.
		expIdx: -1,
	}, {
		words:  []string{"a", "bb", "ccc", "d", "eee"},
		exp:    "ccc",
		expIdx: 2,
	}, {
		words:  []string{"a", "bb", "ccc", "dddd", "eee"},
		exp:    "dddd",
		expIdx: 3,
	}}

	for _, c := range cases {
		got, idx := Longest(c.words)

		test.Assert(t, "word", c.exp, got, true)
		test.Assert(t, "idx", c.expIdx, idx, true)
	}
}

func TestMostFrequentTokens(t *testing.T) {
	cases := []struct {
		words, tokens []string
		sensitive     bool
		exp           string
	}{{
		// Empty
	}, {
		words:  []string{"a", "b", "A"},
		tokens: []string{"a", "b"},
		exp:    "a",
	}, {
		words:     []string{"a", "b", "A", "b"},
		tokens:    []string{"a", "b"},
		sensitive: true,
		exp:       "b",
	}, {
		words:  []string{"a", "b", "A", "B"},
		tokens: []string{"a", "b"},
		exp:    "a",
	}}

	for _, c := range cases {
		got := MostFrequentTokens(c.words, c.tokens, c.sensitive)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestSortByIndex(t *testing.T) {
	dat := []string{"Z", "X", "C", "V", "B", "N", "M"}
	exp := []string{"B", "C", "M", "N", "V", "X", "Z"}
	ids := []int{4, 2, 6, 5, 3, 1, 0}

	SortByIndex(&dat, ids)

	test.Assert(t, "", exp, dat, true)
}

func TestSwap(t *testing.T) {
	ss := []string{"a", "b", "c"}

	cases := []struct {
		x, y int
		exp  []string
	}{{
		x:   -1,
		exp: []string{"a", "b", "c"},
	}, {
		y:   -1,
		exp: []string{"a", "b", "c"},
	}, {
		x:   4,
		exp: []string{"a", "b", "c"},
	}, {
		y:   4,
		exp: []string{"a", "b", "c"},
	}, {
		x:   1,
		y:   1,
		exp: []string{"a", "b", "c"},
	}, {
		x:   1,
		y:   2,
		exp: []string{"a", "c", "b"},
	}}
	for _, c := range cases {
		Swap(ss, c.x, c.y)
		test.Assert(t, "", c.exp, ss, true)
	}
}

func TestTotalFrequencyOfTokens(t *testing.T) {
	cases := []struct {
		words, tokens []string
		sensitive     bool
		exp           float64
	}{{
		// Empty.
	}, {
		words:  []string{"a", "b", "a", "b", "a", "c"},
		tokens: []string{"a", "b"},
		exp:    numbers.Float64Round((3.0/6)+(2.0/6), 3),
	}}

	for _, c := range cases {
		got := TotalFrequencyOfTokens(c.words, c.tokens, c.sensitive)

		test.Assert(t, "", c.exp, numbers.Float64Round(got, 3), true)
	}
}

func TestUniq(t *testing.T) {
	cases := []struct {
		words     []string
		sensitive bool
		expReturn []string
		expWords  []string
	}{{
		words:     []string{"a", "A"},
		sensitive: true,
		expReturn: []string{"a", "A"},
		expWords:  []string{"a", "A"},
	}, {
		words:     []string{"a", "A"},
		expReturn: []string{"a"},
		expWords:  []string{"a", ""},
	}}

	for _, c := range cases {
		got := Uniq(c.words, c.sensitive)
		test.Assert(t, "unique", c.expReturn, got, true)
		test.Assert(t, "words", c.expWords, c.words, true)
	}
}
