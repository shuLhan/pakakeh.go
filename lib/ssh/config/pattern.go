// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import "path/filepath"

type pattern struct {
	value    string
	isNegate bool
}

func newPattern(s string) (pat *pattern) {
	pat = &pattern{}
	if s[0] == '!' {
		pat.isNegate = true
		pat.value = s[1:]
	} else {
		pat.value = s
	}
	return pat
}

//
// isMatch will return true if input string match with regex and isNegate is
// false; otherwise it will return false.
//
func (pat *pattern) isMatch(s string) bool {
	ok, err := filepath.Match(pat.value, s)
	if err != nil {
		return false
	}
	if ok {
		return !pat.isNegate
	}
	return false
}
