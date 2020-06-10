// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

//
// NodeIterator simplify iterating each node from top to bottom.
//
type NodeIterator struct {
	current  *Node
	previous *Node
	next     *html.Node
	hasNext  bool
}

//
// Parse returns the NodeIterator to iterate through HTML tree.
//
func Parse(r io.Reader) (iter *NodeIterator, err error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	iter = &NodeIterator{
		current:  NewNode(node),
		previous: NewNode(nil),
	}
	return iter, nil
}

//
// Next return the first child or the next sibling of current node.
// If no more node in the tree, it will return nil.
//
func (iter *NodeIterator) Next() *Node {
	if iter.hasNext {
		iter.current.Node = iter.next
		iter.next = nil
		iter.hasNext = false
		return iter.current
	}
	if iter.current.Node == nil {
		return nil
	}

	for {
		switch {
		case iter.current.FirstChild != nil &&
			iter.current.FirstChild != iter.previous.Node &&
			iter.current.LastChild != iter.previous.Node:
			iter.current.Node = iter.current.FirstChild
		case iter.current.NextSibling != nil:
			iter.current.Node = iter.current.NextSibling
		default:
			iter.previous.Node = iter.current.Node
			iter.current.Node = iter.current.Parent
		}
		if iter.current.Node == nil {
			return nil
		}

		// Skip empty text node.
		if iter.current.Type != html.TextNode {
			break
		}
		text := strings.TrimSpace(iter.current.Data)
		if len(text) != 0 {
			break
		}
	}
	return iter.current
}

//
// SetNext set the node for iteration to Node "el" only if its not nil.
//
func (iter *NodeIterator) SetNext(el *Node) {
	if el == nil {
		return
	}
	iter.hasNext = true
	iter.next = el.Node
}

//
// SetNextNode set the next iteration node to html.Node "el" only if its not
// nil.
//
func (iter *NodeIterator) SetNextNode(el *html.Node) {
	if el == nil {
		return
	}
	iter.hasNext = true
	iter.next = el
}
