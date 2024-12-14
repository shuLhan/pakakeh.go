// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"errors"
	"os"
	"time"
)

// FileWatcher watch a single file.
// It will send the [os.FileInfo] to the channel C when the file is
// created, updated; or nil if file has been deleted.
//
// The FileWatcher may stop unexpectedly anytime when the file stat cannot
// be retrieved with error other than [os.ErrNotExist].
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
	fwatch.C = fwatch.c
	go fwatch.watch()
	return fwatch
}

// Err return the last error that cause the watch failed to process or
// stopped.
func (fwatch *FileWatcher) Err() error {
	return fwatch.err
}

// Stop watching the file.
func (fwatch *FileWatcher) Stop() {
	fwatch.ticker.Stop()
	close(fwatch.c)
}

func (fwatch *FileWatcher) watch() {
	for range fwatch.ticker.C {
		if fwatch.fstat == nil {
			fwatch.fstat, fwatch.err = os.Stat(fwatch.opts.File)
			if fwatch.err != nil {
				if errors.Is(fwatch.err, os.ErrNotExist) {
					continue
				}
				fwatch.Stop()
				return
			}
			// File created.
			fwatch.err = nil
			fwatch.c <- fwatch.fstat
			continue
		}

		var newStat os.FileInfo
		newStat, fwatch.err = os.Stat(fwatch.opts.File)
		if fwatch.err != nil {
			if errors.Is(fwatch.err, os.ErrNotExist) {
				// File deleted.
				fwatch.fstat = nil
				fwatch.c <- nil
				continue
			}
			fwatch.Stop()
			return
		}
		if newStat.Mode() == fwatch.fstat.Mode() &&
			newStat.ModTime().Equal(fwatch.fstat.ModTime()) &&
			newStat.Size() == fwatch.fstat.Size() {
			continue
		}
		// File updated.
		fwatch.fstat = newStat
		fwatch.c <- newStat
	}
}
