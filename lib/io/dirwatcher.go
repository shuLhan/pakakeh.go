// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/memfs"
)

//
// DirWatcher is a naive implementation of directory change notification.
//
type DirWatcher struct {
	// This struct embed memfs.Options to map the directory to be watched
	// into memory.
	//
	// The Root field define the directory that we want to watch.
	//
	// Includes contains list of regex to filter file names that we want
	// to be notified.
	//
	// Excludes contains list of regex to filter file names that we did
	// not want to be notified.
	memfs.Options

	// Delay define a duration when the new changes will be fetched from
	// system.
	// This field is optional, minimum is 100 milli second and default is
	// 5 seconds.
	Delay time.Duration

	// Callback define a function that will be called when change detected
	// on directory.
	Callback WatchCallback

	// dirs contains list of directory and their sub-directories that is
	// being watched for changes.
	// The map key is relative path to directory and its value is a node
	// information.
	dirs map[string]*memfs.Node

	root   *memfs.Node
	fs     *memfs.MemFS
	ticker *time.Ticker
}

//
// Start watching changes in directory and its content.
//
func (dw *DirWatcher) Start() (err error) {
	logp := "DirWatcher.Start"

	if dw.Delay < 100*time.Millisecond {
		dw.Delay = time.Second * 5
	}
	if dw.Callback == nil {
		return fmt.Errorf("%s: callback is not defined", logp)
	}

	fi, err := os.Stat(dw.Root)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s: %q is not a directory", logp, dw.Root)
	}

	dw.Options.MaxFileSize = -1

	dw.fs, err = memfs.New(&dw.Options)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	dw.root = dw.fs.Root

	dw.dirs = make(map[string]*memfs.Node)
	dw.mapSubdirs(dw.root)
	go dw.start()

	return nil
}

// Stop watching changes on directory.
func (dw *DirWatcher) Stop() {
	dw.ticker.Stop()
}

//
// mapSubdirs iterate each child node and check if its a directory or regular
// file.
// If its a directory add it to map of node and recursively iterate
// the childs.
// If its a regular file, start a NewWatcher.
//
func (dw *DirWatcher) mapSubdirs(node *memfs.Node) {
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
		_, err = newWatcher(node, node.FileInfo, dw.Delay, dw.Callback)
		if err != nil {
			log.Printf("%s: %q: %s", logp, child.SysPath, err)
		}
	}
}

//
// unmapSubdirs find any sub directories in node's childrens and remove it
// from map of node.
//
func (dw *DirWatcher) unmapSubdirs(node *memfs.Node) {
	for _, child := range node.Childs {
		if !child.IsDir() {
			continue
		}

		delete(dw.dirs, child.Path)
		dw.unmapSubdirs(child)
	}
}

//
// onContentChange handle changes on the content of directory.
//
// It will re-read the list of files in node directory and compare them with
// old content to detect deletion and addition of files.
//
func (dw *DirWatcher) onContentChange(node *memfs.Node) {
	logp := "DirWatcher.onContentChange"

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
		if child.IsDir() {
			dw.unmapSubdirs(child)
		}
		dw.fs.RemoveChild(node, child)
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

		ns := &NodeState{
			Node:  newChild,
			State: FileStateCreated,
		}

		dw.Callback(ns)

		if newChild.IsDir() {
			dw.dirs[newChild.Path] = newChild
			dw.mapSubdirs(newChild)
			dw.onContentChange(newChild)
			continue
		}

		// Start watching the file for modification.
		_, err = newWatcher(node, newInfo, dw.Delay, dw.Callback)
		if err != nil {
			log.Printf("%s: %s", logp, err)
		}
	}
}

//
// onRootCreated handle changes when the root directory that we watch get
// created again, after being deleted.
// It will send created event, and re-mount the root directory back to memory,
// recursively.
//
func (dw *DirWatcher) onRootCreated() {
	var (
		logp = "DirWatcher.onRootCreated"
		err  error
	)

	dw.fs, err = memfs.New(&dw.Options)
	if err != nil {
		log.Printf("%s: %s", logp, err)
		return
	}

	dw.root, err = dw.fs.Get("/")
	if err != nil {
		log.Printf("%s: %s", logp, err)
		return
	}

	dw.dirs = make(map[string]*memfs.Node)
	dw.mapSubdirs(dw.root)

	ns := &NodeState{
		Node:  dw.root,
		State: FileStateCreated,
	}

	if debug.Value >= 2 {
		fmt.Printf("%s: %s", logp, dw.root.Path)
	}

	dw.Callback(ns)
}

//
// onRootDeleted handle change when the root directory that we watch get
// deleted.  It will send deleted event and unmount the root directory from
// memory.
//
func (dw *DirWatcher) onRootDeleted() {
	ns := &NodeState{
		Node:  dw.root,
		State: FileStateDeleted,
	}

	dw.fs = nil
	dw.root = nil
	dw.dirs = nil

	if debug.Value >= 2 {
		fmt.Println("DirWatcher.onRootDeleted: root directory deleted")
	}

	dw.Callback(ns)
}

//
// onModified handle change when permission or attribute on node directory
// changed.
//
func (dw *DirWatcher) onModified(node *memfs.Node, newDirInfo os.FileInfo) {
	dw.fs.Update(node, newDirInfo)

	ns := &NodeState{
		Node:  node,
		State: FileStateUpdateMode,
	}

	dw.Callback(ns)

	if debug.Value >= 2 {
		fmt.Printf("DirWatcher.onModified: %s\n", node.Path)
	}
}

func (dw *DirWatcher) start() {
	logp := "DirWatcher"

	dw.ticker = time.NewTicker(dw.Delay)

	for range dw.ticker.C {
		newDirInfo, err := os.Stat(dw.Root)
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
			continue
		}
		if dw.root.Mode() != newDirInfo.Mode() {
			dw.onModified(dw.root, newDirInfo)
			continue
		}
		if dw.root.ModTime().Equal(newDirInfo.ModTime()) {
			dw.processSubdirs()
			continue
		}

		dw.fs.Update(dw.root, newDirInfo)
		dw.onContentChange(dw.root)
		dw.processSubdirs()
	}
}

func (dw *DirWatcher) processSubdirs() {
	logp := "DirWatcher.processSubdirs"

	for _, node := range dw.dirs {
		if debug.Value >= 3 {
			fmt.Printf("%s: %q\n", logp, node.SysPath)
		}

		newDirInfo, err := os.Stat(node.SysPath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("%s: %q: %s", logp, node.SysPath, err)
				continue
			}
			dw.unmapSubdirs(node)
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
