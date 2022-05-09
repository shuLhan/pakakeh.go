// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

// WatchCallback is a function that will be called when Watcher or DirWatcher
// detect any changes on its file or directory.
// The watcher will pass the file information and its state.
type WatchCallback func(*NodeState)
