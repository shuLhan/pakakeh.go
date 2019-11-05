// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type affixRule struct {
	// stripping characters from beginning (at  prefix  rules)  or  end
	// (at  suffix rules) of the word.
	// Zero stripping is indicated by zero, in our case its empty.
	stripping string

	// affix contains the root affix.
	// Zero affix is indicated by empty string.
	affix   string
	flags   string
	affixes []*affix

	//
	// condition is a simplified, regular expression-like pattern, which
	// must be met before the affix can be applied.
	// A dot signs an arbitrary character.
	// Characters in braces sign an arbitrary character from the character
	// subset.
	// Dash hasn't got special meaning, but circumflex (^) next the first
	// brace sets the complementer character set.
	//
	// Zero condition is indicated by dot ("."), or in our case its a nil
	// Regexp.
	//
	condition *regexp.Regexp

	// morphemes contains optional morphological fields separated by
	// spaces or tabulators.
	morphemes []string
}

func newAffixRule(spell *Spell, isPrefix bool,
	stripping, affix, condition string, morphemes []string,
) (
	rule *affixRule, err error,
) {
	affixes := spell.suffixes

	rule = &affixRule{
		morphemes: morphemes,
	}

	if isPrefix {
		affixes = spell.prefixes
	}

	if stripping != "0" {
		rule.stripping = stripping
	}
	if affix != "0" {
		affixflag := strings.Split(affix, "/")

		rule.affix = affixflag[0]

		// Expand the flags into affixes.
		if len(affixflag) > 1 {
			rule.flags = affixflag[1]

			err = rule.unpackFlags(spell, affixes)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(condition) > 0 && condition != "." {
		if isPrefix {
			condition = "^" + condition
		} else {
			condition += "$"
		}
		rule.condition, err = regexp.Compile(condition)
		if err != nil {
			return nil, fmt.Errorf("invalid condition %q: %w", condition, err)
		}
	}

	return rule, nil
}

//
// unpackFlags apply each of flag rule to the "root" string.
//
func (rule *affixRule) unpackFlags(spell *Spell, affixes map[string]*affix) (err error) {
	if len(spell.afAliases) > 1 {
		afIdx, err := strconv.Atoi(rule.flags)
		if err == nil {
			rule.flags = spell.afAliases[afIdx]
		}
	}

	flags, err := unpackFlags(spell.flag, rule.flags)
	if err != nil {
		return err
	}

	for _, s := range flags {
		afx, ok := affixes[s]
		if !ok {
			return fmt.Errorf("unknown affix flag %q", s)
		}

		rule.affixes = append(rule.affixes, afx)
	}

	return nil
}
