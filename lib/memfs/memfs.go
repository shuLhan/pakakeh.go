// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	libints "github.com/shuLhan/share/lib/ints"
	"github.com/shuLhan/share/lib/sanitize"
	libstrings "github.com/shuLhan/share/lib/strings"
)

const (
	defContentType = "text/plain" // Default content type for empty file.
)

//
// MemFS contains directory tree of file system in memory.
//
type MemFS struct {
	http.FileSystem

	PathNodes *PathNode
	Root      *Node
	Opts      *Options
	incRE     []*regexp.Regexp
	excRE     []*regexp.Regexp
}

//
// Merge one or more instances of MemFS into single hierarchy.
//
// If there are two instance of Node that have the same path, the last
// instance will be ignored.
//
func Merge(params ...*MemFS) (merged *MemFS) {
	merged = &MemFS{
		PathNodes: NewPathNode(),
		Root: &Node{
			SysPath: "..",
			Path:    "/",
			mode:    2147484141,
		},
		Opts: &Options{},
	}

	merged.PathNodes.Set("/", merged.Root)

	for _, mfs := range params {
		for _, child := range mfs.Root.Childs {
			gotNode := merged.PathNodes.Get(child.Path)
			if gotNode != nil {
				continue
			}
			merged.Root.AddChild(child)
		}
		paths := mfs.PathNodes.Paths()
		for _, path := range paths {
			if path == "/" {
				continue
			}
			gotNode := merged.PathNodes.Get(path)
			if gotNode == nil {
				merged.PathNodes.Set(path, mfs.PathNodes.Get(path))
			}
		}
	}
	return merged
}

//
// New create and initialize new memory file system from directory Root using
// list of regular expresssion for Including or Excluding files.
//
func New(opts *Options) (mfs *MemFS, err error) {
	logp := "New"

	mfs = &MemFS{
		Opts: opts,
	}

	err = mfs.Init()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return mfs, nil
}

//
// AddChild add new child to parent node.
//
func (mfs *MemFS) AddChild(parent *Node, fi os.FileInfo) (child *Node, err error) {
	sysPath := filepath.Join(parent.SysPath, fi.Name())

	if !mfs.isIncluded(sysPath, fi.Mode()) {
		return nil, nil
	}

	child, err = parent.addChild(sysPath, fi, mfs.Opts.MaxFileSize)
	if err != nil {
		log.Printf("AddChild %s: %s", fi.Name(), err.Error())
		return nil, nil
	}

	mfs.PathNodes.Set(child.Path, child)

	return child, nil
}

//
// AddFile add the external file directly as internal file.
// If the internal file is already exist it will be replaced.
// Any directories in the internal path will be generated automatically if its
// not exist.
//
func (mfs *MemFS) AddFile(internalPath, externalPath string) (node *Node, err error) {
	if len(internalPath) == 0 {
		return nil, nil
	}

	logp := "AddFile"

	fi, err := os.Stat(externalPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	var parent *Node

	internalPath = filepath.ToSlash(filepath.Clean(internalPath))
	paths := strings.Split(internalPath, "/")
	base := paths[len(paths)-1]
	paths = paths[:len(paths)-1]
	path := ""

	for _, p := range paths {
		path = filepath.Join(path, p)
		node, _ := mfs.Get(path)
		if node != nil {
			parent = node
			continue
		}

		node = &Node{
			SysPath: path,
			Path:    path,
			name:    p,
			mode:    os.ModeDir,
			Parent:  parent,
		}
		node.generateFuncName(path)

		if parent == nil {
			mfs.Root.Childs = append(mfs.Root.Childs, node)
		} else {
			parent.Childs = append(parent.Childs, node)
		}

		mfs.PathNodes.Set(node.Path, node)

		parent = node
	}

	path = filepath.Join(path, base)
	node = &Node{
		SysPath: externalPath,
		Path:    path,
		name:    base,
		modTime: fi.ModTime(),
		mode:    fi.Mode(),
		size:    fi.Size(),
		Parent:  parent,
	}
	node.generateFuncName(path)

	err = node.updateContent(mfs.Opts.MaxFileSize)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	err = node.updateContentType()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	parent.Childs = append(parent.Childs, node)
	mfs.PathNodes.Set(node.Path, node)

	return node, nil
}

//
// Get the node representation of file in memory.  If path is not exist it
// will return os.ErrNotExist.
//
func (mfs *MemFS) Get(path string) (node *Node, err error) {
	logp := "Get"

	node = mfs.PathNodes.Get(path)
	if node == nil {
		if mfs.Opts.Development {
			node, err = mfs.refresh(path)
			if err != nil {
				return nil, fmt.Errorf("%s: %s: %w", logp, path, err)
			}
			return node, nil
		}
		return nil, os.ErrNotExist
	}

	if mfs.Opts.Development {
		err = node.Update(nil, mfs.Opts.MaxFileSize)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", logp, path, err)
		}
	}

	return node, nil
}

