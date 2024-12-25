// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

// Package watchfs implement naive file and directory watcher.
//
// This package is deprecated, we keep it here for historical only.
// The new implementation should use "watchfs/v2".
package watchfs

import "git.sr.ht/~shulhan/pakakeh.go/lib/memfs"

// Node represent single file or directory.
type Node = memfs.Node
