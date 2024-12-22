// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package memfs_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs/v2"
)

func ExampleNew() {
	/**
	Let say we have the "testdata" directory,

		testdata/
		├── direct
		│   └── add
		│       ├── file
		│       └── file2
		├── exclude
		│   ├── dir
		│   ├── index-link.css -> ../index.css
		│   ├── index-link.html -> ../index.html
		│   └── index-link.js -> ../index.js
		├── include
		│   ├── dir
		│   ├── index.css -> ../index.css
		│   ├── index.html -> ../index.html
		│   └── index.js -> ../index.js
		├── index.css
		├── index.html
		├── index.js
		└── plain

	Assume that we want to embed all files with extension .css, .html,
	and .js only; but not from directory "exclude".
	We can create the Options like below,
	*/
	opts := &memfs.Options{
		Root:     `./testdata`,
		Includes: []string{`.*/include`, `.*\.(css|html|js)$`},
		Excludes: []string{`.*/exclude`},
	}
	mfs, err := memfs.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	node, err := mfs.Get("/index.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Content of /index.html: %s", node.Content)

	fmt.Printf("List of embedded files: %+v\n", mfs.ListNames())

	_, err = mfs.Get("/exclude/index.html")
	if err != nil {
		fmt.Println(`Error:`, err)
	}

	// Output:
	// Content of /index.html: <html></html>
	// List of embedded files: [/ /direct /direct/add /include /include/dir /include/index.css /include/index.html /include/index.js /index.css /index.html /index.js]
	// Error: Get "/exclude/index.html": file does not exist
}

func ExampleMemFS_Search() {
	opts := &memfs.Options{
		Root: `./testdata`,
	}
	mfs, err := memfs.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	q := []string{`body`}
	results := mfs.Search(q, 0)

	for _, result := range results {
		fmt.Println(`Path:`, result.Path)
		fmt.Printf("Snippets: %q\n", result.Snippets)
	}
	// Unordered output:
	// Path: /include/index.css
	// Snippets: ["body {\n}\n"]
	// Path: /exclude/index-link.css
	// Snippets: ["body {\n}\n"]
	// Path: /index.css
	// Snippets: ["body {\n}\n"]
}

func ExampleMemFS_Watch() {
	var (
		opts memfs.Options
		err  error
	)

	opts.Root, err = os.MkdirTemp(``, `memfs_watch`)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = os.RemoveAll(opts.Root)
	}()

	var mfs *memfs.MemFS

	mfs, err = memfs.New(&opts)
	if err != nil {
		log.Fatal(err)
	}

	var watchOpts = memfs.WatchOptions{
		FileWatcherOptions: watchfs.FileWatcherOptions{
			File:     filepath.Join(opts.Root, memfs.DefaultWatchFile),
			Interval: 50 * time.Millisecond,
		},
		Verbose: true,
	}

	var changesq <-chan []*memfs.Node

	changesq, err = mfs.Watch(watchOpts)
	if err != nil {
		log.Fatal(err)
	}

	var testFile = filepath.Join(opts.Root, `file`)
	err = os.WriteFile(testFile, []byte(`dummy content`), 0600)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(watchOpts.File, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	<-changesq

	_, err = mfs.Get(`/file`)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(testFile)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(watchOpts.File, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	<-changesq
	mfs.StopWatch()
	<-changesq

	// Output:
	// MemFS: file created: "/file"
	// MemFS: file deleted: "/file"
}
