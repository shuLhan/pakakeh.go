// SPDX-FileCopyrightText: 2015 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"fmt"
	"reflect"
)

// BTNode is a data type for node in binary tree.
type BTNode struct {
	// Left branch of node.
	Left *BTNode
	// Right branch of node.
	Right *BTNode
	// Parent of node.
	Parent *BTNode
	// Value of node.
	Value any
}

// NewBTNode create new node for binary tree.
func NewBTNode(v any, l *BTNode, r *BTNode) (p *BTNode) {
	p = &BTNode{
		Left:   l,
		Right:  r,
		Parent: nil,
		Value:  v,
	}
	if l != nil {
		l.Parent = p
	}
	if r != nil {
		r.Parent = p
	}

	return p
}

// SetLeft will set left branch of node to 'c'.
func (n *BTNode) SetLeft(c *BTNode) {
	n.Left = c
	c.Parent = n
}

// SetRight will set right branch of node to 'c'.
func (n *BTNode) SetRight(c *BTNode) {
	n.Right = c
	c.Parent = n
}

// String will convert the node to string.
func (n *BTNode) String() (s string) {
	var p = n.Parent

	// add tab until it reached nil
	for p != nil {
		s += "\t"
		p = p.Parent
	}

	s += fmt.Sprintln(reflect.ValueOf(n.Value))

	return s
}
