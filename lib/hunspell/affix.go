// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import "strings"

//
// affix represent the prefix or suffix and its rules.
//
type affix struct {
	isPrefix bool

	// isCrossProduct indicate whether an affix can be combined with
	// another affix.
	isCrossProduct bool

	// Name of the affix class.
	name string

	rules []*affixRule
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

//
// addRule to affix.
//
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

//
// apply the root string with list of affixes according to its rule.
//
func (afx *affix) apply(root string) (ss []string) {
	var newroot string

	for _, rule := range afx.rules {
		if rule.condition != nil && !rule.condition.MatchString(root) {
			continue
		}

		if afx.isPrefix {
			newroot = strings.TrimPrefix(root, rule.stripping)
			newroot = rule.affix + newroot
		} else {
			newroot = strings.TrimSuffix(root, rule.stripping)
			newroot += rule.affix
		}
		if len(rule.affixes) == 0 {
			ss = append(ss, newroot)
		} else {
			for _, subafx := range rule.affixes {
				sublist := subafx.apply(newroot)

				ss = append(ss, sublist...)
			}
			ss = append(ss, newroot)
		}
	}

	return ss
}
