// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	dirWatcherQueueSize = 64
)

// DirWatcher is a naive implementation of directory change notification.
type DirWatcher struct {
	// C channel on which the changes are delivered to user.
	C <-chan NodeState

	// qchanges channel where all node modification published.
	qchanges chan NodeState

	// qFileChanges channel where file changes from NewWatcher
	// consumed.
	qFileChanges chan NodeState

	qrun chan struct{}

	root *Node
	fs   *MemFS

	// dirs contains list of directory and their sub-directories that is
	// being watched for changes.
	// The map key is relative path to directory and its value is a node
	// information.
	dirs map[string]*Node

	// fileWatcher contains active watcher for file with Node.Path as
	// key.
	fileWatcher map[string]*Watcher

	// This struct embed Options to map the directory to be watched
	// into memory.
	//
	// The Root field define the directory that we want to watch.
	//
	// Includes contains list of regex to filter file names that we want
	// to be notified.
	//
	// Excludes contains list of regex to filter file names that we did
	// not want to be notified.
	Options

	// Delay define a duration when the new changes will be fetched from
	// system.
	// This field is optional, minimum is 100 milli second and default is
	// 5 seconds.
	Delay time.Duration

	dirsLocker sync.Mutex
}

func (dw *DirWatcher) init() (err error) {
	var (
		logp = "init"

		fi fs.FileInfo
	)

	if dw.Delay < 100*time.Millisecond {
		dw.Delay = defWatchDelay
	}

	if dw.fs == nil {
		fi, err = os.Stat(dw.Root)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
		if !fi.IsDir() {
			return fmt.Errorf("%s: %q is not a directory", logp, dw.Root)
		}

		dw.Options.MaxFileSize = -1

		dw.fs, err = New(&dw.Options)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}
	dw.root = dw.fs.Root

	dw.qchanges = make(chan NodeState, dirWatcherQueueSize)
	dw.C = dw.qchanges

	dw.qFileChanges = make(chan NodeState, dirWatcherQueueSize)
	dw.qrun = make(chan struct{})

	dw.dirs = make(map[string]*Node)
	dw.fileWatcher = make(map[string]*Watcher)

	dw.mapSubdirs(dw.root)

	return nil
}

// Start watching changes in directory and its content.
func (dw *DirWatcher) Start() (err error) {
	var (
		logp = "Start"
	)

	err = dw.init()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	go dw.start()

	return nil
}

// Stop watching changes on directory.
func (dw *DirWatcher) Stop() {
	// Stop all file watchers.
	var watcher *Watcher
	for _, watcher = range dw.fileWatcher {
		watcher.Stop()
	}

	select {
	case dw.qrun <- struct{}{}:
		<-dw.qrun
	default:
	}
}

// dirsKeys return all the key in field dirs sorted in ascending order.
func (dw *DirWatcher) dirsKeys() (keys []string) {
	var (
		key string
	)
	dw.dirsLocker.Lock()
	for key = range dw.dirs {
		keys = append(keys, key)
	}
	dw.dirsLocker.Unlock()
	sort.Strings(keys)
	return keys
}

// mapSubdirs iterate each child node recursively and map any sub
// directories into mapSubdirs.
// If its a regular file, start a NewWatcher.
func (dw *DirWatcher) mapSubdirs(node *Node) {
	var (
		logp = `DirWatcher.mapSubdirs`

		child   *Node
		watcher *Watcher
		err     error
	)

	for _, child = range node.Childs {
		if child.IsDir() {
			dw.dirsLocker.Lock()
			dw.dirs[child.Path] = child
			dw.dirsLocker.Unlock()
			dw.mapSubdirs(child)
			continue
		}
		watcher, err = newWatcher(node, child, dw.Delay, dw.qFileChanges)
		if err != nil {
			log.Printf("%s %q: %s", logp, child.SysPath, err)
			continue
		}
		dw.fileWatcher[child.Path] = watcher
	}
}

