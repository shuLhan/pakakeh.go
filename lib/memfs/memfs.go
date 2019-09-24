// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

//nolint:gochecknoglobals
var (
	// MaxFileSize define maximum file size that can be stored on memory.
	// The default value is 5 MB.
	MaxFileSize int64 = 1024 * 1024 * 5

	// Development define a flag to bypass file in memory.  If its
	// true, any call to Get will result in direct read to file system.
	Development bool

	// GeneratedPathNode contains the mapping of path and node.  Its will
	// be used and initialized by ".go" file generated from GoGenerate().
	GeneratedPathNode *PathNode
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
func New(includes, excludes []string, withContent bool) (mfs *MemFS, err error) {
	if GeneratedPathNode != nil {
		if !Development {
			mfs = &MemFS{
				pn:   GeneratedPathNode,
				root: GeneratedPathNode.Get("/"),
			}
			return mfs, nil
		}
	}

	mfs = &MemFS{
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
		return fmt.Errorf("ContentEncode: invalid encoding " + encoding)
	}

	for _, node := range mfs.pn.v {
		if node.Mode.IsDir() || len(node.V) == 0 {
			continue
		}

		_, err = encoder.Write(node.V)
		if err != nil {
			return fmt.Errorf("ContentEncode: " + err.Error())
		}

		err = encoder.Close()
		if err != nil {
			return fmt.Errorf("ContentEncode: " + err.Error())
		}

		node.V = make([]byte, buf.Len())
		copy(node.V, buf.Bytes())

		node.ContentEncoding = encoding
		node.Size = int64(len(node.V))

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
	node = mfs.pn.Get(path)
	if node == nil {
		if Development {
			node, err = mfs.refresh(path)
			if err != nil {
				log.Println("lib/memfs: Get: " + err.Error())
				return nil, os.ErrNotExist
			}
			return node, nil
		}
		return nil, os.ErrNotExist
	}

	if Development {
		err = node.update(nil, mfs.withContent)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

//
// ListNames list all files in memory sorted by name.
//
func (mfs *MemFS) ListNames() (paths []string) {
	paths = make([]string, 0, len(mfs.pn.f)+len(mfs.pn.v))

	for k := range mfs.pn.f {
		paths = append(paths, k)
	}

	for k := range mfs.pn.v {
		_, ok := mfs.pn.f[k]
		if !ok {
			paths = append(paths, k)
		}
	}

	sort.Strings(paths)

	return paths
}

//
// IsMounted will return true if a directory in file system has been mounted
// to memory; otherwise it will return false.
//
func (mfs *MemFS) IsMounted() bool {
	return mfs.root != nil
}

//
// Mount the directory recursively into the memory as root directory.
// For example, if we mount directory "/tmp" and "/tmp" contains file "a", to
// access file "a" we call Get("/a"), not Get("/tmp/a").
//
// Mount does not have any effect if current directory contains ".go"
// file generated from GoGenerate().
//
func (mfs *MemFS) Mount(dir string) error {
	if len(dir) == 0 {
		return nil
	}
	if GeneratedPathNode != nil {
		if !Development {
			return nil
		}
	}

	if mfs.pn == nil {
		mfs.pn = &PathNode{
			v: make(map[string]*Node),
			f: nil,
		}
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

//
// Unmount the root directory from memory.
//
func (mfs *MemFS) Unmount() {
	mfs.root = nil
	mfs.pn = nil
}

//
// Update the node content and information in memory based on new file
// information.
// This method only check if the node name is equal with file name, but it's
// not checking whether the node is part of memfs (node is parent or have the
// same root node).
//
func (mfs *MemFS) Update(node *Node, newInfo os.FileInfo) {
	if node == nil || newInfo == nil {
		return
	}

	err := node.update(newInfo, mfs.withContent)
	if err != nil {
		log.Println("lib/memfs: Update: " + err.Error())
	}
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
		ModTime: fi.ModTime(),
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
		leaf, err := mfs.AddChild(parent, fi)
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

//
// AddChild add new child to parent node.
//
func (mfs *MemFS) AddChild(parent *Node, fi os.FileInfo) (child *Node, err error) {
	sysPath := filepath.Join(parent.SysPath, fi.Name())

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

	if !mfs.isIncluded(sysPath, fi.Mode()) {
		return nil, nil
	}

	child, err = NewNode(parent, fi, mfs.withContent)
	if err != nil {
		log.Printf("memfs: AddChild %s: %s", fi.Name(), err.Error())
		return nil, nil
	}

	child.SysPath = sysPath

	parent.Childs = append(parent.Childs, child)

	mfs.pn.v[child.Path] = child

	return child, nil
}

//
// RemoveChild remove a child on parent, including its map on PathNode.
// If child is not part if node's childrens it will return nil.
//
func (mfs *MemFS) RemoveChild(parent *Node, child *Node) (removed *Node) {
	removed = parent.removeChild(child)
	if removed != nil {
		delete(mfs.pn.v, removed.Path)
	}
	return
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

//
// refresh the tree by rescanning from the root.
//
func (mfs *MemFS) refresh(url string) (node *Node, err error) {
	syspath := filepath.Join(mfs.root.SysPath, url)

	_, err = os.Stat(syspath)
	if err != nil {
		return nil, err
	}

	// Path exist on file system, try to refresh directory.
	f, err := os.Open(mfs.root.SysPath)
	if err != nil {
		return nil, err
	}

	err = mfs.scanDir(mfs.root, f)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	node = mfs.pn.Get(url)
	if node == nil {
		return nil, os.ErrNotExist
	}

	return node, nil
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

	for _, node := range mfs.pn.v {
		if node.Mode.IsDir() {
			continue
		}

		if !strings.HasPrefix(node.ContentType, "text/") {
			continue
		}

		if len(node.lowerv) == 0 {
			err := node.decode()
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
			indexes := libbytes.Indexes(node.lowerv, token)
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
