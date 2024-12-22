// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

// WatchCallback is a function that will be called when Watcher or DirWatcher
// detect any changes on its file or directory.
// The watcher will pass the file information and its state.
type WatchCallback func(*NodeState)
