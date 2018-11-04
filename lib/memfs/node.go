// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"os"
)

type node struct {
	sysPath string
	path    string
	name    string
	mode    os.FileMode
	size    int64
	v       []byte
	parent  *node
	childs  []*node
}

func newNode(path string) (*node, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &node{
		sysPath: path,
		path:    "/",
		name:    "/",
		mode:    fi.Mode(),
		size:    fi.Size(),
		v:       nil,
		parent:  nil,
		childs:  make([]*node, 0),
	}

	return node, nil
}

func (leaf *node) removeChild(child *node) {
	for x := 0; x < len(leaf.childs); x++ {
		if leaf.childs[x] != child {
			continue
		}

		copy(leaf.childs[x:], leaf.childs[x+1:])
		n := len(leaf.childs)
		leaf.childs[n-1] = nil
		leaf.childs = leaf.childs[:n-1]

		child.parent = nil
		child.childs = nil
	}
}
