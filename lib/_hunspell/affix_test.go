// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package hunspell

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestPrefix_apply(t *testing.T) {
	isPrefix := true
	spell := &Spell{
		opts: affixOptions{
			flag: FlagASCII,
		},
	}

	rawrules := []struct {
		stripping, affix, condition string
		morphemes                   []string
	}{{
		"0", "b", "[^a]", nil, // Add prefix "b" if word does not start with "a"
	}, {
		"0", "c", ".", nil, // Add prefix "c"
	}}

	afx := newAffix("P", isPrefix, true, len(rawrules))

	for _, r := range rawrules {
		err := afx.addRule(&spell.opts, r.stripping, r.affix, r.condition,
			r.morphemes)
		if err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		root string
		exp  []string
	}{{
		root: "a",
		exp:  []string{"ca"},
	}, {
		root: "z",
		exp:  []string{"bz", "cz"},
	}}

	for _, c := range cases {
		root := newStem(nil, c.root, nil)

		stems := afx.apply(root)

		got := make([]string, 0, len(stems))
		for _, stem := range stems {
			got = append(got, stem.Word)
		}

		test.Assert(t, "Prefix.apply", c.exp, got)
	}
}

func TestSuffix_apply(t *testing.T) {
	isPrefix := false
	spell := &Spell{
		opts: affixOptions{
			flag: FlagASCII,
		},
	}

	rawrules := []struct {
		stripping, affix, condition string
		morphemes                   []string
	}{{
		"0", "b", "[^a]", nil, // Add suffix "b" if word does not end with "a".
	}, {
		"0", "c", ".", nil, // Add suffix "c".
	}}

	afx := newAffix("S", isPrefix, true, len(rawrules))

	for _, r := range rawrules {
		err := afx.addRule(&spell.opts, r.stripping, r.affix, r.condition,
			r.morphemes)
		if err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		root string
		exp  []string
	}{{
		root: "a",
		exp:  []string{"ac"},
	}, {
		root: "z",
		exp:  []string{"zb", "zc"},
	}}

	for _, c := range cases {
		root := newStem(nil, c.root, nil)

		stems := afx.apply(root)

		got := make([]string, 0, len(stems))
		for _, stem := range stems {
			got = append(got, stem.Word)
		}

		test.Assert(t, "Suffix.apply", c.exp, got)
	}
}
