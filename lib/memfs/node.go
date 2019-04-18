// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

//
// newNode create a new node based on file information "fi".
// If withContent is true, the file content and its type will be saved in
// node as V and ContentType.
//
func newNode(parent *Node, fi os.FileInfo, withContent bool) (node *Node, err error) {
	if parent == nil || fi == nil {
		return nil, nil
	}

	sysPath := filepath.Join(parent.SysPath, fi.Name())

	node = &Node{
		SysPath: sysPath,
		Path:    path.Join(parent.Path, fi.Name()),
		Name:    fi.Name(),
		Mode:    fi.Mode(),
		Size:    fi.Size(),
		V:       nil,
		Parent:  parent,
		Childs:  make([]*Node, 0),
	}

	if node.Mode.IsDir() || !withContent {
		return node, nil
	}

	err = node.updateContent()
	if err != nil {
		return nil, err
	}

	err = node.updateContentType()
	if err != nil {
		return nil, err
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

//
// update the node content and information based on new file information.
//
// If the newInfo is nil, it will read the file information based on node's
// SysPath.
//
// There are two possible changes that will happened: its either change on
// mode or change on content (size and modtime).
// Change on mode will not affect the content of node.
//
func (leaf *Node) update(newInfo os.FileInfo, withContent bool) (err error) {
	if newInfo == nil {
		newInfo, err = os.Stat(leaf.SysPath)
		if err != nil {
			return fmt.Errorf("lib/memfs: Node.update %q: %s",
				leaf.Path, err.Error())
		}
	}

	if leaf.Mode != newInfo.Mode() {
		leaf.Mode = newInfo.Mode()
		return nil
	}

	leaf.Size = newInfo.Size()

	if !withContent || newInfo.IsDir() {
		return nil
	}

	return leaf.updateContent()
}

//
// updateContent read the content of file.
//
func (leaf *Node) updateContent() (err error) {
	if leaf.Size > MaxFileSize {
		return nil
	}

	leaf.V, err = ioutil.ReadFile(leaf.SysPath)
	if err != nil {
		return err
	}

	return nil
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
