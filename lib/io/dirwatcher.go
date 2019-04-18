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
	// Path to directory that we want to watch.
	Path string

	// Includes contains list of regex to filter file names that we want
	// to be notified.
	Includes []string

	// Excludes contains list of regex to filter file names that we did
	// not want to be notified.
	Excludes []string

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
// Start create a new watcher that detect changes in directory and its
// content.
//
func (dw *DirWatcher) Start() (err error) {
	if dw.Delay < 100*time.Millisecond {
		dw.Delay = time.Second * 5
	}
	if dw.Callback == nil {
		return fmt.Errorf("lib/io: NewDirWatcher: callback is not defined")
	}

	fi, err := os.Stat(dw.Path)
	if err != nil {
		return fmt.Errorf("lib/io: NewDirWatcher: " + err.Error())
	}
	if !fi.IsDir() {
		return fmt.Errorf("lib/io: NewDirWatcher: %q is not a directory", dw.Path)
	}

	dw.fs, err = memfs.New(dw.Includes, dw.Excludes, false)
	if err != nil {
		return fmt.Errorf("lib/io: NewDirWatcher: " + err.Error())
	}

	err = dw.fs.Mount(dw.Path)
	if err != nil {
		return fmt.Errorf("lib/io: NewDirWatcher: " + err.Error())
	}

	dw.root, err = dw.fs.Get("/")
	if err != nil {
		return fmt.Errorf("lib/io: NewDirWatcher: " + err.Error())
	}

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
// mapSubdirs find any sub directories on node's childrens and add it
// to map of node.
//
func (dw *DirWatcher) mapSubdirs(node *memfs.Node) {
	for _, child := range node.Childs {
		if !child.Mode.IsDir() {
			continue
		}

		dw.dirs[child.Path] = child
		dw.mapSubdirs(child)
	}
}

//
// unmapSubdirs find any sub directories in node's childrens and remove it
// from map of node.
//
func (dw *DirWatcher) unmapSubdirs(node *memfs.Node) {
	for _, child := range node.Childs {
		if !child.Mode.IsDir() {
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
	if debug.Value >= 2 {
		fmt.Printf("lib/io: DirWatcher.onContentChange: %+v\n", node)
	}

	f, err := os.Open(node.SysPath)
	if err != nil {
		log.Println("lib/io: DirWatcher.onContentChange: " + err.Error())
		return
	}

	fis, err := f.Readdir(0)
	if err != nil {
		log.Println("lib/io: DirWatcher.onContentChange: " + err.Error())
		return
	}

	err = f.Close()
	if err != nil {
		log.Println("lib/io: DirWatcher.onContentChange: " + err.Error())
	}

	// Find deleted files in directory.
	for _, child := range node.Childs {
		found := false
		for _, newInfo := range fis {
			if child.Name == newInfo.Name() {
				found = true
				break
			}
		}
		if !found {
			if debug.Value >= 2 {
				fmt.Printf("lib/io: DirWatcher.onContentChange: deleted %+v\n", child)
			}

			// A node is deleted in node's childs.
			ns := &NodeState{
				Node:  child,
				State: FileStateDeleted,
			}
			dw.Callback(ns)

			if child.Mode.IsDir() {
				dw.unmapSubdirs(child)
			}

			dw.fs.RemoveChild(node, child)
			continue
		}
	}

	// Find new files in directory.
	for _, newInfo := range fis {
		found := false
		for _, child := range node.Childs {
			if newInfo.Name() == child.Name {
				found = true
				break
			}
		}
		if found {
			continue
		}

		newChild, err := dw.fs.AddChild(node, newInfo)
		if err != nil {
			log.Printf("lib/io: DirWatcher.onContentChange: " + err.Error())
			continue
		}
		if newChild == nil {
			log.Printf("lib/io: DirWatcher.onContentChange: exclude %q\n", newInfo.Name())
			continue
		}

		if debug.Value >= 2 {
			fmt.Printf("lib/io: DirWatcher.onContentChange: new child %+v\n", newChild)
		}

		ns := &NodeState{
			Node:  newChild,
			State: FileStateCreated,
		}

		dw.Callback(ns)

		if newChild.Mode.IsDir() {
			dw.dirs[newChild.Path] = newChild
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
	err := dw.fs.Mount(dw.Path)
	if err != nil {
		log.Fatal("lib/io: DirWatcher.onRootCreated: " + err.Error())
	}

	dw.root, err = dw.fs.Get("/")
	if err != nil {
		log.Fatal("lib/io: DirWatcher.onRootCreated: " + err.Error())
	}

	dw.dirs = make(map[string]*memfs.Node)
	dw.mapSubdirs(dw.root)

	ns := &NodeState{
		Node:  dw.root,
		State: FileStateCreated,
	}

	if debug.Value >= 2 {
		fmt.Printf("lib/io: DirWatcher.onRootCreated: %+v\n", dw.root)
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

	dw.fs.Unmount()
	dw.root = nil
	dw.dirs = nil

	if debug.Value >= 2 {
		fmt.Println("lib/io: DirWatcher: root directory deleted")
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
		State: FileStateModified,
	}

	dw.Callback(ns)

	if debug.Value >= 2 {
		fmt.Printf("lib/io: DirWatcher.onModified: %+v\n", node)
	}
}

func (dw *DirWatcher) start() {
	dw.ticker = time.NewTicker(dw.Delay)

	for range dw.ticker.C {
		newDirInfo, err := os.Stat(dw.Path)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Println("lib/io: DirWatcher: " + err.Error())
				continue
			}
			if dw.fs.IsMounted() {
				dw.onRootDeleted()
			}
			continue
		}
		if !dw.fs.IsMounted() {
			dw.onRootCreated()
			continue
		}
		if dw.root.Mode != newDirInfo.Mode() {
			dw.onModified(dw.root, newDirInfo)
			continue
		}
		if dw.root.ModTime.Equal(newDirInfo.ModTime()) {
			dw.processSubdirs()
			continue
		}

		dw.fs.Update(dw.root, newDirInfo)
		dw.onContentChange(dw.root)
		dw.processSubdirs()
	}
}

func (dw *DirWatcher) processSubdirs() {
	for _, node := range dw.dirs {
		if debug.Value >= 2 {
			fmt.Printf("lib/io: DirWatcher: processSubdirs: %q\n", node.SysPath)
		}

		newDirInfo, err := os.Stat(node.SysPath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Println("lib/io: DirWatcher: " + err.Error())
				continue
			}
			dw.unmapSubdirs(node)
			continue
		}
		if node.Mode != newDirInfo.Mode() {
			dw.onModified(node, newDirInfo)
			continue
		}
		if node.ModTime.Equal(newDirInfo.ModTime()) {
			continue
		}

		dw.fs.Update(node, newDirInfo)
		dw.onContentChange(node)
	}
}
