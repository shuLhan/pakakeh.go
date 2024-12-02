// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"regexp"
	"time"
)

// WatchOptions define an options for the MemFS Watch method.
type WatchOptions struct {
	// Watches contain list of regular expressions for files to be watched
	// inside the Root, as addition to Includes pattern.
	// If this field is empty, only files pass the Includes filter will be
	// watched.
	Watches []string

	watchRE []*regexp.Regexp

	// Delay define the duration when the new changes will be checked from
	// system.
	// This field set the DirWatcher.Delay returned from Watch().
	// This field is optional, default is 5 seconds.
	Delay time.Duration
}

func (watchopts *WatchOptions) init() (err error) {
	if watchopts.Delay < 100*time.Millisecond {
		watchopts.Delay = defWatchDelay
	}

	var (
		v  string
		re *regexp.Regexp
	)
	watchopts.watchRE = nil
	for _, v = range watchopts.Watches {
		re, err = regexp.Compile(v)
		if err != nil {
			return err
		}
		watchopts.watchRE = append(watchopts.watchRE, re)
	}
	return nil
}

// isWatched return true if the sysPath is filtered to be watched.
func (watchopts *WatchOptions) isWatched(sysPath string) bool {
	var re *regexp.Regexp
	for _, re = range watchopts.watchRE {
		if re.MatchString(sysPath) {
			return true
		}
	}
	return false
}
