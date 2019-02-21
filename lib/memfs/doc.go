// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package memfs provide a library for mapping file system into memory.
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
//	mfs, err := memfs.New(incs, excs)
//
// and then we mount the system directory that we want into memory using
// "Mount()",
//
//	err := mfs.Mount("./testdata")
//
// Later, if we want to get the file from memory, call "Get()" and access the
// content with "node.V".  Remember that if file size is larger that maximum,
// the content will need to be read manually,
//
//	node, err := mfs.Get()
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
// After we Mount the directory, we can call,
//
//	mfs.GoGenerate("mypackage", "output/path/file.go")
//
// The Go generate file will be defined with package named "mypackage" in file
// "output/path/file.go".
//
package memfs
