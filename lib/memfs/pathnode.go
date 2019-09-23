// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

//
// PathNode contains a mapping between path and Node.
//
type PathNode struct {
	v map[string]*Node
	f map[string]func() *Node
}

//
// NewPathNode create and initialize new PathNode.
//
func NewPathNode() *PathNode {
	return &PathNode{
		v: make(map[string]*Node),
		f: make(map[string]func() *Node),
	}
}

//
// Get the node by path, or nil if path is not exist.
//
func (pn *PathNode) Get(path string) *Node {
	node, ok := pn.v[path]
	if ok {
		return node
	}
	if pn.f != nil {
		f, ok := pn.f[path]
		if ok {
			return f()
		}
	}
	return nil
}

//
// Set mapping of path to Node.
//
func (pn *PathNode) Set(path string, node *Node) {
	if len(path) == 0 || node == nil {
		return
	}
	pn.v[path] = node
}
