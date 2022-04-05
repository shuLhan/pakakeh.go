// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import "time"

// WatchOptions define an options for the MemFS Watch method.
type WatchOptions struct {
	// Watches contain list of regular expressions for files to be watched
	// inside the Root, as addition to Includes pattern.
	// If this field is empty, only files pass the Includes filter will be
	// watched.
	Watches []string

	// Delay define the duration when the new changes will be checked from
	// system.
	// This field set the DirWatcher.Delay returned from Watch().
	// This field is optional, default is 5 seconds.
	Delay time.Duration
}
