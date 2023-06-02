// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import "github.com/elliotwutingfeng/asciiset"

var specialChars, _ = asciiset.MakeASCIISet("()<>[]:;@\\,\"")

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
	for x := 0; x < len(local); x++ {
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
		if ok := specialChars.Contains(local[x]); ok {
			return false
		}
	}
	return true
}
