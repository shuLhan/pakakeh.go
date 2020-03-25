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

//
// affixRule represent each prefix or suffix rule.
//
// Syntax for affix rule,
//
//	AFFIX_RULE := STRIPPING AFFIX ( FLAGS ) CONDITION MORPHEMES
//
//	STRIPPING  := "0" / 1*UTF8_VCHAR
//
//	AFFIX      := 1*UTF8_VCHAR
//
//	FLAGS      := "/" ( 1*AF_ALIAS / AFFIX_NAME )
//	AF_ALIAS   := 1*DIGIT
//
//	CONDITION  := "." / COND_RE
//	COND_RE    := "[" ( "^" ) 1*UTF8_VCHAR "]"
//
//	MORPHEMES  := 1*AM_ALIAS / *MORPHEME
//	AM_ALIAS   := 1*DIGIT
//
// For example, the affix rule for prefix line "PFX A 0 x . 1" is "0 x . 1".
// The "0" means no stripping, "x" is the prefix to be added to the word, "."
// means zero condition, and "1" is an alias to the first morpheme defined in
// "AM".
//
//
type affixRule struct {
	// stripping characters from beginning (at  prefix  rules)  or  end
	// (at  suffix rules) of the word.
	// Zero stripping is indicated by zero, in our case its empty.
	stripping string

	// affix contains the root affix.
	// Zero affix is indicated by empty string.
	affix string

	// An affix rule can contains another affix rules, chaining one or
	// more affix together.
	rawFlags string
	affixes  []*affix

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

func newAffixRule(opts *affixOptions, isPrefix bool,
	stripping, affix, condition string, morphemes []string,
) (
	rule *affixRule, err error,
) {
	affixes := opts.suffixes

	rule = &affixRule{
		morphemes: morphemes,
	}

	if isPrefix {
		affixes = opts.prefixes
	}

	if stripping != "0" {
		rule.stripping = stripping
	}
	if affix != "0" {
		affixflag := strings.Split(affix, "/")

		rule.affix = affixflag[0]

		// Expand the flags into affixes.
		if len(affixflag) > 1 {
			rule.rawFlags = affixflag[1]

			err = rule.unpackFlags(opts, affixes)
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
func (rule *affixRule) unpackFlags(
	opts *affixOptions, affixes map[string]*affix,
) (err error) {
	if len(opts.afAliases) > 1 {
		afIdx, err := strconv.Atoi(rule.rawFlags)
		if err == nil {
			rule.rawFlags = opts.afAliases[afIdx]
		}
	}

	flags, err := unpackFlags(opts.flag, rule.rawFlags)
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
