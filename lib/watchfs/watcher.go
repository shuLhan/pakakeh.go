// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
)

const watcherQueueSize = 16

// Watcher is a naive implementation of file event change notification.
type Watcher struct {
	// The channel on which the changes are delivered.
	C <-chan NodeState

	qchanges chan NodeState

	done chan struct{}

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
	node, err = memfs.NewNode(parent, fi, -1)
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
		done:     make(chan struct{}),
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
		ever = true

		newInfo fs.FileInfo
		ns      NodeState
		err     error
	)
	for ever {
		select {
		case <-w.ticker.C:
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
				ever = false
				w.ticker.Stop()
				continue
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
		case <-w.done:
			ever = false
			w.ticker.Stop()
			w.done <- struct{}{}
		}
	}
}

// Stop watching the file.
func (w *Watcher) Stop() {
	select {
	case w.done <- struct{}{}:
		<-w.done
	default:
		// Ticker has been stopped due to file being deleted.
	}
}
