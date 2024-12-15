// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"regexp"
	"strings"
)

// DirWatcherOptions contains options to watch directory.
type DirWatcherOptions struct {
	FileWatcherOptions

	// The root directory where files to be scanned.
	Root string

	// List of regex for files or directories to be excluded from
	// scanning.
	// The Excludes option will be processed before Includes.
	Excludes []string

	// List of regex for files or directories to be included from
	// scanning.
	Includes []string

	reExcludes []*regexp.Regexp
	reIncludes []*regexp.Regexp
}

func (opts *DirWatcherOptions) init() (err error) {
	var (
		str string
		re  *regexp.Regexp
	)
	for _, str = range opts.Excludes {
		str = strings.TrimSpace(str)
		if len(str) == 0 {
			// Accidentally using empty string here may cause
			// everying get excluded.
			continue
		}
		re, err = regexp.Compile(str)
		if err != nil {
			return err
		}
		opts.reExcludes = append(opts.reExcludes, re)
	}
	for _, str = range opts.Includes {
		str = strings.TrimSpace(str)
		if len(str) == 0 {
			continue
		}
		re, err = regexp.Compile(str)
		if err != nil {
			return err
		}
		opts.reIncludes = append(opts.reIncludes, re)
	}
	return nil
}

func (opts *DirWatcherOptions) isExcluded(pathFile string) bool {
	var re *regexp.Regexp
	for _, re = range opts.reExcludes {
		if re.MatchString(pathFile) {
			return true
		}
	}
	return false
}

// isIncluded will return true if the list Includes is empty or it is match
// with one of the Includes regex.
func (opts *DirWatcherOptions) isIncluded(pathFile string) bool {
	if len(opts.reIncludes) == 0 {
		return true
	}
	var re *regexp.Regexp
	for _, re = range opts.reIncludes {
		if re.MatchString(pathFile) {
			return true
		}
	}
	return false
}
