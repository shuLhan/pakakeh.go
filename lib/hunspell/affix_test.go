// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestPrefix_apply(t *testing.T) {
	isPrefix := true
	spell := &Spell{
		flag: FlagASCII,
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
		err := afx.addRule(spell, r.stripping, r.affix, r.condition,
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
		got := afx.apply(c.root)

		test.Assert(t, "Prefix.apply", c.exp, got, true)
	}
}

func TestSuffix_apply(t *testing.T) {
	isPrefix := false
	spell := &Spell{
		flag: FlagASCII,
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
		err := afx.addRule(spell, r.stripping, r.affix, r.condition,
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
		got := afx.apply(c.root)

		test.Assert(t, "Suffix.apply", c.exp, got, true)
	}
}
