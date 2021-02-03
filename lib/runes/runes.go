// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package runes provide a library for working with a single rune or slice of
// rune.
//
package runes

import (
	"unicode"
)

//
// Contain find a rune `c` inside `s`.
// If `c` found in `s` it will return boolean true and index of `c` in `s`;
// otherwise it will return false and -1.
//
func Contain(s []rune, c rune) (bool, int) {
	for x, v := range s {
		if v == c {
			return true, x
		}
	}
	return false, -1
}

//
// Diff return the difference between two slice of rune.
//
func Diff(l []rune, r []rune) (diff []rune) {
	var found bool
	dupDiff := []rune{}

	// Find l not in r
	for _, v := range l {
		found, _ = Contain(r, v)
		if !found {
			dupDiff = append(dupDiff, v)
		}
	}

	// Find r not in diff
	for _, v := range r {
		found, _ = Contain(l, v)
		if !found {
			dupDiff = append(dupDiff, v)
		}
	}

	// Remove duplicate in dupDiff
	duplen := len(dupDiff)
	for x, v := range dupDiff {
		found = false
		for y := x + 1; y < duplen; y++ {
			if v == dupDiff[y] {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, v)
		}
	}

	return diff
}

//
// EncloseRemove given a line, remove all characters inside it, starting
// from `leftcap` until the `rightcap` and return cutted line and changed to
// true.
//
// If no `leftcap` or `rightcap` is found, the line will unchanged, and
// returned status will be false.
//
func EncloseRemove(line, leftcap, rightcap []rune) ([]rune, bool) {
	lidx := TokenFind(line, leftcap, 0)
	ridx := TokenFind(line, rightcap, lidx+1)

	if lidx < 0 || ridx < 0 || lidx >= ridx {
		return line, false
	}

	var newline []rune
	newline = append(newline, line[:lidx]...)
	newline = append(newline, line[ridx+len(rightcap):]...)
	newline, _ = EncloseRemove(newline, leftcap, rightcap)

	return newline, true
}

//
// FindSpace find any unicode spaces in line start from index `startAt` and
// return their index.
// If no spaces found it will return -1.
//
func FindSpace(line []rune, startAt int) (idx int) {
	lineLen := len(line)
	if startAt < 0 {
		startAt = 0
	}

	for idx = startAt; idx < lineLen; idx++ {
		if unicode.IsSpace(line[idx]) {
			return
		}
	}
	return -1
}

//
// Inverse the input slice of rune with inplace reversion (without allocating
// another slice).
//
func Inverse(in []rune) []rune {
	var (
		left, right rune
		y           = len(in) - 1
	)
	for x := 0; x < y; x++ {
		left = in[x]
		right = in[y]
		in[x] = right
		in[y] = left
		y--
	}
	return in
}

//
// TokenFind will search token in text starting from index `startAt` and
// return the position where the match start.
//
// If no token is found it will return -1.
//
func TokenFind(line, token []rune, startAt int) (at int) {
	y := 0
	tokenlen := len(token)
	linelen := len(line)

	at = -1
	for x := startAt; x < linelen; x++ {
		if line[x] == token[y] {
			if y == 0 {
				at = x
			}
			y++
			if y == tokenlen {
				// we found it!
				return
			}
		} else if at != -1 {
			// reset back
			y = 0
			at = -1
		}
	}
	// x run out before y
	if y < tokenlen {
		at = -1
	}
	return
}
