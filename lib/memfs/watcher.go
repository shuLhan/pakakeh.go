// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	defWatchDelay    = 5 * time.Second
	watcherQueueSize = 16
)

// Watcher is a naive implementation of file event change notification.
type Watcher struct {
	C        <-chan NodeState // The channel on which the changes are delivered.
	qchanges chan NodeState

	node   *Node
	ticker *time.Ticker

	// Delay define a duration when the new changes will be fetched from
	// system.
	// This field is optional, minimum is 100 millisecond and default is
	// 5 seconds.
	delay time.Duration
}

// NewWatcher return a new file watcher that will inspect the file for changes
// for `path` with period specified by duration `d` argument.
//
// If duration is less or equal to 100 millisecond, it will be set to default
// duration (5 seconds).
//
// The changes can be consumed from the channel C.
// If the consumer is slower, channel is full, the changes will be dropped.
func NewWatcher(path string, d time.Duration) (w *Watcher, err error) {
	var (
		logp = "NewWatcher"

		dummyParent *Node
		fi          fs.FileInfo
	)

	if len(path) == 0 {
		return nil, fmt.Errorf("%s: path is empty", logp)
	}

	fi, err = os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("%s: path is directory", logp)
	}

	dummyParent = &Node{
		SysPath: filepath.Dir(path),
	}
	dummyParent.Path = dummyParent.SysPath

	return newWatcher(dummyParent, fi, d, nil)
}

// newWatcher create and initialize new Watcher like NewWatcher but using
// parent node.
func newWatcher(parent *Node, fi os.FileInfo, d time.Duration, qchanges chan NodeState) (
	w *Watcher, err error,
) {
	var (
		logp = "newWatcher"

		node *Node
	)

	// Create new node based on FileInfo without caching the content.
	node, err = NewNode(parent, fi, -1)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if d < 100*time.Millisecond {
		d = defWatchDelay
	}

	w = &Watcher{
		qchanges: qchanges,
		delay:    d,
		ticker:   time.NewTicker(d),
		node:     node,
	}
	if w.qchanges == nil {
		w.qchanges = make(chan NodeState, watcherQueueSize)
		w.C = w.qchanges
	}

	go w.start()

	return w, nil
}

// start fetching new file information every tick.
// This method run as goroutine and will finish when the file is deleted.
func (w *Watcher) start() {
	var (
		logp = "Watcher"

		newInfo fs.FileInfo
		ns      NodeState
		err     error
	)
	for range w.ticker.C {
		newInfo, err = os.Stat(w.node.SysPath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("%s: %s: %s", logp, w.node.SysPath, err)
				continue
			}

			ns.Node = *w.node
			ns.State = FileStateDeleted

			select {
			case w.qchanges <- ns:
			default:
			}
			w.node = nil
			return
		}

		if w.node.Mode() != newInfo.Mode() {
			w.node.SetMode(newInfo.Mode())

			ns.Node = *w.node
			ns.State = FileStateUpdateMode

			select {
			case w.qchanges <- ns:
			default:
			}
			continue
		}
		if w.node.ModTime().Equal(newInfo.ModTime()) {
			continue
		}

		w.node.SetModTime(newInfo.ModTime())
		w.node.SetSize(newInfo.Size())

		ns.Node = *w.node
		ns.State = FileStateUpdateContent

		select {
		case w.qchanges <- ns:
		default:
		}
	}
}

// Stop watching the file.
func (w *Watcher) Stop() {
	w.ticker.Stop()
}
