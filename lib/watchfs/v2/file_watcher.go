// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"errors"
	"os"
	"time"
)

// FileWatcher watch a single file.
// It will send the [os.FileInfo] to the channel C when the file is created,
// updated; or nil if file has been deleted.
//
// The FileWatcher may stop unexpectedly when the [os.Stat] return an
// error other than [os.ErrNotExist].
// The last error can be inspected using [FileWatcher.Err].
type FileWatcher struct {
	fstat os.FileInfo
	err   error

	// C receive new file information.
	C <-chan os.FileInfo
	c chan os.FileInfo

	ticker *time.Ticker
	opts   FileWatcherOptions
}

// WatchFile watch the file [watchfs.FileWatcherOptions.File] for being
// created, updated, or deleted; on every
// [watchfs.FileWatcherOptions.Interval].
func WatchFile(opts FileWatcherOptions) (fwatch *FileWatcher) {
	fwatch = &FileWatcher{
		c:      make(chan os.FileInfo, 1),
		ticker: time.NewTicker(opts.Interval),
		opts:   opts,
	}
	fwatch.fstat, fwatch.err = os.Stat(opts.File)
	fwatch.C = fwatch.c
	go fwatch.watch()
	return fwatch
}

// Err return the last error that cause the watch stopped.
func (fwatch *FileWatcher) Err() error {
	return fwatch.err
}

// Stop watching the file.
func (fwatch *FileWatcher) Stop() {
	fwatch.ticker.Stop()
	close(fwatch.c)
}

func (fwatch *FileWatcher) watch() {
	var newStat os.FileInfo
	for range fwatch.ticker.C {
		newStat, fwatch.err = os.Stat(fwatch.opts.File)
		if fwatch.err != nil {
			if errors.Is(fwatch.err, os.ErrNotExist) {
				if fwatch.fstat != nil {
					// File deleted.
					fwatch.fstat = nil
					fwatch.c <- nil
				}
				continue
			}
			// Other errors cause the watcher stop unexpectedly.
			fwatch.Stop()
			return
		}
		if fwatch.fstat != nil {
			if newStat.Mode() == fwatch.fstat.Mode() &&
				newStat.ModTime().Equal(fwatch.fstat.ModTime()) &&
				newStat.Size() == fwatch.fstat.Size() {
				continue
			}
		}
		// File created or updated.
		fwatch.fstat = newStat
		fwatch.c <- newStat
	}
}