//
// Init initialize the MemFS instance.
// This method provided to initialize MemFS if its Options is set directly,
// not through New() function.
//
func (mfs *MemFS) Init() (err error) {
	var (
		logp = "Init"
		v    string
		re   *regexp.Regexp
	)

	if mfs.Opts == nil {
		mfs.Opts = &Options{}
	}
	mfs.Opts.init()

	if mfs.PathNodes == nil {
		mfs.PathNodes = NewPathNode()
	}

	for _, v = range mfs.Opts.Includes {
		re, err = regexp.Compile(v)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
		mfs.incRE = append(mfs.incRE, re)
	}
	for _, v = range mfs.Opts.Excludes {
		re, err = regexp.Compile(v)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
		mfs.excRE = append(mfs.excRE, re)
	}

	err = mfs.mount()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

//
// ListNames list all files in memory sorted by name.
//
func (mfs *MemFS) ListNames() (paths []string) {
	paths = mfs.PathNodes.Paths()
	return paths
}

//
// MarshalJSON encode the MemFS object into JSON format.
//
// The field that being encoded is the Root node.
//
func (mfs *MemFS) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	mfs.Root.packAsJson(&buf, 0)
	return buf.Bytes(), nil
}

//
// MustGet return the Node representation of file in memory by its path if its
// exist or nil the path is not exist.
//
func (mfs *MemFS) MustGet(path string) (node *Node) {
	node, _ = mfs.Get(path)
	return node
}

//
// Open the named file for reading.
// This is an alias to Get() method, to make it implement http.FileSystem.
//
func (mfs *MemFS) Open(path string) (http.File, error) {
	return mfs.Get(path)
}

//
// RemoveChild remove a child on parent, including its map on PathNode.
// If child is not part if node's childrens it will return nil.
//
func (mfs *MemFS) RemoveChild(parent *Node, child *Node) (removed *Node) {
	removed = parent.removeChild(child)
	if removed != nil {
		mfs.PathNodes.Delete(removed.Path)
	}
	return
}

//
// Search one or more strings in each content of files.
//
func (mfs *MemFS) Search(words []string, snippetLen int) (results []SearchResult) {
	if len(words) == 0 {
		return nil
	}
	if snippetLen <= 0 {
		snippetLen = 60
	}

	tokens := libstrings.ToBytes(words)
	for x := 0; x < len(tokens); x++ {
		tokens[x] = bytes.ToLower(tokens[x])
	}

	nodes := mfs.PathNodes.Nodes()
	for _, node := range nodes {
		if node.mode.IsDir() {
			continue
		}

		if !strings.HasPrefix(node.ContentType, "text/") {
			continue
		}

		if len(node.lowerv) == 0 {
			node.plainv = node.Content

			if strings.HasPrefix(node.ContentType, "text/html") {
				node.plainv = sanitize.HTML(node.plainv)
			}

			node.lowerv = bytes.ToLower(node.plainv)
		}

		result := SearchResult{
			Path: node.Path,
		}

		var allIndexes []int
		for _, token := range tokens {
			indexes := libbytes.WordIndexes(node.lowerv, token)
			allIndexes = append(allIndexes, indexes...)
		}
		if len(allIndexes) == 0 {
			continue
		}

		allIndexes = libints.MergeByDistance(allIndexes, nil, snippetLen)
		snippets := libbytes.SnippetByIndexes(node.lowerv, allIndexes, snippetLen)
		for _, snippet := range snippets {
			result.Snippets = append(result.Snippets, string(snippet))
		}

		results = append(results, result)
	}

	return results
}

//
// Update the node content and information in memory based on new file
// information.
// This method only check if the node name is equal with file name, but it's
// not checking whether the node is part of memfs (node is parent or have the
// same Root node).
//
func (mfs *MemFS) Update(node *Node, newInfo os.FileInfo) {
	if node == nil {
		return
	}

	err := node.Update(newInfo, mfs.Opts.MaxFileSize)
	if err != nil {
		log.Printf("Update: %s: %s", node.SysPath, err)
	}
}

