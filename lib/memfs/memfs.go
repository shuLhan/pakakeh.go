// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	MaxFileSize int64 = 1024 * 1024 * 5 //nolint: gochecknoglobals

	// Development define a flag to bypass file in memory.  If its
	// true, any call to Get will result in direct read to file system.
	Development bool //nolint: gochecknoglobals

	// GeneratedPathNode contains the mapping of path and node.  Its will
	// be used and initialized by ".go" file generated from GoGenerate().
	GeneratedPathNode *PathNode //nolint: gochecknoglobals
)

//
// MemFS contains directory tree of file system in memory.
//
type MemFS struct {
	incRE       []*regexp.Regexp
	excRE       []*regexp.Regexp
	root        *Node
	pn          *PathNode
	withContent bool
}

//
// New create and initialize new memory file system using list of regular
// expresssion for including or excluding files.
//
// The includes and excludes pattern applied to path of file in file system,
// not to the path in memory.
//
// The "withContent" parameter tell the MemFS to read the content of file and
// detect its content type.  If this paramater is false, the content of file
// will not be mapped to memory, the MemFS will behave as directory tree.
//
// On directory that contains output from GoGenerate(), the includes and
// excludes does not have any effect, since the content of path and nodes will
// be overwritten by GeneratedPathNode.
//
func New(includes, excludes []string, withContent bool) (*MemFS, error) {
	if !Development && GeneratedPathNode != nil {
		mfs := &MemFS{
			pn: GeneratedPathNode,
		}
		return mfs, nil
	}

	mfs := &MemFS{
		pn: &PathNode{
			v: make(map[string]*Node),
			f: nil,
		},
		withContent: withContent,
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
// Get the node representation of file in memory.  If path is not exist it
// will return os.ErrNotExist.
//
func (mfs *MemFS) Get(path string) (*Node, error) {
	node := mfs.pn.Get(path)
	if node == nil {
		return nil, os.ErrNotExist
	}

	if Development {
		path = filepath.Join("/", path)
		path = filepath.Join(mfs.root.SysPath, path)
		return newNode(path)
	}

	return node, nil
}

//
// ListNames list all files in memory sorted by name.
//
func (mfs *MemFS) ListNames() (paths []string) {
	if len(mfs.pn.v) > 0 {
		paths = make([]string, 0, len(mfs.pn.v))
	}

	for k := range mfs.pn.v {
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
// Mount the directory recursively into the memory as root directory.
// For example, if we mount directory "/tmp" and "/tmp" contains file "a", to
// access file "a" we call Get("/a"), not Get("/tmp/a").
//
// Mount does not have any effect if current directory contains ".go"
// generated file from GoGenerate().
//
func (mfs *MemFS) Mount(dir string) error {
	if len(dir) == 0 {
		return nil
	}
	if !Development && GeneratedPathNode != nil {
		return nil
	}

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

	if mfs.withContent {
		mfs.pruneEmptyDirs()
	}

	return nil
}

func (mfs *MemFS) createRoot(dir string, f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("mount must be a directory")
	}

	mfs.root = &Node{
		SysPath: dir,
		Path:    "/",
		Name:    "/",
		Mode:    fi.Mode(),
		Size:    fi.Size(),
		V:       nil,
		Parent:  nil,
	}

	mfs.pn.v[mfs.root.Path] = mfs.root

	return nil
}

func (mfs *MemFS) scanDir(parent *Node, f *os.File) error {
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
		if !leaf.Mode.IsDir() {
			continue
		}

		fdir, err := os.Open(leaf.SysPath)
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

func (mfs *MemFS) addChild(parent *Node, fi os.FileInfo) (*Node, error) {
	var err error

	if fi.Mode()&os.ModeSymlink != 0 {
		symPath := filepath.Join(parent.SysPath, fi.Name())
		absPath, err := filepath.EvalSymlinks(symPath)
		if err != nil {
			return nil, err
		}

		fi, err = os.Lstat(absPath)
		if err != nil {
			return nil, err
		}
	}

	child := &Node{
		Mode:   fi.Mode(),
		Size:   fi.Size(),
		Parent: parent,
	}
	child.Name = fi.Name()
	child.SysPath = filepath.Join(parent.SysPath, child.Name)
	child.Path = path.Join(parent.Path, child.Name)

	if !mfs.isIncluded(child) {
		return nil, nil
	}

	parent.Childs = append(parent.Childs, child)

	mfs.pn.v[child.Path] = child

	if child.Mode.IsDir() {
		return child, nil
	}

	if !mfs.withContent {
		return child, nil
	}

	err = child.updateContentType()
	if err != nil {
		return nil, err
	}

	if child.Size > MaxFileSize {
		return child, nil
	}

	child.V, err = ioutil.ReadFile(child.SysPath)
	if err != nil {
		return nil, err
	}

	return child, nil
}

//
// isIncluded will return true if the child node pass the included filter or
// excluded filter; otherwise it will return false.
//
func (mfs *MemFS) isIncluded(child *Node) bool {
	if len(mfs.incRE) == 0 && len(mfs.excRE) == 0 {
		return true
	}
	for _, re := range mfs.excRE {
		if re.MatchString(child.SysPath) {
			return false
		}
	}
	if len(mfs.incRE) > 0 {
		for _, re := range mfs.incRE {
			if re.MatchString(child.SysPath) {
				return true
			}
		}
		return child.Mode.IsDir()
	}

	return true
}

//
// pruneEmptyDirs remove node that is directory and does not have childs.
//
func (mfs *MemFS) pruneEmptyDirs() {
	for k, node := range mfs.pn.v {
		if !node.Mode.IsDir() {
			continue
		}
		if len(node.Childs) != 0 {
			continue
		}
		if node.Parent == nil {
			continue
		}

		node.Parent.removeChild(node)
		delete(mfs.pn.v, k)
	}
}
