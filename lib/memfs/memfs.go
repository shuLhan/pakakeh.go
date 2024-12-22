// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package memfs

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	libhtml "git.sr.ht/~shulhan/pakakeh.go/lib/html"
	libints "git.sr.ht/~shulhan/pakakeh.go/lib/ints"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs/v2"
)

const (
	defContentType = "text/plain" // Default content type for empty file.
)

// MemFS contains directory tree of file system in memory.
type MemFS struct {
	PathNodes *PathNode
	Root      *Node
	Opts      *Options
	dw        *watchfs.DirWatcher

	// subfs contains another MemFS instances.
	// During Get, it will evaluated in order.
	subfs []*MemFS

	dwOptions watchfs.DirWatcherOptions
	watchopts WatchOptions
}

// New create and initialize new memory file system from directory Root using
// list of regular expresssion for Including or Excluding files.
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

// AddChild add FileInfo fi as new child of parent node.
//
// It will return nil without an error if,
//   - the system path of parent+fi.Name() is excluded by one of
//     Options.Excludes pattern, or
//   - fi is symlink to not existen node.
func (mfs *MemFS) AddChild(parent *Node, fi os.FileInfo) (child *Node, err error) {
	var (
		logp    = "AddChild"
		sysPath = filepath.Join(parent.SysPath, fi.Name())
	)

	if mfs.Opts.isExcluded(sysPath) {
		return nil, nil
	}
	if !mfs.Opts.isIncluded(sysPath, fi) {
		if child != nil {
			// The path being watched, but not included.
			// Set the generate function name to empty, to prevent
			// GoEmbed embed the content of this node.
			child.GenFuncName = ""
		}
		return child, nil
	}

	child, err = parent.addChild(sysPath, fi, mfs.Opts.MaxFileSize)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf(`%s %q: %w`, logp, sysPath, err)
	}

	mfs.PathNodes.Set(child.Path, child)

	return child, nil
}

// AddFile add the external file directly as internal file.
// If the internal file is already exist it will be replaced.
// Any directories in the internal path will be generated automatically if its
// not exist.
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
		node, _ = mfs.Get(path)
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

// Get the node representation of file in memory.
// If path is not exist it will return [fs.ErrNotExist].
func (mfs *MemFS) Get(path string) (node *Node, err error) {
	var logp = `Get`

	if mfs == nil || mfs.PathNodes == nil {
		return nil, fmt.Errorf(`%s %q: %w`, logp, path, fs.ErrNotExist)
	}
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return nil, fmt.Errorf(`%s: empty path`, logp)
	}
	if path != `/` {
		path = strings.TrimSuffix(path, `/`)
	}

	node = mfs.PathNodes.Get(path)
	if node != nil {
		if mfs.Opts.TryDirect {
			_ = node.Update(nil, mfs.Opts.MaxFileSize)

			// Ignore error if the file is not exist in storage.
			// Use case: the node maybe have been result of embed and the
			// merged with other MemFS instance that use TryDirect flag.
		}
		return node, nil
	}

	// Get node from sub fs.
	var sub *MemFS

	for _, sub = range mfs.subfs {
		node, _ = sub.Get(path)
		if node != nil {
			return node, nil
		}
	}

	// Refresh the root FS first.

	if mfs.Opts.TryDirect {
		node, err = mfs.refresh(path)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf(`%s %q: %w`, logp, path, err)
			}
		}
		if node != nil {
			return node, nil
		}
	}

	// Refresh the subfs.

	for _, sub = range mfs.subfs {
		if !sub.Opts.TryDirect {
			continue
		}
		node, err = sub.refresh(path)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf(`%s %q: %w`, logp, path, err)
			}
		}
		if node != nil {
			return node, nil
		}
	}

	return nil, fmt.Errorf(`%s %q: %w`, logp, path, fs.ErrNotExist)
}

// Init initialize the MemFS instance.
// This method provided to initialize MemFS if its Options is set directly,
// not through New() function.
func (mfs *MemFS) Init() (err error) {
	var logp = `Init`

	if mfs.Opts == nil {
		mfs.Opts = &Options{}
	}
	if mfs.PathNodes == nil {
		mfs.PathNodes = NewPathNode()
	}

	err = mfs.Opts.init()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = mfs.mount()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

// ListNames list all files in memory sorted by name.
func (mfs *MemFS) ListNames() (paths []string) {
	paths = mfs.PathNodes.Paths()
	return paths
}

// MarshalJSON encode the MemFS object into JSON format.
//
// The field that being encoded is the Root node.
func (mfs *MemFS) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	mfs.Root.packAsJSON(&buf, 0, true, true)
	return buf.Bytes(), nil
}

