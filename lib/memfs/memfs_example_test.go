// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
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
		watchOpts = memfs.WatchOptions{
			Delay: 200 * time.Millisecond,
		}

		mfs  *memfs.MemFS
		dw   *memfs.DirWatcher
		node *memfs.Node
		opts memfs.Options
		ns   memfs.NodeState
		err  error
	)

	opts.Root, err = os.MkdirTemp(``, `memfs_watch`)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = os.RemoveAll(opts.Root)
	}()

	mfs, err = memfs.New(&opts)
	if err != nil {
		log.Fatal(err)
	}

	dw, err = mfs.Watch(watchOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the goroutine on Watch run.
	time.Sleep(200 * time.Millisecond)

	testFile := filepath.Join(opts.Root, `file`)
	err = os.WriteFile(testFile, []byte(`dummy content`), 0700)
	if err != nil {
		log.Fatal(err)
	}

	ns = <-dw.C

	node, err = mfs.Get(`/file`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Node: %s: %s\n", node.Path, ns.State)

	err = os.Remove(testFile)
	if err != nil {
		log.Fatal(err)
	}

	ns = <-dw.C
	fmt.Printf("Node: %s: %s\n", ns.Node.Path, ns.State)

	dw.Stop()

	//Output:
	//Node: /file: FileStateCreated
	//Node: /file: FileStateDeleted
}
