package hunspell

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewStem(t *testing.T) {
	cases := []struct {
		desc     string
		line     string
		exp      *stem
		expError string
	}{{
		desc: "With empty line",
	}, {
		desc: "With single word",
		line: `a`,
		exp: &stem{
			value: "a",
		},
	}, {
		desc: "With single word and trailing space",
		line: `a `,
		exp: &stem{
			value: "a",
		},
	}, {
		desc: "With single word and flags",
		line: `a/bc`,
		exp: &stem{
			value:    "a",
			rawFlags: "bc",
		},
	}, {
		desc: "With single word and morpheme",
		line: `a ph:x`,
		exp: &stem{
			value:        "a",
			rawMorphemes: []string{"ph:x"},
		},
	}, {
		desc: "With single word, flags, and morphemes",
		line: `a/bc ph:x st:y`,
		exp: &stem{
			value:    "a",
			rawFlags: "bc",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With escaped slash",
		line: `a\/b`,
		exp: &stem{
			value: "a/b",
		},
	}, {
		desc: "With escaped slash and flags",
		line: `a\//bc`,
		exp: &stem{
			value:    "a/",
			rawFlags: "bc",
		},
	}, {
		desc: "With escaped slash and morphemes",
		line: `a\/ ph:x st:y`,
		exp: &stem{
			value: "a/",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With escaped slash, flags, and morphemes",
		line: `a\//bc ph:x st:y`,
		exp: &stem{
			value:    "a/",
			rawFlags: "bc",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With word pair",
		line: "a lot",
		exp: &stem{
			value: "a lot",
		},
	}, {
		desc: "With word pair and flags",
		line: "a lot/bc",
		exp: &stem{
			value:    "a lot",
			rawFlags: "bc",
		},
	}, {
		desc: "With word pair and morphemes",
		line: `a lot ph:x st:y`,
		exp: &stem{
			value: "a lot",
			rawMorphemes: []string{
				"ph:x",
				"st:y",
			},
		},
	}, {
		desc: "With word pair, flags, and morphemes",
		line: `a lot/bc ph:x st:y`,
		exp: &stem{
			value:    "a lot",
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

		got, err := newStem(c.line)
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
			"A": &affix{
				isPrefix:       true,
				isCrossProduct: true,
				rules: []*affixRule{{
					affix: "x",
				}},
			},
		},
		suffixes: map[string]*affix{
			"B": &affix{
				isCrossProduct: true,
				rules: []*affixRule{{
					affix: "y",
				}},
			},
		},
	}

	cases := []struct {
		desc           string
		in             *stem
		expError       string
		expStem        *stem
		expDerivatives []string
	}{{
		desc: "Simple prefix",
		in: &stem{
			value:    "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &stem{
			value:    "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"p:q",
			},
			morphemes: map[string][]string{
				"p": []string{"q"},
			},
		},
		expDerivatives: []string{
			"xa",
		},
	}, {
		desc: "Simple suffix",
		in: &stem{
			value:    "a",
			rawFlags: "B",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &stem{
			value:    "a",
			rawFlags: "B",
			rawMorphemes: []string{
				"p:q",
			},
			morphemes: map[string][]string{
				"p": []string{"q"},
			},
		},
		expDerivatives: []string{
			"ay",
		},
	}, {
		desc: "Simple suffix with alias",
		in: &stem{
			value:    "a",
			rawFlags: "2",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &stem{
			value:    "a",
			rawFlags: "B",
			rawMorphemes: []string{
				"p:q",
			},
			morphemes: map[string][]string{
				"p": []string{"q"},
			},
		},
		expDerivatives: []string{
			"ay",
		},
	}, {

		desc: "Prefix with alias",
		in: &stem{
			value:    "a",
			rawFlags: "1",
			rawMorphemes: []string{
				"p:q",
			},
		},
		expStem: &stem{
			value:    "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"p:q",
			},
			morphemes: map[string][]string{
				"p": []string{"q"},
			},
		},
		expDerivatives: []string{
			"xa",
		},
	}, {
		desc: "Prefix and morpheme with alias",
		in: &stem{
			value:    "a",
			rawFlags: "1",
			rawMorphemes: []string{
				"1",
			},
		},
		expStem: &stem{
			value:    "a",
			rawFlags: "A",
			rawMorphemes: []string{
				"1",
			},
			morphemes: map[string][]string{
				"p": []string{"q"},
			},
		},
		expDerivatives: []string{
			"xa",
		},
	}, {
		desc: "Prefix and suffix",
		in: &stem{
			value:    "a",
			rawFlags: "AB",
		},
		expStem: &stem{
			value:    "a",
			rawFlags: "AB",
		},
		expDerivatives: []string{
			"xa",
			"xay",
			"ay",
		},
	}, {
		desc: "Suffix and prefix",
		in: &stem{
			value:    "a",
			rawFlags: "BA",
		},
		expStem: &stem{
			value:    "a",
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

		test.Assert(t, c.desc+" derivatives", c.expDerivatives, gotDerivatives, true)
		test.Assert(t, c.desc+" after", c.expStem, c.in, true)
	}
}
