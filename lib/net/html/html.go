// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package html extends the golang.org/x/net/html by providing simplified
// methods to Node.
//
// The x/net/html package currently only provide bare raw functionalities
// to iterate tree, there is no check for empty node, and no function to
// get attribute by name without looping it manually.
//
// This package extends the parent package by adding methods to get node's
// attribute by name, get the first non-empty child, get the next
// non-empty sibling, and method to iterate the tree.
package html

import (
	"github.com/shuLhan/share/lib/ascii"
)

// NormalizeForID given an input string normalize it to HTML ID.
// The normalization follow [Mozilla specification] rules,
//
//   - it must not contain whitespace (spaces, tabs etc.),
//   - only ASCII letters, digits, '_', and '-' should be used, and
//   - it should start with a letter.
//
// This function,
//
//   - Return "_" if input is empty string,
//   - replace unknown characters with '_',
//   - prefix output with '_' unless it start with '-', '_', or letters, and
//   - convert letters to lower cases.
//
// [Mozilla specification]: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/id.
func NormalizeForID(in string) (out string) {
	var (
		bin = []byte(in)
		x   int
		b   byte
	)

	for x, b = range bin {
		if ascii.IsAlpha(b) {
			if b >= 'A' && b <= 'Z' {
				bin[x] = b + 32
			}
		} else if !(ascii.IsDigit(b) || b == '-' || b == '_') {
			bin[x] = '_'
		}
	}
	if len(bin) == 0 {
		bin = append(bin, '_')
	} else if !ascii.IsAlpha(bin[0]) && bin[0] != '_' {
		bin = append(bin, '_')
		copy(bin[1:], bin[:])
		bin[0] = '_'
	}

	return string(bin)
}
