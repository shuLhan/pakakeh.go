// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

// FileState define the state of file.
// There are four states of file: created, updated on mode, updated on content
// or deleted.
type FileState byte

const (
	FileStateCreated       FileState = iota // New file is created.
	FileStateUpdateContent                  // The content of file is modified.
	FileStateUpdateMode                     // The mode of file is modified.
	FileStateDeleted                        // The file has been deleted.
)

// String return the string representation of FileState.
func (fs FileState) String() (s string) {
	switch fs {
	case FileStateCreated:
		s = "FileStateCreated"
	case FileStateUpdateContent:
		s = "FileStateUpdateContent"
	case FileStateUpdateMode:
		s = "FileStateUpdateMode"
	case FileStateDeleted:
		s = "FileStateDeleted"
	}
	return s
}