// Merge other MemFS instance as sub file system.
//
// When Get method called, each sub fs will be evaluated in order of Merge.
func (mfs *MemFS) Merge(sub *MemFS) {
	if sub == nil {
		return
	}
	mfs.subfs = append(mfs.subfs, sub)
}

// MustGet return the Node representation of file in memory by its path if its
// exist or nil the path is not exist.
func (mfs *MemFS) MustGet(path string) (node *Node) {
	node, _ = mfs.Get(path)
	return node
}

// Open the named file for reading.
// This is an alias to Get() method, to make it implement http.FileSystem.
func (mfs *MemFS) Open(path string) (http.File, error) {
	return mfs.Get(path)
}

// RemoveChild remove a child on parent, including its map on PathNode.
// If child is not part if node's childrens it will return nil.
func (mfs *MemFS) RemoveChild(parent *Node, child *Node) (removed *Node) {
	if parent != nil {
		removed = parent.removeChild(child)
		if removed != nil {
			mfs.PathNodes.Delete(removed.Path)
		}
	}
	return
}

// Search one or more strings in each content of files.
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
				node.plainv = libhtml.Sanitize(node.plainv)
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

// StopWatch stop watching for update, from calling Watch.
func (mfs *MemFS) StopWatch() {
	if mfs.dw != nil {
		mfs.dw.Stop()
		mfs.dw = nil
	}
}

