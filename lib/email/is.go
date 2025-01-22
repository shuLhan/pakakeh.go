// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package email

import "git.sr.ht/~shulhan/pakakeh.go/lib/ascii"

var specialChars, _ = ascii.MakeSet(`()<>[]:;@\,"`)

// IsValidLocal will return true if local part contains valid characters.
// Local part must,
//   - start or end without dot character,
//   - contains only printable US-ASCII characters, excluding special
//     characters
//   - no multiple sequence of dots.
//
// List of special characters,
//
//	"(" / ")" / "<" / ">" / "[" / "]" / ":" / ";" / "@" / "\" / "," / "." / DQUOTE
func IsValidLocal(local []byte) bool {
	if len(local) == 0 {
		return false
	}
	if local[0] == '.' || local[len(local)-1] == '.' {
		return false
	}
	dot := false
	for x := range len(local) {
		if local[x] < 33 || local[x] > 126 {
			return false
		}
		if local[x] == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		dot = false
		// if _, ok := specialCharsOld[local[x]]; ok {
		if specialChars.Contains(local[x]) {
			return false
		}
	}
	return true
}
