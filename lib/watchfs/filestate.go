// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

// FileState define the state of file.
// There are four states of file: created, updated on mode, updated on
// content or deleted.
type FileState byte

const (
	// FileStateCreated when new file is created.
	FileStateCreated FileState = iota
	// FileStateUpdateContent when the content of file is modified.
	FileStateUpdateContent
	// FileStateUpdateMode when the mode of file is modified.
	FileStateUpdateMode
	// FileStateDeleted when the file has been deleted.
	FileStateDeleted
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
