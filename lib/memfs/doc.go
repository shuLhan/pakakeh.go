// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package memfs provide a library for mapping file system into memory and to
// generate a go file.
//
// Usage
//
// By default only file with size less or equal to 5 MB will be included in
// memory.  To increase the default size set the MaxFileSize (in bytes).  For
// example, to set maximum file size to 10 MB,
//
//	memfs.MaxFileSize = 1024 * 1024 * 10
//
// The first step is to create new instance of memfs using "New()".
//
//	incs := []string{
//		`.*/include`,
//		`.*\.(css|html|js)$`,
//	}
//	excs := []string{
//		`.*/exclude`,
//	}
//	withContent := true
//	mfs, err := memfs.New("./mystaticsite", incs, excs, withContent)
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
// Go Generate
//
// memfs also support generating the files into Go generated source file.
// After we create memfs instance, we call GoGenerate() to dump all directory
// and files as Go source code,
//
//	mfs.GoGenerate("mypackage", "output/path/file.go", memfs.EncodingGzip)
//
// The Go generated file will be defined with package named "mypackage" in
// file "output/path/file.go" with each content encoded (compressed) using
// gzip.
//
package memfs
