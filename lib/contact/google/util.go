// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

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
