// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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

//
// List of valid content encoding for ContentEncode().
//
const (
	EncodingGzip = "gzip"
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
		PathNodes: &PathNode{
			v: make(map[string]*Node),
		},
		Root: &Node{
			SysPath: "..",
			Path:    "/",
			mode:    2147484141,
		},
		Opts: &Options{},
	}

	merged.PathNodes.v["/"] = merged.Root

	for _, mfs := range params {
		for _, child := range mfs.Root.Childs {
			_, exist := merged.PathNodes.v[child.Path]
			if exist {
				continue
			}
			merged.Root.AddChild(child)
		}
		for path, node := range mfs.PathNodes.v {
			if path == "/" {
				continue
			}
			_, exist := merged.PathNodes.v[path]
			if !exist {
				merged.PathNodes.v[path] = node
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
	if opts == nil {
		opts = &Options{}
	}
	opts.init()

	mfs = &MemFS{
		PathNodes: &PathNode{
			v: make(map[string]*Node),
			f: nil,
		},
		Opts: opts,
	}

	for _, inc := range opts.Includes {
		re, err := regexp.Compile(inc)
		if err != nil {
			return nil, fmt.Errorf("memfs.New: %w", err)
		}
		mfs.incRE = append(mfs.incRE, re)
	}
	for _, exc := range opts.Excludes {
		re, err := regexp.Compile(exc)
		if err != nil {
			return nil, fmt.Errorf("memfs.New: %w", err)
		}
		mfs.excRE = append(mfs.excRE, re)
	}

	err = mfs.mount()
	if err != nil {
		return nil, fmt.Errorf("memfs.New: %w", err)
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

	mfs.PathNodes.v[child.Path] = child

	return child, nil
}

//
// AddFile add the external file directly as internal file.
// If the internal file is already exist it will be replaced.
// Any directories in the internal path will be generated automatically if its
// not exist.
//
func (mfs *MemFS) AddFile(internalPath, externalPath string) (*Node, error) {
	if len(internalPath) == 0 {
		return nil, nil
	}
	fi, err := os.Stat(externalPath)
	if err != nil {
		return nil, fmt.Errorf("memfs.AddFile: %w", err)
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

		mfs.PathNodes.v[node.Path] = node

		parent = node
	}

	path = filepath.Join(path, base)
	node := &Node{
		SysPath: externalPath,
		Path:    path,
		name:    base,
		modTime: fi.ModTime(),
		mode:    fi.Mode(),
		size:    fi.Size(),
		Parent:  parent,
	}
	node.generateFuncName(path)

	if mfs.Opts.MaxFileSize <= 0 {
		node.size = 0
		return node, nil
	}

	err = node.updateContent(mfs.Opts.MaxFileSize)
	if err != nil {
		return nil, err
	}

	err = node.updateContentType()
	if err != nil {
		return nil, err
	}

	parent.Childs = append(parent.Childs, node)
	mfs.PathNodes.v[node.Path] = node

	return node, nil
}

//
// ContentEncode encode each node's content into specific encoding, in other
// words this method can be used to compress the content of file in memory
// or before being served or written.
//
// Only file with size greater than 0 will be encoded.
//
// List of known encoding is "gzip".
//
func (mfs *MemFS) ContentEncode(encoding string) (err error) {
	var (
		buf     bytes.Buffer
		encoder io.WriteCloser
	)

	encoding = strings.ToLower(encoding)

	switch encoding {
	case EncodingGzip:
		encoder = gzip.NewWriter(&buf)
	default:
		return fmt.Errorf("memfs.ContentEncode: invalid encoding " + encoding)
	}

	for _, node := range mfs.PathNodes.v {
		if node.mode.IsDir() || len(node.V) == 0 {
			continue
		}

		_, err = encoder.Write(node.V)
		if err != nil {
			return fmt.Errorf("memfs.ContentEncode: %w", err)
		}

		err = encoder.Close()
		if err != nil {
			return fmt.Errorf("memfs.ContentEncode: %w", err)
		}

		node.V = make([]byte, buf.Len())
		copy(node.V, buf.Bytes())

		node.ContentEncoding = encoding
		node.size = int64(len(node.V))

		buf.Reset()

		if encoding == EncodingGzip {
			gziper := encoder.(*gzip.Writer)
			gziper.Reset(&buf)
		}
	}

	return nil
}

//
// Get the node representation of file in memory.  If path is not exist it
// will return os.ErrNotExist.
//
func (mfs *MemFS) Get(path string) (node *Node, err error) {
	node = mfs.PathNodes.Get(path)
	if node == nil {
		if mfs.Opts.Development {
			node, err = mfs.refresh(path)
			if err != nil {
				log.Println("lib/memfs: Get: " + err.Error())
				return nil, fmt.Errorf("memfs.Get: %w", os.ErrNotExist)
			}
			return node, nil
		}
		return nil, os.ErrNotExist
	}

	if mfs.Opts.Development {
		err = node.update(nil, mfs.Opts.MaxFileSize)
		if err != nil {
			return nil, fmt.Errorf("memfs.Get: %w", err)
		}
	}

	return node, nil
}

//
// ListNames list all files in memory sorted by name.
//
func (mfs *MemFS) ListNames() (paths []string) {
	paths = make([]string, 0, len(mfs.PathNodes.f)+len(mfs.PathNodes.v))

	for k := range mfs.PathNodes.f {
		paths = append(paths, k)
	}

	for k := range mfs.PathNodes.v {
		_, ok := mfs.PathNodes.f[k]
		if !ok {
			paths = append(paths, k)
		}
	}

	sort.Strings(paths)

	return paths
}

//
// MarshalJSON encode the MemFS object into JSON format.
//
// The field that being encoded is the PathNode field, where the key sorted in
// ascending order.
//
func (mfs *MemFS) MarshalJSON() ([]byte, error) {
	return mfs.PathNodes.MarshalJSON()
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
		delete(mfs.PathNodes.v, removed.Path)
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

	for _, node := range mfs.PathNodes.v {
		if node.mode.IsDir() {
			continue
		}

		if !strings.HasPrefix(node.ContentType, "text/") {
			continue
		}

		if len(node.lowerv) == 0 {
			_, err := node.Decode()
			if err != nil {
				log.Printf("memfs.Search: " + err.Error())
				continue
			}

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
	if node == nil || newInfo == nil {
		return
	}

	err := node.update(newInfo, mfs.Opts.MaxFileSize)
	if err != nil {
		log.Println("lib/memfs: Update: " + err.Error())
	}
}

func (mfs *MemFS) createRoot(f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("%q must be a directory", mfs.Opts.Root)
	}

	mfs.Root = &Node{
		SysPath: mfs.Opts.Root,
		Path:    "/",
		name:    "/",
		modTime: fi.ModTime(),
		mode:    fi.Mode(),
		size:    0,
		V:       nil,
		Parent:  nil,
	}
	mfs.Root.generateFuncName(mfs.Opts.Root)

	mfs.PathNodes.v[mfs.Root.Path] = mfs.Root

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
	if len(mfs.incRE) > 0 {
		for _, re := range mfs.incRE {
			if re.MatchString(sysPath) {
				return true
			}
		}
		return mode.IsDir()
	}

	return true
}

//
// mount the directory recursively into the memory as root directory.
// For example, if we mount directory "/tmp" and "/tmp" contains file "a", to
// access file "a" we call Get("/a"), not Get("/tmp/a").
//
// mount does not have any effect if current directory contains ".go"
// file generated from GoGenerate().
//
func (mfs *MemFS) mount() error {
	if len(mfs.Opts.Root) == 0 {
		return nil
	}

	if mfs.PathNodes == nil {
		mfs.PathNodes = &PathNode{
			v: make(map[string]*Node),
			f: nil,
		}
	}

	f, err := os.Open(mfs.Opts.Root)
	if err != nil {
		return fmt.Errorf("mount: %w", err)
	}

	err = mfs.createRoot(f)
	if err != nil {
		return fmt.Errorf("mount: %w", err)
	}

	err = mfs.scanDir(mfs.Root, f)
	_ = f.Close()
	if err != nil {
		return fmt.Errorf("mount: %w", err)
	}

	if mfs.Opts.MaxFileSize > 0 {
		mfs.pruneEmptyDirs()
	}

	return nil
}

func (mfs *MemFS) scanDir(parent *Node, f *os.File) error {
	fis, err := f.Readdir(0)
	if err != nil {
		return err
	}

	sort.SliceStable(fis, func(x, y int) bool {
		return fis[x].Name() < fis[y].Name()
	})

	for _, fi := range fis {
		leaf, err := mfs.AddChild(parent, fi)
		if err != nil {
			return err
		}
		if leaf == nil {
			continue
		}
		if !leaf.mode.IsDir() {
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

//
// pruneEmptyDirs remove node that is directory and does not have childs.
//
func (mfs *MemFS) pruneEmptyDirs() {
	for k, node := range mfs.PathNodes.v {
		if !node.mode.IsDir() {
			continue
		}
		if len(node.Childs) != 0 {
			continue
		}
		if node.Parent == nil {
			continue
		}

		node.Parent.removeChild(node)
		delete(mfs.PathNodes.v, k)
	}
}

//
// refresh the tree by rescanning from the root.
//
func (mfs *MemFS) refresh(url string) (node *Node, err error) {
	syspath := filepath.Join(mfs.Root.SysPath, url)

	_, err = os.Stat(syspath)
	if err != nil {
		return nil, err
	}

	// Path exist on file system, try to refresh directory.
	f, err := os.Open(mfs.Root.SysPath)
	if err != nil {
		return nil, err
	}

	err = mfs.scanDir(mfs.Root, f)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
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
