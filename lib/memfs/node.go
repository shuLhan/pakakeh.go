// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"mime"
	"net/http"
	"os"
	"path"
)

//
// Node represent a single file.
//
type Node struct {
	SysPath     string      // The original file path in system.
	Path        string      // Absolute file path in memory.
	Name        string      // File name.
	ContentType string      // File type per MIME, e.g. "application/json".
	Mode        os.FileMode // File mode.
	Size        int64       // Size of file.
	V           []byte      // Content of file.
	Parent      *Node       // Pointer to parent directory.
	Childs      []*Node     // List of files in directory.
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

func (leaf *Node) updateContentType() error {
	leaf.ContentType = mime.TypeByExtension(path.Ext(leaf.Name))
	if len(leaf.ContentType) > 0 {
		return nil
	}

	if len(leaf.V) > 0 {
		leaf.ContentType = http.DetectContentType(leaf.V)
		return nil
	}

	data := make([]byte, 512)

	f, err := os.Open(leaf.SysPath)
	if err != nil {
		return err
	}

	_, err = f.Read(data)
	if err != nil {
		errc := f.Close()
		if errc != nil {
			panic(errc)
		}
		return err
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	leaf.ContentType = http.DetectContentType(data)

	return nil
}
