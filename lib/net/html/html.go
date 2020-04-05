// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package html extends the golang.org/x/net/html by providing simplified
// methods to Node.
//
// The x/net/html package currently only provide bare raw functionalities
// to iterate tree, there is no check for empty node, and no function to
// get attribute by name without looping it manually.
//
// This package extends the package by adding methods to get node's attribute
// by name, get the first non-empty child, and get the next non-empty sibling
//
package html

import (
	"io"

	"golang.org/x/net/html"
)

//
// Parse returns the parse tree for the HTML from the given Reader.
//
func Parse(r io.Reader) (*Node, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return NewNode(node), nil
}
