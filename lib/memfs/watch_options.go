// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package memfs

import (
	"path/filepath"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs/v2"
)

// DefaultWatchFile define default file name to be watch for changes.
// Any update to this file will trigger rescan on the memfs tree.
const DefaultWatchFile = `.memfs_rescan`

const defWatchInterval = 5 * time.Second

// WatchOptions define an options for the MemFS Watch method.
//
// If the [watchfs.FileWatcherOptions.File] is empty it will default to
// [DefaultWatchFile] inside the [memfs.Options.Root].
// The [watchfs.FileWatcherOptions.Interval] must be greater than 10
// milliseconds, otherwise it will default to 5 seconds.
type WatchOptions struct {
	watchfs.FileWatcherOptions

	// Verbose if true print the file changes information to stdout.
	Verbose bool
}

func (watchopts *WatchOptions) init(root string) {
	if len(watchopts.File) == 0 {
		watchopts.File = filepath.Join(root, DefaultWatchFile)
	}
	if watchopts.Interval < 10*time.Millisecond {
		watchopts.Interval = defWatchInterval
	}
}
