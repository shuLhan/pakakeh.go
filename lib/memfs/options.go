// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package memfs

import (
	"os"
	"regexp"
	"strings"
)

const (
	defaultMaxFileSize = 1024 * 1024 * 5
)

// Options to create and initialize the MemFS.
type Options struct {
	// Embed options for GoEmbed method.
	Embed EmbedOptions

	// Root define the path to directory where its contents will be mapped
	// to memory or to be embedded as Go source code using GoEmbed.
	Root string

	// The includes and excludes pattern applied relative to the system
	// path.
	// The Excludes patterns will be applied first before the Includes.
	// If the path is not excluded and Includes is empty, it will be
	// assumed as included.
	Includes []string
	Excludes []string

	incRE []*regexp.Regexp
	excRE []*regexp.Regexp

	// MaxFileSize define maximum file size that can be stored on memory.
	// The default value is 5 MB.
	// If its value is negative, the content of file will not be mapped to
	// memory, the MemFS will behave as directory tree.
	MaxFileSize int64

	// TryDirect define a flag to bypass file in memory.
	// If its true, any call to Get will try direct read to file system.
	// This flag has several use cases.
	// First, to test serving file system directly from disk during
	// development.
	// Second, to combine embedded MemFS instance with non-embedded
	// instance.
	// One is reading content from memory, one is reading content from
	// disk directly.
	TryDirect bool
}

// init initialize the options with default value.
func (opts *Options) init() (err error) {
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = defaultMaxFileSize
	}

	opts.Root = strings.TrimSuffix(opts.Root, `/`)
	if len(opts.Root) == 0 {
		opts.Root = `.`
	}

	var (
		v  string
		re *regexp.Regexp
	)
	for _, v = range opts.Includes {
		re, err = regexp.Compile(v)
		if err != nil {
			return err
		}
		opts.incRE = append(opts.incRE, re)
	}
	for _, v = range opts.Excludes {
		re, err = regexp.Compile(v)
		if err != nil {
			return err
		}
		opts.excRE = append(opts.excRE, re)
	}
	return nil
}

// isExcluded return true if the sysPath is match with one of regex in
// Excludes.
func (opts *Options) isExcluded(sysPath string) bool {
	var re *regexp.Regexp
	for _, re = range opts.excRE {
		if re.MatchString(sysPath) {
			return true
		}
	}
	return false
}

// isIncluded return true if the sysPath is pass the list of Includes
// regexp, or no filter defined.
func (opts *Options) isIncluded(sysPath string, fi os.FileInfo) bool {
	if len(opts.incRE) == 0 {
		// No filter defined, default to always included.
		return true
	}
	var re *regexp.Regexp
	for _, re = range opts.incRE {
		if re.MatchString(sysPath) {
			return true
		}
	}
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		// File is symlink, get the real FileInfo to check if its
		// directory or not.
		var err error
		fi, err = os.Stat(sysPath)
		if err != nil {
			return false
		}
	}

	return fi.IsDir()
}
