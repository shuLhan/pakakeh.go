// Copyright (c) 2022, Wu Tingfeng <wutingfeng@outlook.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ascii

import (
	"unicode/utf8"
)

// Set is a 36-byte value, where each bit in the first 32-bytes
// represents the presence of a given ASCII character in the set.
// The remaining 4-bytes is a counter for the number of ASCII characters in
// the set.
// The 128-bits of the first 16 bytes, starting with the least-significant bit
// of the lowest word to the most-significant bit of the highest word,
// map to the full range of all 128 ASCII characters.
// The 128-bits of the next 16 bytes will be zeroed,
// ensuring that any non-ASCII character will be reported as not in the set.
// Rejecting non-ASCII characters in this way avoids bounds checks in
// Set.Contains.
type Set [9]uint32

// MakeSet creates a set of ASCII characters and reports whether all
// characters in chars are ASCII.
func MakeSet(chars string) (as Set, ok bool) {
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if c >= utf8.RuneSelf {
			return as, false
		}
		as.Add(c)
	}
	return as, true
}

// Add inserts character c into the set.
func (as *Set) Add(c byte) {
	if c < utf8.RuneSelf { // ensure that c is an ASCII byte
		before := as[c/32]
		as[c/32] |= 1 << (c % 32)
		if before != as[c/32] {
			as[8]++
		}
	}
}

// Contains reports whether c is inside the set.
func (as *Set) Contains(c byte) bool {
	return (as[c/32] & (1 << (c % 32))) != 0
}

// Remove removes c from the set
//
// if c is not in the set, the set contents will remain unchanged.
func (as *Set) Remove(c byte) {
	if c < utf8.RuneSelf { // ensure that c is an ASCII byte
		before := as[c/32]
		as[c/32] &^= 1 << (c % 32)
		if before != as[c/32] {
			as[8]--
		}
	}
}

// Size returns the number of characters in the set.
func (as *Set) Size() int {
	return int(as[8])
}

// Union returns a new set containing all characters that belong to either as and as2.
func (as *Set) Union(as2 Set) (as3 Set) {
	as3[0] = as[0] | as2[0]
	as3[1] = as[1] | as2[1]
	as3[2] = as[2] | as2[2]
	as3[3] = as[3] | as2[3]
	return
}

// Intersection returns a new set containing all characters that belong to both as and as2.
func (as *Set) Intersection(as2 Set) (as3 Set) {
	as3[0] = as[0] & as2[0]
	as3[1] = as[1] & as2[1]
	as3[2] = as[2] & as2[2]
	as3[3] = as[3] & as2[3]
	return
}

// Subtract returns a new set containing all characters that belong to as but not as2.
func (as *Set) Subtract(as2 Set) (as3 Set) {
	as3[0] = as[0] &^ as2[0]
	as3[1] = as[1] &^ as2[1]
	as3[2] = as[2] &^ as2[2]
	as3[3] = as[3] &^ as2[3]
	return
}

// Equal reports whether as contains the same characters as as2.
func (as *Set) Equal(as2 Set) bool {
	return as[0] == as2[0] && as[1] == as2[1] && as[2] == as2[2] && as[3] == as2[3]
}

// Visit calls the do function for each character of as in ascending numerical order.
// If do returns true, Visit returns immediately, skipping any remaining
// characters, and returns true. It is safe for do to Add or Remove
// characters. The behavior of Visit is undefined if do changes
// the set in any other way.
func (as *Set) Visit(do func(n byte) (skip bool)) (aborted bool) {
	var currentChar byte
	for i := uint(0); i < 4; i++ {
		for j := uint(0); j < 32; j++ {
			if (as[i] & (1 << j)) != 0 {
				if do(currentChar) {
					return true
				}
			}
			currentChar++
		}
	}
	return false
}
