// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

// Package watchfs implement naive file and directory watcher.
//
// This package is deprecated, we keep it here for historical only.
// The new implementation should use "watchfs/v2".
package watchfs

import (
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
)

const defWatchDelay = 5 * time.Second

// Node represent single file or directory.
type Node = memfs.Node

// NodeState contains the information about the file and its state.
type NodeState struct {
	// Node represent the file information.
	Node Node
	// State of file, its either created, modified, or deleted.
	State FileState
}

// WatchCallback is a function that will be called when [Watcher] or
// [DirWatcher] detect any changes on its file or directory.
// The watcher will pass the file information and its state.
type WatchCallback func(*NodeState)
