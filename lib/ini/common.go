// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"strings"
	"unicode"
)

//
// IsValidVarName check if "v" is valid variable name, where the
// first character must be a letter and the rest should contains only letter,
// digit, period, hyphen, or underscore.
// If "v" is valid it will return true.
//
func IsValidVarName(v string) bool {
	for x, r := range v {
		if x == 0 && !unicode.IsLetter(r) {
			return false
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) ||
			r == tokHyphen || r == tokDot || r == tokUnderscore {
			continue
		}
		return false
	}
	return true
}

//
// IsValueBoolTrue will return true if variable contains boolean value for
// true. The following conditions is boolean true for value: "" (empty
// string), "true", "yes", "ya", "t", "1" (all of string is case insensitive).
//
func IsValueBoolTrue(v string) bool {
	if len(v) == 0 {
		return false
	}
	v = strings.ToLower(v)
	if v == "true" || v == "t" || v == "ya" || v == "yes" || v == "1" {
		return true
	}
	return false
}
