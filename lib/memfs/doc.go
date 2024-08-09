// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package memfs provide a library for mapping file system into memory and/or
// to embed it inside go source file.
//
// # Usage
//
// The first step is to create new instance of memfs using [New] function.
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
// To increase the default size set the [Options.MaxFileSize] (in bytes).
// For example, to change maximum file size to 10 MB,
//
//	var opts = memfs.Options{
//		MaxFileSize: 1024 * 1024 * 10,
//	}
//
// Later, if we want to get the file from memory, call [MemFS.Get] which
// will return the [Node] object with content can be accessed from field
// "Content".
// If file size is larger than maximum, the content will need to be read
// manually (as long as the file exist on system),
//
//	node, err := mfs.Get("/")
//	if err != nil {
//		// Handle file not exist
//	}
//	if node.mode.IsDir() {
//		// Handle directory.
//	}
//	if node.Content == nil {
//		// Handle large file.
//		node.V, err = os.ReadFile(child.SysPath)
//	}
//	// Do something with content of file system.
//
// # Go embed
//
// The memfs package support embedding the files into Go generated source
// code.
// After we create the MemFS instance, call the [GoEmbed] method to dump all
// directory and files into Go source code.
//
// First, define global variable as container for all files to be embedded
// in the same package as generated code,
//
//	package mypackage
//
//	var myFS *memfs.MemFS
//
// Second, create new instance of MemFS with [Options.Embed] is set,
//
//	var opts = &Options{
//		Embed: EmbedOptions{
//			PackageName:     `mypackage`,
//			VarName:         `myFS`,
//			GoFileName:      `mypackage/embed.go`,
//		},
//		Root: `./mydir`,
//		Includes: []string{
//			`.*/include`,
//			`.*\.(css|html|js)$`,
//		},
//		Excludes: []string{
//			`.*/exclude`,
//		},
//	}
//
//	var mfs *memfs.MemFS
//	mfs, err = memfs.New(opts)
//	...
//
// Third, call method [MemFS.GoEmbed] from the instance,
//
//	mfs.GoEmbed()
//
// This method will create Go file "mypackage/embed.go" that contains all
// path and content of files inside the mfs instance, under package named
// "mypackage".
// Code that can read "myFS" variable then can access any files using
// [MemFS.Get] method, with  "/" as root prefix (not "./mydir").
//
// # Comparison with builtin go:embed
//
// This section list the disadvantages of "go:embed" directive.
//
// The memfs package created on [November 2018], based on my experiences
// maintains the fork of [go-bindata] project.
// The "go:embed" directive introduced into Go tools since
// [Go version 1.16], released February 2021, three years after the first
// release of memfs package.
//
// Given the following directory structure,
//
//	module-root/
//	+-- cmd/prog/main.go
//	+-- _content/
//	     +-- index.adoc
//	     +-- index.html
//	     +-- static/
//	         +-- index.png
//	         +-- index.png
//
// We want to embed the directory "_content" but only html files and all
// files inside the "static/" directory.
//
// Cons #1: The "go:embed" only works if files or directory to be embedded
// is in the same parent directory.
//
// The "go:embed" directive define in "cmd/prog/main.go" will not
// able to embed files in their parent.
// The following code will not compile,
//
//	//go:embed ../../_content/*.html
//	//go:embed ../../_content/static
//	var contentFS embed.FS
//
//	// go build output,
//	// pattern ../../_content/*.html: invalid pattern syntax
//
// If we remove the ".." and execute "go build" from module-root, it will
// still not compile,
//
//	//go:embed _content/*.html
//	//go:embed _content/static
//	var contentFS embed.FS
//
//	// go build or run output,
//	// pattern _content/*.html: no matching files found
//
// The only solution is to create and export the variable "ContentFS" in the
// same parent directory as "_content".
//
// The memfs package does not have this limitation.
// As long as the Go commands are executed from the module-root directory,
// you can define the variable in any packages.
//
// Cons #2: Accessing the embedded file require the original path.
//
// Let say we have embeded the "_content" directory using the following
// syntax,
//
//	//go:embed _content/*.html
//	//go:embed _content/static
//	var ContentFS embed.FS
//
// To access the file "index.html" you need to pass their full path, in this
// case "_content/index.html".
// The path "_content" leaked to the parent FS and not portable.
//
// In the memfs package, the content of [Options.Root] directory can be
// accessed with "/", so it would become "/index.html".
// This design allow flexibility and consistency between modules and
// packages.
// If an external, third-party package accept the MemFS instance and the
// first thing they do is to read all contents of "/" directory, the caller
// can embed any path without have specific prefix or name.
//
// Case example, when we embed SQL files for migration under directory
// "db/migration" using the "go:embed" directive,
//
//	//go:embed db/migration/*.sql
//	var DBMigrationFS embed.FS
//
// and then call the [Migrate] function, it cannot found any ".sql" files
// inside the "/" directory because the files is stored under
// "db/migration/" prefix.
//
// Cons #3: No development mode.
//
// Let say we run our server that served the content from FS instance.
// If we changes the html files, and refresh the browser, the new content
// will not reflected because it serve the content on the first embed.
//
// The memfs package have [Options.TryDirect] that bypass file in memory and
// read directly to the file system.
// This allow quick development when changes only template or non-code
// files.
//
// [November 2018]: https://git.sr.ht/~shulhan/pakakeh.go/commit/05b02c7b
// [go-bindata]: https://github.com/shuLhan/go-bindata
// [Go version 1.16]: https://go.dev/doc/go1.16
// [Migrate]: https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/sql#Client.Migrate
package memfs
