// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

import (
	"strings"
)

// ParseRel will parse Google "rel" value and return the type.
func ParseRel(in string) string {
	kv := strings.Split(in, "#")
	if len(kv) != 2 {
		return ""
	}

	return kv[1]
}
