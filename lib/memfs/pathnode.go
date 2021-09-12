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
	f  map[string]func() *Node
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
func (pn *PathNode) Get(path string) *Node {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	node, ok := pn.v[path]
	if ok {
		return node
	}
	if pn.f != nil {
		f, ok := pn.f[path]
		if ok {
			node = f()
			return node
		}
	}
	return nil
}

func (pn *PathNode) MarshalJSON() ([]byte, error) {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	// Merge the path with function to node into v.
	for k, fn := range pn.f {
		pn.v[k] = fn()
	}

	buf := bytes.Buffer{}

	// Sort the paths.
	keys := make([]string, 0, len(pn.v))
	for path := range pn.v {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	_ = buf.WriteByte('{')
	for x, key := range keys {
		if x > 0 {
			_ = buf.WriteByte(',')
		}
		fmt.Fprintf(&buf, "%q:", key)
		node := pn.v[key]
		node.packAsJson(&buf, 0)
	}
	_ = buf.WriteByte('}')

	return buf.Bytes(), nil
}

//
// Nodes return all the nodes.
//
func (pn *PathNode) Nodes() (nodes []*Node) {
	pn.mu.Lock()
	for _, node := range pn.v {
		nodes = append(nodes, node)
	}
	pn.mu.Unlock()
	return nodes
}

//
// Paths return all the nodes keys as list of path.
//
func (pn *PathNode) Paths() (paths []string) {
	pn.mu.Lock()
	for key := range pn.v {
		paths = append(paths, key)
	}
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
