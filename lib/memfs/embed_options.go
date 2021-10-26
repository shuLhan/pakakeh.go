// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

// EmbedOptions define an options for GoEmbed.
type EmbedOptions struct {
	// The generated package name for GoEmbed().
	// If its not defined it will be default to "main".
	PackageName string

	// VarName is the global variable name with type *memfs.MemFS which
	// will be initialized by generated Go source code on init().
	// If its empty it will default to "memFS".
	VarName string

	// GoFileName the path to Go generated file, where the file
	// system will be embedded.
	// If its not defined it will be default to "memfs_generate.go"
	// in current directory from where its called.
	GoFileName string

	// ContentEncoding if this value is not empty, it will encode the
	// content of node and set the node ContentEncoding.
	//
	// List of available encoding is "gzip".
	//
	// For example, if the value is "gzip" it will compress the content of
	// file using gzip and set Node.ContentEncoding to "gzip".
	ContentEncoding string

	// WithoutModTime if its true, the modification time for all
	// files and directories are not stored inside generated code, instead
	// all files will use the current time when the program is running.
	WithoutModTime bool
}
