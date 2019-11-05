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
		desc: "With single word and flags",
		line: `a/bc`,
		exp: &stem{
			value: "a",
			flags: "bc",
		},
	}, {
		desc: "With single word and morpheme",
		line: `a ph:x`,
		exp: &stem{
			value: "a",
			morphemes: map[string][]string{
				"ph": {"x"},
			},
		},
	}, {
		desc: "With single word, flags, and morphemes",
		line: `a/bc ph:x st:y`,
		exp: &stem{
			value: "a",
			flags: "bc",
			morphemes: map[string][]string{
				"ph": {"x"},
				"st": {"y"},
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
			value: "a/",
			flags: "bc",
		},
	}, {
		desc: "With escaped slash and morphemes",
		line: `a\/ ph:x st:y`,
		exp: &stem{
			value: "a/",
			morphemes: map[string][]string{
				"ph": {"x"},
				"st": {"y"},
			},
		},
	}, {
		desc: "With escaped slash, flags, and morphemes",
		line: `a\//bc ph:x st:y`,
		exp: &stem{
			value: "a/",
			flags: "bc",
			morphemes: map[string][]string{
				"ph": {"x"},
				"st": {"y"},
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
			value: "a lot",
			flags: "bc",
		},
	}, {
		desc: "With word pair and morphemes",
		line: `a lot ph:x st:y`,
		exp: &stem{
			value: "a lot",
			morphemes: map[string][]string{
				"ph": {"x"},
				"st": {"y"},
			},
		},
	}, {
		desc: "With word pair, flags, and morphemes",
		line: `a lot/bc ph:x st:y`,
		exp: &stem{
			value: "a lot",
			flags: "bc",
			morphemes: map[string][]string{
				"ph": {"x"},
				"st": {"y"},
			},
		},
	}, {
		desc:     "With three words",
		line:     `a lot of`,
		expError: `only one or two words allowed: "a lot of"`,
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
