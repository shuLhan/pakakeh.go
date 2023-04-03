// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bytes is the internal, shared functions for manipulating bytes,
// created to prevent import cycle.
package bytes

// TokenFind return the first index of matched token in text, start at custom
// index.
// If "startat" parameter is less than 0, then it will be set to 0.
// If token is empty or no token found it will return -1.
func TokenFind(text, token []byte, startat int) (at int) {
	var (
		textlen  = len(text)
		tokenlen = len(token)
	)
	if tokenlen == 0 {
		return -1
	}
	if startat < 0 {
		startat = 0
	}

	var (
		y = 0
		x = startat
	)
	at = -1
	for ; x < textlen; x++ {
		if text[x] == token[y] {
			if y == 0 {
				at = x
			}
			y++
			if y == tokenlen {
				// We found it!
				return at
			}
		} else if at != -1 {
			// Reset back.
			y = 0
			at = -1
		}
	}
	// x run out before y.
	if y < tokenlen {
		at = -1
	}

	return at
}
