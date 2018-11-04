// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package memfs provide a library for mapping fily system into memory.
//
package memfs

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

var (
	// MaxFileSize define maximum file size that can be stored on memory.
	// The default value is 5 MB.
	MaxFileSize int64 = 1024 * 1024 * 5
)

//
// MemFS contains the configuration and content of memory file system.
//
type MemFS struct {
	incRE       []*regexp.Regexp
	excRE       []*regexp.Regexp
	root        *node
	mapPathNode map[string]*node
}

//
// New create and initialize new memory file system using list of regular
// expresssion for including or excluding files.
// The includes and excludes pattern applied to path of file in file system,
// not to the path in memory.
//
func New(includes, excludes []string) (*MemFS, error) {
	mfs := &MemFS{
		mapPathNode: make(map[string]*node),
	}
	for _, inc := range includes {
		re, err := regexp.Compile(inc)
		if err != nil {
			return nil, err
		}
		mfs.incRE = append(mfs.incRE, re)
	}
	for _, exc := range excludes {
		re, err := regexp.Compile(exc)
		if err != nil {
			return nil, err
		}
		mfs.excRE = append(mfs.excRE, re)
	}

	return mfs, nil
}

//
// Get the content of file in path.  If path is not exist it will return
// os.ErrNotExist.
//
func (mfs *MemFS) Get(path string) ([]byte, error) {
	node, ok := mfs.mapPathNode[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	if node.mode.IsDir() {
		return nil, nil
	}
	if len(node.v) > 0 {
		return node.v, nil
	}

	v, err := ioutil.ReadFile(node.sysPath)

	return v, err
}

//
// ListNames list all files in memory sorted by name.
//
func (mfs *MemFS) ListNames() (paths []string) {
	for k, _ := range mfs.mapPathNode {
		if len(paths) == 0 {
			paths = append(paths, k)
			continue
		}

		x := 0
		for ; x < len(paths); x++ {
			if k > paths[x] {
				continue
			}
			break
		}
		if x == len(paths) {
			paths = append(paths, k)
		} else {
			paths = append(paths, k)
			copy(paths[x+1:], paths[x:])
			paths[x] = k
		}
	}
	return
}

//
// Mount the source directory recursively into the memory root directory.
//
func (mfs *MemFS) Mount(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}

	err = mfs.createRoot(dir, f)
	if err != nil {
		return err
	}

	err = mfs.scanDir(mfs.root, f)
	_ = f.Close()
	if err != nil {
		return err
	}

	mfs.pruneEmptyDirs()

	return err
}

func (mfs *MemFS) createRoot(dir string, f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("Mount must be a directory.")
	}

	mfs.root = &node{
		sysPath: dir,
		path:    "/",
		name:    "/",
		mode:    fi.Mode(),
		size:    fi.Size(),
		v:       nil,
		parent:  nil,
	}

	mfs.mapPathNode[mfs.root.path] = mfs.root

	return nil
}

func (mfs *MemFS) scanDir(parent *node, f *os.File) error {
	fis, err := f.Readdir(0)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		leaf, err := mfs.addChild(parent, fi)
		if err != nil {
			return err
		}
		if leaf == nil {
			continue
		}
		if !leaf.mode.IsDir() {
			continue
		}

		fdir, err := os.Open(leaf.sysPath)
		if err != nil {
			return err
		}

		err = mfs.scanDir(leaf, fdir)
		_ = fdir.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (mfs *MemFS) addChild(parent *node, fi os.FileInfo) (*node, error) {
	child := &node{
		mode:   fi.Mode(),
		size:   fi.Size(),
		parent: parent,
	}
	child.name = fi.Name()
	child.sysPath = filepath.Join(parent.sysPath, child.name)
	child.path = path.Join(parent.path, child.name)

	if !mfs.isIncluded(child) {
		return nil, nil
	}

	parent.childs = append(parent.childs, child)

	mfs.mapPathNode[child.path] = child

	if child.mode.IsDir() {
		return child, nil
	}

	if child.size > MaxFileSize {
		return child, nil
	}

	var err error
	child.v, err = ioutil.ReadFile(child.sysPath)

	return child, err
}

//
// isIncluded will return true if the child node pass the included filter or
// excluded filter; otherwise it will return false.
//
func (mfs *MemFS) isIncluded(child *node) bool {
	if len(mfs.incRE) == 0 && len(mfs.excRE) == 0 {
		return true
	}
	for _, re := range mfs.excRE {
		if re.MatchString(child.sysPath) {
			return false
		}
	}
	if len(mfs.incRE) > 0 {
		for _, re := range mfs.incRE {
			if re.MatchString(child.sysPath) {
				return true
			}
		}
		if child.mode.IsDir() {
			return true
		}

		// Its neither excluded or included.
		return false
	}

	return true
}

//
// pruneEmptyDirs remove node that is directory and does not have childs.
//
func (mfs *MemFS) pruneEmptyDirs() {
	for k, node := range mfs.mapPathNode {
		if !node.mode.IsDir() {
			continue
		}
		if len(node.childs) != 0 {
			continue
		}
		if node.parent == nil {
			continue
		}

		node.parent.removeChild(node)
		delete(mfs.mapPathNode, k)
	}
}