func (mfs *MemFS) createRoot() error {
	logp := "createRoot"

	fi, err := os.Stat(mfs.Opts.Root)
	if err != nil {
		return fmt.Errorf("%s: %s: %w", logp, mfs.Opts.Root, err)
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s: %s must be a directory", logp, mfs.Opts.Root)
	}

	mfs.Root = &Node{
		SysPath: mfs.Opts.Root,
		Path:    "/",
		name:    "/",
		modTime: fi.ModTime(),
		mode:    fi.Mode(),
	}
	mfs.Root.generateFuncName(mfs.Opts.Root)

	mfs.PathNodes.Set(mfs.Root.Path, mfs.Root)

	return nil
}

//
// isIncluded will return true if the child node pass the included filter or
// excluded filter; otherwise it will return false.
//
func (mfs *MemFS) isIncluded(sysPath string, mode os.FileMode) bool {
	if len(mfs.incRE) == 0 && len(mfs.excRE) == 0 {
		return true
	}
	for _, re := range mfs.excRE {
		if re.MatchString(sysPath) {
			return false
		}
	}
	if len(mfs.incRE) == 0 {
		// No filter defined, default to always included.
		return true
	}

	for _, re := range mfs.incRE {
		if re.MatchString(sysPath) {
			return true
		}
	}
	if mode&os.ModeSymlink == 0 {
		// If file is NOT a symlink andits a directory, include it.
		return mode.IsDir()
	}

	// File is symlink, get the real FileInfo to check if its
	// directory or not.
	absPath, err := filepath.EvalSymlinks(sysPath)
	if err != nil {
		return false
	}

	fi, err := os.Lstat(absPath)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return true
	}
	return false
}

//
// mount the directory recursively into the memory as root directory.
// For example, if we mount directory "/tmp" and "/tmp" contains file "a", to
// access file "a" we call Get("/a"), not Get("/tmp/a").
//
// mount does not have any effect if current directory contains ".go"
// file generated from GoEmbed().
//
func (mfs *MemFS) mount() (err error) {
	if len(mfs.Opts.Root) == 0 {
		return nil
	}
	if mfs.Root != nil {
		// The directory has been initialized by embedded.
		return nil
	}

	logp := "mount"

	if mfs.PathNodes == nil {
		mfs.PathNodes = NewPathNode()
	}

	err = mfs.createRoot()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	_, err = mfs.scanDir(mfs.Root)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

// scanDir scan the directory node for files and add them to memory file
// system.
// It returns number of childs added to the node or an error.
func (mfs *MemFS) scanDir(node *Node) (n int, err error) {
	var (
		logp    = "scanDir"
		child   *Node
		nchilds int
		f       *os.File
		fi      os.FileInfo
		fis     []os.FileInfo
	)

	f, err = os.Open(node.SysPath)
	if err != nil {
		return 0, fmt.Errorf("%s: %s: %w", logp, node.SysPath, err)
	}

	fis, err = f.Readdir(0)
	if err != nil {
		return 0, fmt.Errorf("%s: %s: %w", logp, node.SysPath, err)
	}

	sort.SliceStable(fis, func(x, y int) bool {
		return fis[x].Name() < fis[y].Name()
	})

	for _, fi = range fis {
		child, err = mfs.AddChild(node, fi)
		if err != nil {
			err = fmt.Errorf("%s: %s: %w", logp, node.SysPath, err)
			goto out
		}
		if child == nil {
			continue
		}
		n++
		if !child.mode.IsDir() {
			continue
		}

		nchilds, err = mfs.scanDir(child)
		if err != nil {
			err = fmt.Errorf("%s: %s: %w", logp, node.SysPath, err)
			goto out
		}
		if nchilds == 0 {
			// No childs added, remove it from node.
			mfs.RemoveChild(node, child)
			n--
		}
	}
out:
	errClose := f.Close()
	if errClose != nil {
		if err == nil {
			err = fmt.Errorf("%s: %s: %w", logp, node.SysPath, errClose)
		} else {
			log.Printf("%s: %s: %s", logp, node.SysPath, errClose)
		}
	}

	return n, err
}

//
// refresh the tree by rescanning from the root.
//
func (mfs *MemFS) refresh(url string) (node *Node, err error) {
	logp := "refresh"
	syspath := filepath.Join(mfs.Root.SysPath, url)

	_, err = os.Stat(syspath)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", logp, url, err)
	}

	_, err = mfs.scanDir(mfs.Root)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", logp, url, err)
	}

	node = mfs.PathNodes.Get(url)
	if node == nil {
		return nil, os.ErrNotExist
	}

	return node, nil
}

//
// resetAllModTime set the modTime on Root and its childs to the t.
// This method is only intended for testing.
//
func (mfs *MemFS) resetAllModTime(t time.Time) {
	mfs.Root.resetAllModTime(t)
}
