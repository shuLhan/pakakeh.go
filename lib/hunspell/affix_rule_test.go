// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"regexp"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewAffixRule_prefix(t *testing.T) {
	var isPrefix = true
	var spell = &Spell{
		flag: FlagASCII,
	}

	cases := []struct {
		stripping string
		affix     string
		condition string
		exp       *affixRule
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
		got, err := newAffixRule(spell, isPrefix,
			c.stripping, c.affix, c.condition, nil)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "newAffixRule", c.exp, got, true)
	}
}
