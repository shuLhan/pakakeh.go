// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"regexp"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewAffixRule_prefix(t *testing.T) {
	var isPrefix = true
	var spell = &Spell{
		opts: affixOptions{
			flag: FlagASCII,
		},
	}

	cases := []struct {
		exp       *affixRule
		stripping string
		affix     string
		condition string
	}{{
		stripping: "0",
		affix:     "pre1",
		condition: ".",
		exp: &affixRule{
			affix: "pre1",
		},
	}, {
		stripping: "0",
		affix:     "pre2",
		condition: "o",
		exp: &affixRule{
			affix:     "pre2",
			condition: regexp.MustCompile("^o"),
		},
	}, {
		stripping: "0",
		affix:     "pre3",
		condition: "[aeou]",
		exp: &affixRule{
			affix:     "pre3",
			condition: regexp.MustCompile("^[aeou]"),
		},
	}}

	for _, c := range cases {
		got, err := newAffixRule(&spell.opts, isPrefix,
			c.stripping, c.affix, c.condition, nil)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "newAffixRule", c.exp, got)
	}
}
