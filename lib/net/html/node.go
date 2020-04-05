// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"strings"

	"golang.org/x/net/html"
)

//
// Node extends the html.Node.
//
type Node struct {
	*html.Node
}

//
// NewNode create new node by embedding html.Node "el".
//
func NewNode(el *html.Node) *Node {
	return &Node{Node: el}
}

//
// GetAttrValue get the value of node's attribute with specific key or empty
// if key not found.
//
func (node *Node) GetAttrValue(key string) string {
	for _, attr := range node.Attr {
		if key == attr.Key {
			return attr.Val
		}
	}
	return ""
}

//
// GetFirstChild get the first non-empty child of node or nil if no child
// left.
//
func (node *Node) GetFirstChild() *Node {
	el := node.FirstChild
	for el != nil {
		if el.Type == html.TextNode {
			if len(strings.TrimSpace(el.Data)) == 0 {
				el = el.NextSibling
				continue
			}
		}
		break
	}
	if el == nil {
		return nil
	}
	return NewNode(el)
}

//
// GetNextSibling get the next non-empty sibling of node or nil if no more
// sibling left.
//
func (node *Node) GetNextSibling() *Node {
	el := node.NextSibling
	for el != nil {
		if el.Type == html.TextNode {
			if len(strings.TrimSpace(el.Data)) == 0 {
				el = el.NextSibling
				continue
			}
		}
		break
	}
	if el == nil {
		return nil
	}
	return NewNode(el)
}

//
// IsElement will return true if node type is html.ElementNode.
//
func (node *Node) IsElement() bool {
	return node.Type == html.ElementNode
}
