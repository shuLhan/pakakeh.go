// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package memfs provide a library for mapping file system into memory and/or
// to embed it inside go source file.
//
// Usage
//
// The first step is to create new instance of memfs using `New()`.
// The following example embed all files inside directory named "include" or
// any files with extension ".css", ".html", and ".js";
// but exclude any files inside directory named "exclude".
//
//	opts := &Options{
//		Root: "./mydir",
//		Includes: []string{
//			`.*/include`,
//			`.*\.(css|html|js)$`,
//		},
//		Excludes: []string{
//			`.*/exclude`,
//		},
//	}
//	mfs, err := memfs.New(opts)
//
// By default only file with size less or equal to 5 MB will be included in
// memory.
// To increase the default size set the MaxFileSize (in bytes).
// For example, to set maximum file size to 10 MB,
//
//	opts.MaxFileSize = 1024 * 1024 * 10
//
// Later, if we want to get the file from memory, call Get() which will return
// the node object with content can be accessed from field "V".
// Remember that if file size is larger that maximum,
// the content will need to be read manually,
//
//	node, err := mfs.Get("/")
//	if err != nil {
//		// Handle file not exist
//	}
//	if node.mode.IsDir() {
//		// Handle directory.
//	}
//	if node.V == nil {
//		// Handle large file.
//		node.V, err = ioutil.ReadFile(child.SysPath)
//	}
//	// Do something with content of file system.
//
// Thats it!
//
// Go embed
//
// The memfs package also support embedding the files into Go generated source
// file.
// After we create memfs instance, we call GoEmbed() to dump all directory
// and files as Go source code.
//
// First, define global variable as container for all files later in the same
// package as generated code,
//
//	package mypackage
//
//	var myFS *memfs.MemFS
//
// Second, write the content of memfs instance as Go source code file,
//
//	mfs.GoEmbed("mypackage", "myFS", "mypackage/file.go", memfs.EncodingGzip)
//
// The Go generated file will be defined with package named "mypackage" using
// global variable "myFS" as container stored in
// file "mypackage/file.go" with each content encoded (compressed) using
// gzip.
//
package memfs
