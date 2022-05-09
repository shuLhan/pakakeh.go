// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

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

	// MaxFileSize define maximum file size that can be stored on memory.
	// The default value is 5 MB.
	// If its value is negative, the content of file will not be mapped to
	// memory, the MemFS will behave as directory tree.
	MaxFileSize int64

	// Development define a flag to bypass file in memory.
	// If its true, any call to Get will result in direct read to file
	// system.
	Development bool
}

// init initialize the options with default value.
func (opts *Options) init() {
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = defaultMaxFileSize
	}
}
