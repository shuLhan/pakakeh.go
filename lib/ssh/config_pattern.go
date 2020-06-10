// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import "path/filepath"

type configPattern struct {
	pattern  string
	isNegate bool
}

func newConfigPattern(s string) (pat *configPattern) {
	pat = new(configPattern)
	if s[0] == '!' {
		pat.isNegate = true
		s = s[1:]
	}
	pat.pattern = s
	return pat
}

//
// isMatch will return true if input string match with regex and isNegate is
// false; otherwise it will return false.
//
func (pat *configPattern) isMatch(s string) bool {
	ok, err := filepath.Match(pat.pattern, s)
	if err != nil {
		return false
	}
	if ok {
		return !pat.isNegate
	}
	return false
}
