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
// An empty string is equal to "_".
// Any other unknown characters will be replaced with '_'.
// If the input does not start with letter, it will be prefixed with
// '_', unless it start with '_'.
//
// [Mozilla specification]: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/id.
func NormalizeForID(in string) (out string) {
	var (
		bin  = []byte(in)
		bout = make([]byte, 0, len(bin)+1)
		b    byte
	)

	for _, b = range bin {
		if ascii.IsAlnum(b) || b == '-' || b == '_' {
			bout = append(bout, b)
		} else {
			bout = append(bout, '_')
		}
	}
	if len(bout) == 0 {
		bout = append(bout, '_')
	} else if !ascii.IsAlpha(bout[0]) && bout[0] != '_' {
		bout = append(bout, '_')
		copy(bout[1:], bout[:])
		bout[0] = '_'
	}

	return string(bout)
}
