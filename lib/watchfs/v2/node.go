// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"os"
	"time"
)

// FileFlagDeleted the flag that indicated that a file has been
// deleted.
// The flag is stored inside the [os.FileInfo.Size].
const FileFlagDeleted = -1

const nodeFlagExcluded = -2

var nodeExcluded = node{
	size: nodeFlagExcluded,
}

type node struct {
	// The name contains the relative path, not only base name.
	name string

	// The file modification time in milliseconds.
	// This field also store the flag for excluded and deleted file.
	mtimems int64 `noequal:""`

	// Size of file.
	// Comparing size on directory does not works on all file system.
	size int64 `noequal:""`

	mode os.FileMode
}

func newNode(apath string) (n *node, err error) {
	var fi os.FileInfo
	fi, err = os.Stat(apath)
	if err != nil {
		return nil, err
	}
	n = &node{
		name:    apath,
		mtimems: fi.ModTime().UnixMilli(),
		size:    fi.Size(),
		mode:    fi.Mode(),
	}
	return n, nil
}

func (node *node) IsDir() bool {
	return node.mode.IsDir()
}

func (node *node) Mode() os.FileMode {
	return node.mode
}

func (node *node) ModTime() time.Time {
	return time.UnixMilli(node.mtimems)
}

// Name return the relative path to the file, not base name of file.
func (node *node) Name() string {
	return node.name
}

func (node *node) Size() int64 {
	return node.size
}

func (node *node) Sys() any {
	return node
}

func (node *node) equal(other *node) bool {
	if node.mtimems != other.mtimems {
		return false
	}
	if node.size != other.size {
		return false
	}
	if node.mode != other.mode {
		return false
	}
	return true
}
