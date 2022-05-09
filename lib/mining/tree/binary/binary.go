// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package binary contain implementation of binary tree.
package binary

import (
	"fmt"
)

// Tree is a abstract data type for tree with only two branch: left and right.
type Tree struct {
	// Root is pointer to root of tree.
	Root *BTNode
}

// NewTree will create new binary tree with empty root.
func NewTree() *Tree {
	return &Tree{
		Root: nil,
	}
}

// String will print all the branch and leaf in tree.
func (btree *Tree) String() (s string) {
	var parent, node *BTNode

	parent = btree.Root

	s = fmt.Sprint(parent)

	node = parent.Right

	for parent != nil {
		// Print right node down to the leaf.
		for node.Right != nil {
			s += fmt.Sprint(node)

			parent = node
			node = node.Right
		}
		s += fmt.Sprint(node)

		// crawling to the stop one at a time ...
		for parent != nil && node.Parent == parent {
			if parent.Right == node {
				node = parent.Left
				break
			} else if parent.Left == node {
				node = parent
				parent = parent.Parent
			}
		}
	}

	return s
}
