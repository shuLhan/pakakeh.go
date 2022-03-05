// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shuLhan/share/lib/debug"
)

//
// Watcher is a naive implementation of file event change notification.
//
type Watcher struct {
	node   *Node
	ticker *time.Ticker

	// cb define a function that will be called when file modified or
	// deleted.
	cb WatchCallback

	// Delay define a duration when the new changes will be fetched from
	// system.
	// This field is optional, minimum is 100 millisecond and default is
	// 5 seconds.
	delay time.Duration
}

//
// NewWatcher return a new file watcher that will inspect the file for changes
// with period specified by duration `d` argument.
//
// If duration is less or equal to 100 millisecond, it will be set to default
// duration (5 seconds).
//
func NewWatcher(path string, d time.Duration, cb WatchCallback) (w *Watcher, err error) {
	logp := "NewWatcher"

	if len(path) == 0 {
		return nil, fmt.Errorf("%s: path is empty", logp)
	}
	if cb == nil {
		return nil, fmt.Errorf("%s: callback is not defined", logp)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("%s: path is directory", logp)
	}

	dummyParent := &Node{
		SysPath: filepath.Dir(path),
	}
	dummyParent.Path = dummyParent.SysPath

	return newWatcher(dummyParent, fi, d, cb)
}

// newWatcher create and initialize new Watcher like NewWatcher but using
// parent node.
func newWatcher(parent *Node, fi os.FileInfo, d time.Duration, cb WatchCallback) (
	w *Watcher, err error,
) {
	logp := "newWatcher"

	// Create new node based on FileInfo without caching the content.
	node, err := NewNode(parent, fi, -1)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if d < 100*time.Millisecond {
		d = time.Second * 5
	}

	w = &Watcher{
		delay:  d,
		cb:     cb,
		ticker: time.NewTicker(d),
		node:   node,
	}

	go w.start()

	return w, nil
}

// start fetching new file information every tick.
// This method run as goroutine and will finish when the file is deleted.
func (w *Watcher) start() {
	logp := "Watcher"
	if debug.Value >= 2 {
		fmt.Printf("%s: %s: watching for changes\n", logp, w.node.SysPath)
	}
	for range w.ticker.C {
		ns := &NodeState{
			Node: w.node,
		}

		newInfo, err := os.Stat(w.node.SysPath)
		if err != nil {
			if debug.Value >= 2 {
				fmt.Printf("%s: %s: deleted\n", logp, w.node.SysPath)
			}
			if !os.IsNotExist(err) {
				log.Printf("%s: %s: %s", logp, w.node.SysPath, err)
				continue
			}
			ns.State = FileStateDeleted
			w.cb(ns)
			w.node = nil
			return
		}

		if w.node.Mode() != newInfo.Mode() {
			if debug.Value >= 2 {
				fmt.Printf("%s: %s: mode updated\n", logp, w.node.SysPath)
			}
			ns.State = FileStateUpdateMode
			w.node.SetMode(newInfo.Mode())
			w.cb(ns)
			continue
		}
		if w.node.ModTime().Equal(newInfo.ModTime()) {
			continue
		}
		if debug.Value >= 2 {
			fmt.Printf("%s: %s: content updated\n", logp, w.node.SysPath)
		}

		w.node.SetModTime(newInfo.ModTime())
		w.node.SetSize(newInfo.Size())

		ns.State = FileStateUpdateContent
		w.cb(ns)
	}
}

//
// Stop watching the file.
//
func (w *Watcher) Stop() {
	w.ticker.Stop()
}
