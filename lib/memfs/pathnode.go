// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
)

//
// PathNode contains a mapping between path and Node.
//
type PathNode struct {
	mu sync.Mutex
	v  map[string]*Node
}

//
// NewPathNode create and initialize new PathNode.
//
func NewPathNode() *PathNode {
	return &PathNode{
		v: make(map[string]*Node),
	}
}

//
// Delete the the node by its path.
//
func (pn *PathNode) Delete(path string) {
	pn.mu.Lock()
	delete(pn.v, path)
	pn.mu.Unlock()
}

//
// Get the node by path, or nil if path is not exist.
//
func (pn *PathNode) Get(path string) (node *Node) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	if pn.v == nil {
		return nil
	}
	return pn.v[path]
}

func (pn *PathNode) MarshalJSON() ([]byte, error) {
	var (
		buf   bytes.Buffer
		paths = pn.Paths()
		x     int
		path  string
		node  *Node
	)

	pn.mu.Lock()
	_ = buf.WriteByte('{')
	for x, path = range paths {
		if x > 0 {
			_ = buf.WriteByte(',')
		}
		fmt.Fprintf(&buf, "%q:", path)
		node = pn.v[path]
		if node != nil {
			node.packAsJson(&buf, 0)
		}
	}
	_ = buf.WriteByte('}')
	pn.mu.Unlock()

	return buf.Bytes(), nil
}

//
// Nodes return all the nodes.
//
func (pn *PathNode) Nodes() (nodes []*Node) {
	var (
		node *Node
	)

	pn.mu.Lock()
	for _, node = range pn.v {
		nodes = append(nodes, node)
	}
	pn.mu.Unlock()
	return nodes
}

//
// Paths return all the nodes paths sorted in ascending order.
//
func (pn *PathNode) Paths() (paths []string) {
	var path string
	pn.mu.Lock()
	for path = range pn.v {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	pn.mu.Unlock()
	return paths
}

//
// Set mapping of path to Node.
//
func (pn *PathNode) Set(path string, node *Node) {
	if len(path) == 0 || node == nil {
		return
	}
	pn.mu.Lock()
	pn.v[path] = node
	pn.mu.Unlock()
}
