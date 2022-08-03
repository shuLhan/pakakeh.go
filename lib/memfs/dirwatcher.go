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
	"time"

	"github.com/shuLhan/share/lib/debug"
)

const (
	dirWatcherQueueSize = 64
)

// DirWatcher is a naive implementation of directory change notification.
type DirWatcher struct {
	C            <-chan NodeState // The channel on which the changes are delivered to user.
	qchanges     chan NodeState
	qFileChanges chan NodeState
	qrun         chan bool

	root   *Node
	fs     *MemFS
	ticker *time.Ticker

	// dirs contains list of directory and their sub-directories that is
	// being watched for changes.
	// The map key is relative path to directory and its value is a node
	// information.
	dirs map[string]*Node

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
	dw.qrun = make(chan bool, 1)

	dw.dirs = make(map[string]*Node)
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
	dw.qrun <- false
	dw.ticker.Stop()
}

// dirsKeys return all the key in field dirs sorted in ascending order.
func (dw *DirWatcher) dirsKeys() (keys []string) {
	var (
		key string
	)
	for key = range dw.dirs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// mapSubdirs iterate each child node and check if its a directory or regular
// file.
// If its a directory add it to map of node and recursively iterate
// the childs.
// If its a regular file, start a NewWatcher.
func (dw *DirWatcher) mapSubdirs(node *Node) {
	var (
		logp = "DirWatcher.mapSubdirs"
		err  error
	)

	for _, child := range node.Childs {
		if child.IsDir() {
			dw.dirs[child.Path] = child
			dw.mapSubdirs(child)
			continue
		}
		_, err = newWatcher(node, child, dw.Delay, dw.qFileChanges)
		if err != nil {
			log.Printf("%s: %q: %s", logp, child.SysPath, err)
		}
	}
}

// unmapSubdirs find sub directories in node's childrens, recursively and
// remove it from map of node.
func (dw *DirWatcher) unmapSubdirs(node *Node) {
	for _, child := range node.Childs {
		if child.IsDir() {
			delete(dw.dirs, child.Path)
			dw.unmapSubdirs(child)
		}
		dw.fs.RemoveChild(node, child)
	}
	if node.IsDir() {
		delete(dw.dirs, node.Path)
	}
	dw.fs.RemoveChild(node.Parent, node)
}

// onContentChange handle changes on the content of directory.
//
// It will re-read the list of files in node directory and compare them with
// old content to detect deletion and addition of files.
func (dw *DirWatcher) onContentChange(node *Node) {
	var (
		logp = "onContentChange"
	)

	if debug.Value >= 2 {
		fmt.Printf("%s: %+v\n", logp, node)
	}

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

	// Find deleted files in directory.
	for _, child := range node.Childs {
		found := false
		for _, newInfo := range fis {
			if child.Name() == newInfo.Name() {
				found = true
				break
			}
		}
		if found {
			continue
		}
		if debug.Value >= 2 {
			fmt.Printf("%s: %q deleted\n", logp, child.Path)
		}
		dw.unmapSubdirs(child)
	}

	// Find new files in directory.
	for _, newInfo := range fis {
		found := false
		for _, child := range node.Childs {
			if newInfo.Name() == child.Name() {
				found = true
				break
			}
		}
		if found {
			continue
		}

		newChild, err := dw.fs.AddChild(node, newInfo)
		if err != nil {
			log.Printf("%s: %s", logp, err)
			continue
		}
		if newChild == nil {
			// a node is excluded.
			continue
		}

		if debug.Value >= 2 {
			fmt.Printf("%s: new child %s\n", logp, newChild.Path)
		}

		ns := NodeState{
			Node:  *newChild,
			State: FileStateCreated,
		}

		//nolint
		select {
		case dw.qchanges <- ns:
		}

		if newChild.IsDir() {
			dw.dirs[newChild.Path] = newChild
			dw.mapSubdirs(newChild)
			dw.onContentChange(newChild)
			continue
		}

		// Start watching the file for modification.
		_, err = newWatcher(node, newInfo, dw.Delay, dw.qFileChanges)
		if err != nil {
			log.Printf("%s: %s", logp, err)
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

	dw.dirs = make(map[string]*Node)
	dw.mapSubdirs(dw.root)

	if debug.Value >= 2 {
		fmt.Printf("%s: %s\n", logp, dw.root.Path)
	}

	ns := NodeState{
		Node:  *dw.root,
		State: FileStateCreated,
	}

	//nolint
	select {
	case dw.qchanges <- ns:
	}
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
	dw.dirs = nil

	if debug.Value >= 2 {
		fmt.Println("DirWatcher.onRootDeleted: root directory deleted")
	}

	//nolint
	select {
	case dw.qchanges <- ns:
	}
}

// onModified handle change when permission or attribute on node directory
// changed.
func (dw *DirWatcher) onModified(node *Node, newDirInfo os.FileInfo) {
	dw.fs.Update(node, newDirInfo)

	var (
		ns = NodeState{
			Node:  *node,
			State: FileStateUpdateMode,
		}
	)

	//nolint
	select {
	case dw.qchanges <- ns:
	}

	if debug.Value >= 2 {
		fmt.Printf("DirWatcher.onModified: %s\n", node.Path)
	}
}

func (dw *DirWatcher) start() {
	var (
		logp = "DirWatcher"
		ever = true

		node *Node
		fi   os.FileInfo
		ns   NodeState
		err  error
	)

	dw.ticker = time.NewTicker(dw.Delay)

	for ever {
		select {
		case <-dw.ticker.C:
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
				dw.onContentChange(dw.root)
				continue
			}
			if dw.root.Mode() != fi.Mode() {
				dw.onModified(dw.root, fi)
				continue
			}
			if dw.root.ModTime().Equal(fi.ModTime()) {
				dw.processSubdirs()
				continue
			}

			dw.fs.Update(dw.root, fi)
			dw.onContentChange(dw.root)
			dw.processSubdirs()

		case ns = <-dw.qFileChanges:
			node, err = dw.fs.Get(ns.Node.Path)
			if err != nil {
				log.Printf("%s: on file changes %s: %s", logp, ns.Node.Path, err)
			} else {
				ns.Node = *node
				switch ns.State {
				case FileStateDeleted:
					dw.fs.RemoveChild(node.Parent, node)
				default:
					dw.fs.Update(node, nil)
				}
			}
			dw.qchanges <- ns

		case <-dw.qrun:
			ever = false
		}
	}
}

func (dw *DirWatcher) processSubdirs() {
	logp := "processSubdirs"

	for _, node := range dw.dirs {
		if debug.Value >= 3 {
			fmt.Printf("%s: %q\n", logp, node.SysPath)
		}

		newDirInfo, err := os.Stat(node.SysPath)
		if err != nil {
			if os.IsNotExist(err) {
				dw.unmapSubdirs(node)
			} else {
				log.Printf("%s: %q: %s", logp, node.SysPath, err)
			}
			continue
		}
		if node.Mode() != newDirInfo.Mode() {
			dw.onModified(node, newDirInfo)
			continue
		}
		if node.ModTime().Equal(newDirInfo.ModTime()) {
			continue
		}

		dw.fs.Update(node, newDirInfo)
		dw.onContentChange(node)
	}
}