// onCreated handle new child created on parent node.
func (dw *DirWatcher) onCreated(parent, child *Node) (err error) {
	if child.IsDir() {
		dw.dirsLocker.Lock()
		dw.dirs[child.Path] = child
		dw.dirsLocker.Unlock()
	} else {
		// Start watching the file for modification.
		var watcher *Watcher

		watcher, err = newWatcher(parent, child, dw.Delay, dw.qFileChanges)
		if err != nil {
			return fmt.Errorf(`onCreated: %w`, err)
		}

		dw.fileWatcher[child.Path] = watcher
	}

	var ns = NodeState{
		Node:  *child,
		State: FileStateCreated,
	}

	select {
	case dw.qchanges <- ns:
	default:
	}
	return nil
}

// onDirDeleted remove the node from being watched and from memfs, including
// its childs if its a directory.
func (dw *DirWatcher) onDirDeleted(node *Node) {
	var child *Node

	for _, child = range node.Childs {
		if child.IsDir() {
			dw.onDirDeleted(child)
		}
		dw.fs.RemoveChild(node, child)
	}

	dw.dirsLocker.Lock()
	delete(dw.dirs, node.Path)
	dw.dirsLocker.Unlock()

	dw.fs.RemoveChild(node.Parent, node)

	var ns = NodeState{
		State: FileStateDeleted,
		Node:  *node,
	}
	select {
	case dw.qchanges <- ns:
	default:
	}
}

func (dw *DirWatcher) onFileDeleted(node *Node) {
	var watcher = dw.fileWatcher[node.Path]
	if watcher != nil {
		watcher.Stop()
		delete(dw.fileWatcher, node.Path)
	}

	dw.fs.RemoveChild(node.Parent, node)

	var ns = NodeState{
		State: FileStateDeleted,
		Node:  *node,
	}
	select {
	case dw.qchanges <- ns:
	default:
	}
}

// onUpdateDir handle changes on the directory "node".
//
// It will re-read the list of files in node directory and compare them with
// old content to detect deletion or addition of new file.
func (dw *DirWatcher) onUpdateDir(node *Node) {
	var (
		logp = "onUpdateDir"
	)

	f, err := os.Open(node.SysPath)
	if err != nil {
		log.Printf("%s: %s", logp, err)
		return
	}

	fis, err := f.Readdir(0)
	if err != nil {
		log.Printf("%s: %s", logp, err)
		return
	}

	err = f.Close()
	if err != nil {
		log.Printf("%s: %s", logp, err)
	}

	var (
		mapChild = make(map[string]*Node, len(node.Childs))

		child   *Node
		newInfo os.FileInfo
	)

	// Store the current childs into a map first to easily get
	// existing nodes.
	for _, child = range node.Childs {
		mapChild[child.name] = child
	}
	node.Childs = nil

	// Find new files in directory.
	for _, newInfo = range fis {
		child = mapChild[newInfo.Name()]
		if child != nil {
			// The node already exist previously.
			node.Childs = append(node.Childs, child)
			delete(mapChild, newInfo.Name())
			continue
		}

		// Process the new child.

		child, err = dw.fs.AddChild(node, newInfo)
		if err != nil {
			log.Printf("%s: %s", logp, err)
			continue
		}
		if child == nil {
			// The child is being excluded.
			continue
		}

		err = dw.onCreated(node, child)
		if err != nil {
			log.Printf(`%s: %s`, logp, err)
		}
		if child.IsDir() {
			dw.onUpdateDir(child)
		}
	}

	// The rest of the mapChild now contains the deleted nodes.
	for _, child = range mapChild {
		if child.IsDir() {
			// Only process directory, files is processed by
			// qFileChanges.
			dw.onDirDeleted(child)
		}
	}
}

// onRootCreated handle changes when the root directory that we watch get
// created again, after being deleted.
// It will send created event, and re-mount the root directory back to memory,
// recursively.
func (dw *DirWatcher) onRootCreated() {
	var (
		logp = "DirWatcher.onRootCreated"
		err  error
	)

	dw.fs, err = New(&dw.Options)
	if err != nil {
		log.Printf("%s: %s", logp, err)
		return
	}

	dw.root, err = dw.fs.Get("/")
	if err != nil {
		log.Printf("%s: %s", logp, err)
		return
	}

	dw.dirsLocker.Lock()
	dw.dirs = make(map[string]*Node)
	dw.dirsLocker.Unlock()

	_ = dw.onCreated(nil, dw.root)

	dw.mapSubdirs(dw.root)
}

