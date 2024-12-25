// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"time"
)

// DirWatcherOptions to create and initialize [DirWatcher].
//
// The includes and excludes pattern applied relative to the system
// path.
// The Excludes patterns will be applied first before the Includes.
// If the path is not excluded and Includes is empty, it will be
// assumed as included.
type DirWatcherOptions struct {
	// The Root field define the directory that we want to watch.
	Root string

	// Includes contains list of regex to filter file names that we want
	// to be notified.
	Includes []string

	// Excludes contains list of regex to filter file names that we did
	// not want to be notified.
	Excludes []string

	// Delay define a duration when the new changes will be fetched from
	// system.
	// This field is optional, minimum is 100 milli second and default
	// is 5 seconds.
	Delay time.Duration
}