// Update the node content and information in memory based on new file
// information.
// This method only check if the node name is equal with file name, but it's
// not checking whether the node is part of memfs (node is parent or have the
// same Root node).
func (mfs *MemFS) Update(node *Node, newInfo os.FileInfo) {
	if node == nil {
		return
	}

	var (
		logp = "MemFS.Update"
		err  error
	)

	err = node.Update(newInfo, mfs.Opts.MaxFileSize)
	if err != nil {
		log.Printf("%s %s: %s", logp, node.SysPath, err)
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

// mount the directory recursively into the memory as root directory.
// For example, if we mount directory "/tmp" and "/tmp" contains file "a", to
// access file "a" we call Get("/a"), not Get("/tmp/a").
//
// mount does not have any effect if current directory contains ".go"
// file generated from GoEmbed().
func (mfs *MemFS) mount() (err error) {
	if len(mfs.Opts.Root) == 0 {
		return nil
	}
	if mfs.Root != nil {
		// The Root node may has been initialized through embedding.
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

	err = mfs.scanDir(mfs.Root)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

// Remount reset the memfs instance to force rescanning the files again from
// file system.
func (mfs *MemFS) Remount() (err error) {
	mfs.Root = nil
	mfs.PathNodes = nil
	return mfs.mount()
}

// scanDir scan the content of node directory and add them to mfs.
// It returns number of childs added to the node or an error.
func (mfs *MemFS) scanDir(node *Node) (err error) {
	var (
		logp  = "scanDir"
		child *Node
		f     *os.File
		fi    os.FileInfo
		fis   []os.FileInfo
	)

	f, err = os.Open(node.SysPath)
	if err != nil {
		if os.IsPermission(err) {
			// Ignore error due to permission
			return nil
		}
		return fmt.Errorf(`%s %q: %w`, logp, node.SysPath, err)
	}

	fis, err = f.Readdir(0)
	if err != nil {
		return fmt.Errorf(`%s %q: %w`, logp, node.SysPath, err)
	}

	sort.SliceStable(fis, func(x, y int) bool {
		return fis[x].Name() < fis[y].Name()
	})

	for _, fi = range fis {
		child, err = mfs.AddChild(node, fi)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			err = fmt.Errorf(`%s %q: %w`, logp, node.SysPath, err)
			goto out
		}
		if child == nil {
			continue
		}
		if !child.mode.IsDir() {
			continue
		}

		err = mfs.scanDir(child)
		if err != nil {
			err = fmt.Errorf(`%s %q: %w`, logp, node.SysPath, err)
			goto out
		}
	}
out:
	errClose := f.Close()
	if errClose != nil {
		if err == nil {
			err = fmt.Errorf(`%s %q: %w`, logp, node.SysPath, errClose)
		} else {
			log.Printf(`%s %q: %s`, logp, node.SysPath, errClose)
		}
	}

	return err
}

// refresh the tree by rescanning from the root.
func (mfs *MemFS) refresh(url string) (node *Node, err error) {
	var syspath = filepath.Join(mfs.Root.SysPath, url)

	if mfs.Root.SysPath == `.` {
		// When "." joined with url "/file", the syspath become
		// "file" instead of "./file", this cause
		// [strings.HasPrefix] return false.
		syspath = `./` + syspath
	}

	if !strings.HasPrefix(syspath, mfs.Root.SysPath) {
		// Make sure the requested url always under the served
		// directory, to prevent accessing private files.
		return nil, fs.ErrNotExist
	}

	_, err = os.Stat(syspath)
	if err != nil {
		return nil, err
	}

	// The syspath is already cleaning-up, use it to get the relative
	// url path back.
	var dir = strings.TrimPrefix(syspath, mfs.Root.SysPath)

	// syspath exist in the file system but not in the mfs, try to scan
	// all trees start from the closes directory exist in the dir.
	// This is to handle detecting new file inside new directory.
	for node == nil {
		dir = filepath.Dir(dir)
		if dir == `.` {
			// On "./file" it will return ".".
			node = mfs.Root
			break
		}
		node = mfs.PathNodes.Get(dir)
	}

	err = mfs.scanDir(node)
	if err != nil {
		return nil, err
	}

	node = mfs.PathNodes.Get(url)
	if node == nil {
		return nil, fs.ErrNotExist
	}

	return node, nil
}

// resetAllModTime set the modTime on Root and its childs to the t.
// This method is only intended for testing.
func (mfs *MemFS) resetAllModTime(t time.Time) {
	mfs.Root.resetAllModTime(t)
}

// Watch create and start the [watchfs/v2.DirWatcher] that re-scan the content
// of Root directory recursively on every [memfs.WatchOptions.Interval],
// triggered by changes on [memfs.WatchOptions.File].
//
// The watcher will remove or update the tree and node content automatically
// if the included files is being deleted, created, or updated.
//
// The returned channel changes return list of Node that has been deleted,
// created, or updated.
// To stop watching for update call the [MemFS.StopWatch].
func (mfs *MemFS) Watch(watchopts WatchOptions) (
	changes <-chan []*Node, err error,
) {
	var logp = `Watch`

	watchopts.init(mfs.Opts.Root)
	mfs.watchopts = watchopts

	if mfs.dw != nil {
		mfs.StopWatch()
	}

	err = mfs.scanDir(mfs.Root)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	mfs.dwOptions = watchfs.DirWatcherOptions{
		FileWatcherOptions: watchopts.FileWatcherOptions,
		Root:               mfs.Opts.Root,
		Includes:           mfs.Opts.Includes,
		Excludes:           mfs.Opts.Excludes,
	}

	mfs.dw, err = watchfs.WatchDir(mfs.dwOptions)
	if err != nil {
		// There should be no error here, since we already check and
		// filled the required fields for DirWatcher.
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var c = make(chan []*Node, 1)
	go mfs.watch(c)

	return c, nil
}

func (mfs *MemFS) watch(c chan []*Node) {
	var (
		changes []os.FileInfo
		fi      os.FileInfo
		node    *Node
		err     error
	)
	for changes = range mfs.dw.C {
		var listNode []*Node
		for _, fi = range changes {
			var sysPath = fi.Name()
			var intPath = strings.TrimPrefix(sysPath, mfs.Opts.Root)
			intPath = filepath.Join(`/`, intPath)
			node, err = mfs.Get(intPath)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					// Nothing we can do.
					continue
				}
				node, err = mfs.addWatchFile(intPath, fi)
				if err != nil {
					log.Println(err)
					continue
				}
				listNode = append(listNode, node)
				continue
			}
			if fi.Size() == watchfs.FileFlagDeleted {
				if mfs.watchopts.Verbose {
					fmt.Printf("MemFS: file deleted: %q\n", intPath)
				}
				mfs.RemoveChild(node.Parent, node)
				listNode = append(listNode, node)
				continue
			}
			if mfs.watchopts.Verbose {
				fmt.Printf("MemFS: file updated: %q\n", intPath)
			}
			mfs.Update(node, fi)
			listNode = append(listNode, node)
		}
		c <- listNode
	}
	close(c)
}

func (mfs *MemFS) addWatchFile(intPath string, watchfi os.FileInfo) (node *Node, err error) {
	var dirPath = filepath.Dir(intPath)
	var parent *Node
	parent, err = mfs.Get(dirPath)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf(`addWatchFile: cannot find parent of %q`, intPath)
	}
	var fi = &Node{
		modTime: watchfi.ModTime(),
		Path:    intPath,
		name:    filepath.Base(intPath),
		size:    watchfi.Size(),
		mode:    watchfi.Mode(),
	}
	node, err = mfs.AddChild(parent, fi)
	if err != nil {
		return nil, err
	}
	if mfs.watchopts.Verbose {
		fmt.Printf("MemFS: file created: %q\n", intPath)
	}
	return node, nil
}