// onRootDeleted handle change when the root directory that we watch get
// deleted.  It will send deleted event and unmount the root directory from
// memory.
func (dw *DirWatcher) onRootDeleted() {
	var (
		ns = NodeState{
			Node:  *dw.root,
			State: FileStateDeleted,
		}
	)

	dw.fs = nil
	dw.root = nil
	dw.dirsLocker.Lock()
	dw.dirs = nil
	dw.dirsLocker.Unlock()

	select {
	case dw.qchanges <- ns:
	default:
	}
}

// onUpdateContent handle when the file modification changes.
func (dw *DirWatcher) onUpdateContent(node *Node, newInfo os.FileInfo) {
	var (
		logp = `onUpdateContent`

		err error
	)

	if newInfo == nil {
		newInfo, err = os.Stat(node.SysPath)
		if err != nil {
			log.Printf(`%s %q: %s`, logp, node.Path, err)
			return
		}
	}

	node.modTime = newInfo.ModTime()
	node.size = newInfo.Size()

	if !node.IsDir() {
		err = node.updateContent(dw.fs.Opts.MaxFileSize)
		if err != nil {
			log.Printf(`%s %q: %s`, logp, node.Path, err)
		}
	}

	var ns = NodeState{
		Node:  *node,
		State: FileStateUpdateContent,
	}

	select {
	case dw.qchanges <- ns:
	default:
	}
}

// onUpdateMode handle change when permission or attribute of node changed.
func (dw *DirWatcher) onUpdateMode(node *Node, newInfo os.FileInfo) {
	if newInfo == nil {
		var err error

		newInfo, err = os.Stat(node.SysPath)
		if err != nil {
			log.Printf(`onUpdateMode %q: %s`, node.Path, err)
			return
		}
	}

	node.mode = newInfo.Mode()

	var ns = NodeState{
		Node:  *node,
		State: FileStateUpdateMode,
	}

	select {
	case dw.qchanges <- ns:
	default:
	}
}

func (dw *DirWatcher) start() {
	var (
		logp   = "DirWatcher"
		ticker = time.NewTicker(dw.Delay)
		ever   = true

		node *Node
		fi   os.FileInfo
		ns   NodeState
		err  error
	)

	for ever {
		select {
		case <-ticker.C:
			fi, err = os.Stat(dw.Root)
			if err != nil {
				if !os.IsNotExist(err) {
					log.Printf("%s: %s", logp, err)
					continue
				}
				if dw.fs != nil {
					dw.onRootDeleted()
				}
				continue
			}
			if dw.fs == nil {
				dw.onRootCreated()
				dw.onUpdateDir(dw.root)
				continue
			}
			if dw.root.Mode() != fi.Mode() {
				dw.onUpdateMode(dw.root, fi)
			}
			if !dw.root.ModTime().Equal(fi.ModTime()) {
				dw.onUpdateDir(dw.root)
			}
			dw.processSubdirs()

		case ns = <-dw.qFileChanges:
			node, err = dw.fs.Get(ns.Node.Path)
			if err != nil {
				log.Printf("%s: on file changes %s: %s", logp, ns.Node.Path, err)
				dw.onFileDeleted(&ns.Node)
			} else {
				ns.Node = *node
				switch ns.State {
				case FileStateDeleted:
					dw.onFileDeleted(node)
				case FileStateUpdateMode:
					dw.onUpdateMode(node, nil)
				case FileStateUpdateContent:
					dw.onUpdateContent(node, nil)
				}
			}

		case <-dw.qrun:
			ever = false
			ticker.Stop()
			// Signal back to the Stop caller.
			dw.qrun <- struct{}{}
		}
	}
}

func (dw *DirWatcher) processSubdirs() {
	var (
		logp = `processSubdirs`

		node       *Node
		newDirInfo os.FileInfo
		err        error
	)

	for _, node = range dw.dirs {
		newDirInfo, err = os.Stat(node.SysPath)
		if err != nil {
			if os.IsNotExist(err) {
				dw.onDirDeleted(node)
			} else {
				log.Printf("%s: %q: %s", logp, node.SysPath, err)
			}
			continue
		}
		if node.Mode() != newDirInfo.Mode() {
			dw.onUpdateMode(node, newDirInfo)
		}
		if !node.ModTime().Equal(newDirInfo.ModTime()) {
			dw.onUpdateDir(node)
		}
	}
}
