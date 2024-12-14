// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import "time"

// FileWatcherOptions define the options to watch file.
type FileWatcherOptions struct {
	// Path to the file to be watched.
	File string

	// Interval to check for file changes.
	Interval time.Duration
}
