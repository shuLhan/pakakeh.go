// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

// FileState define the state of file.
// There are three state of file: created, modified, or deleted.
type FileState byte

const (
	// FileStateCreated indicate that the file has been created.
	FileStateCreated FileState = iota
	// FileStateModified indicate that the file has been modified.
	FileStateModified
	// FileStateDeleted indicate that the file has been deleted.
	FileStateDeleted
)
