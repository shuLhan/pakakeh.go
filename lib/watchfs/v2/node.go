// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"os"
	"time"
)

// FileFlagDeleted indicated that a file has been deleted.
// The flag is stored inside the [os.FileInfo.Size].
const FileFlagDeleted = -1

const nodeFlagExcluded = -2

const nodeFlagForced = -3

var nodeExcluded = node{
	size: nodeFlagExcluded,
}

type node struct {
	// The file modification time.
	// This field also store the flag for excluded and deleted file.
	mtime time.Time `noequal:""`

	// The name contains the relative path, not only base name.
	name string

	// Size of file.
	// For directory the size contains the number of childs, the length of
	// slice returned by [os.ReadDir].
	size int64

	mode os.FileMode
}

func newNode(apath string) (n *node, err error) {
	var fi os.FileInfo
	fi, err = os.Stat(apath)
	if err != nil {
		return nil, err
	}
	n = &node{
		name:  apath,
		mtime: fi.ModTime(),
		mode:  fi.Mode(),
	}
	if !fi.IsDir() {
		n.size = fi.Size()
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
	return node.mtime
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
	if !node.mtime.Equal(other.mtime) {
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
