// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseStem(t *testing.T) {
	cases := []struct {
		desc     string
		line     string
		exp      *Stem
		expError string
	}{{
		desc: "With empty line",
	}, {
		desc: "With single word",
		line: `a`,
		exp: &Stem{
			Word: "a",
		},
	}, {
		desc: "With single word and trailing space",
		line: `a `,
		exp: &Stem{
			Word: "a",
		},
	}, {
		desc: "With single word and flags",
		line: `a/bc`,
		exp: &Stem{
			Word:     "a",
			rawFlags: "bc",
		},
	}, {
		desc: "With single word and morpheme",
		line: `a ph:x`,
		exp: &Stem{
			Word:         "a",
			rawMorphemes: []string{"ph:x"},
		},
	}, {
		desc: "With single word, flags, and morphemes",
		line: `a/bc ph:x st:y`,
		exp: &Stem{
			Word:     "a",
			rawFlags: "bc",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With escaped slash",
		line: `a\/b`,
		exp: &Stem{
			Word: "a/b",
		},
	}, {
		desc: "With escaped slash and flags",
		line: `a\//bc`,
		exp: &Stem{
			Word:     "a/",
			rawFlags: "bc",
		},
	}, {
		desc: "With escaped slash and morphemes",
		line: `a\/ ph:x st:y`,
		exp: &Stem{
			Word: "a/",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With escaped slash, flags, and morphemes",
		line: `a\//bc ph:x st:y`,
		exp: &Stem{
			Word:     "a/",
			rawFlags: "bc",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With word pair",
		line: "a lot",
		exp: &Stem{
			Word: "a lot",
		},
	}, {
		desc: "With word pair and flags",
		line: "a lot/bc",
		exp: &Stem{
			Word:     "a lot",
			rawFlags: "bc",
		},
	}, {
		desc: "With word pair and morphemes",
		line: `a lot ph:x st:y`,
		exp: &Stem{
			Word: "a lot",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With word pair, flags, and morphemes",
		line: `a lot/bc ph:x st:y`,
		exp: &Stem{
			Word:     "a lot",
			rawFlags: "bc",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc:     "With three words",
		line:     `a lot of`,
		expError: `only one or two words allowed: "a lot of"`,
	}, {
		desc:     "With invalid escape",
		line:     `a\ b`,
		expError: `invalid escape "a\\"`,
	}, {
		desc:     "With invalid morpheme",
		line:     `a :q`,
		expError: errInvalidMorpheme(":q").Error(),
	}, {
		desc:     "With invalid morpheme (2)",
		line:     `a p:q s`,
		expError: errInvalidMorpheme("s").Error(),
	}, {
		desc:     "With invalid morpheme (2)",
		line:     `a p:q :s`,
		expError: errInvalidMorpheme(":s").Error(),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := parseStem(c.line)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}

		test.Assert(t, "stem", c.exp, got, true)
	}
}

func TestStem_unpack(t *testing.T) {
	opts := &affixOptions{
		flag: DefaultFlag,
		afAliases: []string{
			"",
			"A",
			"B",
			"AB",
		},
		amAliases: []string{
			"",
			"p:q",
			"p:q r:s",
		},
		prefixes: map[string]*affix{
			"A": {
				isPrefix:       true,
				isCrossProduct: true,
				rules: []*affixRule{{
					affix: "x",
				}},
			},
		},
		suffixes: map[string]*affix{
			"B": {
				isCrossProduct: true,
				rules: []*affixRule{{
					affix: "y",
				}},
			},
		},
	}

	cases := []struct {
		desc           string
		in             *Stem
		expError       string
		expStem        *Stem
		expDerivatives []string
	}{{
		desc: "Simple prefix",
		in: &Stem{
			Word:     "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &Stem{
			Word:     "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"p:q",
			},
			Morphemes: Morphemes{
				"p": "q",
			},
		},
		expDerivatives: []string{
			"xa",
		},
	}, {
		desc: "Simple suffix",
		in: &Stem{
			Word:     "a",
			rawFlags: "B",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &Stem{
			Word:     "a",
			rawFlags: "B",
			rawMorphemes: []string{
				"p:q",
			},
			Morphemes: Morphemes{
				"p": "q",
			},
		},
		expDerivatives: []string{
			"ay",
		},
	}, {
		desc: "Simple suffix with alias",
		in: &Stem{
			Word:     "a",
			rawFlags: "2",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &Stem{
			Word:     "a",
			rawFlags: "B",
			rawMorphemes: []string{
				"p:q",
			},
			Morphemes: Morphemes{
				"p": "q",
			},
		},
		expDerivatives: []string{
			"ay",
		},
	}, {

		desc: "Prefix with alias",
		in: &Stem{
			Word:     "a",
			rawFlags: "1",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &Stem{
			Word:     "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"p:q",
			},
			Morphemes: Morphemes{
				"p": "q",
			},
		},
		expDerivatives: []string{
			"xa",
		},
	}, {
		desc: "Prefix and morpheme with alias",
		in: &Stem{
			Word:     "a",
			rawFlags: "1",
			rawMorphemes: []string{
				"1",
			},
		},
		expStem: &Stem{
			Word:     "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"1",
			},
			Morphemes: Morphemes{
				"p": "q",
			},
		},
		expDerivatives: []string{
			"xa",
		},
	}, {
		desc: "Prefix and suffix",
		in: &Stem{
			Word:     "a",
			rawFlags: "AB",
		},
		expStem: &Stem{
			Word:     "a",
			rawFlags: "AB",
		},
		expDerivatives: []string{
			"xa",
			"xay",
			"ay",
		},
	}, {
		desc: "Suffix and prefix",
		in: &Stem{
			Word:     "a",
			rawFlags: "BA",
		},
		expStem: &Stem{
			Word:     "a",
			rawFlags: "BA",
		},
		expDerivatives: []string{
			"ay",
			"xa",
			"xay",
		},
	}}

	for _, c := range cases {
		gotDerivatives, err := c.in.unpack(opts)
		if err != nil {
			test.Assert(t, "unpack error", c.expError, err.Error(), true)
		}

		got := make([]string, 0, len(gotDerivatives))
		for _, der := range gotDerivatives {
			got = append(got, der.Word)
		}

		test.Assert(t, c.desc+" derivatives", c.expDerivatives, got, true)
		test.Assert(t, c.desc+" after", c.expStem, c.in, true)
	}
}
