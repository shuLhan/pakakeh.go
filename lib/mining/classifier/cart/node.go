// SPDX-FileCopyrightText: 2015 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package cart

import (
	"fmt"
	"reflect"
)

// NodeValue of tree in CART.
type NodeValue struct {
	// SplitV define the split value.
	SplitV any

	// Class of leaf node.
	Class string

	// SplitAttrName define the name of attribute which cause the split.
	SplitAttrName string

	// Size define number of sample that this node hold before splitting.
	Size int

	// SplitAttrIdx define the attribute which cause the split.
	SplitAttrIdx int

	// IsLeaf define whether node is a leaf or not.
	IsLeaf bool

	// IsContinu define whether the node split is continuous or discrete.
	IsContinu bool
}

// String will return the value of node for printable.
func (nodev *NodeValue) String() (s string) {
	if nodev.IsLeaf {
		s = `Class: ` + nodev.Class
	} else {
		s = fmt.Sprintf("(SplitValue: %v)",
			reflect.ValueOf(nodev.SplitV))
	}

	return s
}
