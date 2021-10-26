// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

const (
	defaultMaxFileSize = 1024 * 1024 * 5
)

type Options struct {
	// Root contains path to directory where its contents will be mapped
	// to memory.
	Root string

	// The includes and excludes pattern applied to path of file in file
	// system.
	Includes []string
	Excludes []string

	// MaxFileSize define maximum file size that can be stored on memory.
	// The default value is 5 MB.
	// If its value is negative, the content of file will not be mapped to
	// memory, the MemFS will behave as directory tree.
	MaxFileSize int64

	// EmbedOptions for GoEmbed.
	Embed EmbedOptions

	// Development define a flag to bypass file in memory.
	// If its true, any call to Get will result in direct read to file
	// system.
	Development bool
}

//
// init initialize the options with default value.
//
func (opts *Options) init() {
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = defaultMaxFileSize
	}
}
