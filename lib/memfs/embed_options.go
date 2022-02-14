// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

// EmbedOptions define an options for GoEmbed.
type EmbedOptions struct {
	// CommentHeader define optional comment to be added to the header of
	// generated file, for example copyright holder and/or license.
	// The string value is not checked, whether it's a comment or not, it
	// will rendered as is.
	CommentHeader string

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

	// WithoutModTime if its true, the modification time for all
	// files and directories are not stored inside generated code, instead
	// all files will use the current time when the program is running.
	WithoutModTime bool
}
