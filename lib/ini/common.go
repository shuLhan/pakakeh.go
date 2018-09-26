// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"strings"
)

//
// IsValueBoolTrue will return true if variable contains boolean value for
// true. The following conditions is boolean true for value: "" (empty
// string), "true", "yes", "ya", "t", "1" (all of string is case insensitive).
//
func IsValueBoolTrue(v string) bool {
	if len(v) == 0 {
		return true
	}
	v = strings.ToLower(v)
	if v == "true" || v == "t" || v == "ya" || v == "yes" || v == "1" {
		return true
	}
	return false
}
