// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"strings"
)

// affix represent the prefix or suffix and its rules.
//
// Syntax,
//
//	AFFIX          := "PFX" / "SFX" WSP AFFIX_NAME WSP CROSS_PRODUCT WSP NRULES
//
//	AFFIX_NAME     := 1*UTF8_VCHAR
//
//	CROSS_PRODUCT  := "N" / "Y"
//
//	NRULES         := 1*DIGIT
type affix struct {
	// Name of the affix class.
	name string

	rules []*affixRule

	isPrefix bool

	// isCrossProduct indicate whether an affix can be combined with
	// another affix.
	isCrossProduct bool
}

func newAffix(name string, isPrefix, isCrossProduct bool, nrules int) (afx *affix) {
	afx = &affix{
		isPrefix:       isPrefix,
		isCrossProduct: isCrossProduct,
		name:           name,
		rules:          make([]*affixRule, 0, nrules),
	}
	return afx
}

// addRule to affix.
func (afx *affix) addRule(opts *affixOptions,
	stripping, affix, condition string, morphemes []string,
) (
	err error,
) {
	rule, err := newAffixRule(opts, afx.isPrefix, stripping,
		affix, condition, morphemes)
	if err != nil {
		return err
	}

	afx.rules = append(afx.rules, rule)

	return nil
}

// apply the affixes to the root stem, return the list of stem.
func (afx *affix) apply(root *Stem) (ss []*Stem) {
	var word string

	for _, rule := range afx.rules {
		if rule.condition != nil && !rule.condition.MatchString(root.Word) {
			continue
		}

		if afx.isPrefix {
			word = strings.TrimPrefix(root.Word, rule.stripping)
			word = rule.affix + word
		} else {
			word = strings.TrimSuffix(root.Word, rule.stripping)
			word += rule.affix
		}

		stem := newStem(root, word, rule.morphemes)

		if len(rule.affixes) == 0 {
			ss = append(ss, stem)
		} else {
			for _, subafx := range rule.affixes {
				sublist := subafx.apply(stem)

				ss = append(ss, sublist...)
			}
			ss = append(ss, stem)
		}
	}

	return ss
}
