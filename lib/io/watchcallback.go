// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

//
// WatchCallback is a function that will be called when DirWatcher detect any
// changes.  It will send the file information and its state.
//
type WatchCallback func(*NodeState)
