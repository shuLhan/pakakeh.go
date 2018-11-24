// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"os"
)

type Node struct {
	SysPath string
	Path    string
	Name    string
	Mode    os.FileMode
	Size    int64
	V       []byte
	Parent  *Node
	Childs  []*Node
}

func newNode(path string) (*Node, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &Node{
		SysPath: path,
		Path:    "/",
		Name:    "/",
		Mode:    fi.Mode(),
		Size:    fi.Size(),
		V:       nil,
		Parent:  nil,
		Childs:  make([]*Node, 0),
	}

	return node, nil
}

func (leaf *Node) removeChild(child *Node) {
	for x := 0; x < len(leaf.Childs); x++ {
		if leaf.Childs[x] != child {
			continue
		}

		copy(leaf.Childs[x:], leaf.Childs[x+1:])
		n := len(leaf.Childs)
		leaf.Childs[n-1] = nil
		leaf.Childs = leaf.Childs[:n-1]

		child.Parent = nil
		child.Childs = nil
	}
}
