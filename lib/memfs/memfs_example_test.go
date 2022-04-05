package memfs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
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
	opts := &Options{
		Root:     "./testdata",
		Includes: []string{`.*/include`, `.*\.(css|html|js)$`},
		Excludes: []string{`.*/exclude`},
	}
	mfs, err := New(opts)
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
		fmt.Printf("Error on Get /exclude/index.html: %s\n", err)
	}

	// Output:
	// Content of /index.html: <html></html>
	// List of embedded files: [/ /include /include/index.css /include/index.html /include/index.js /index.css /index.html /index.js]
	// Error on Get /exclude/index.html: file does not exist
}

func ExampleMemFS_Search() {
	opts := &Options{
		Root: "./testdata",
	}
	mfs, err := New(opts)
	if err != nil {
		log.Fatal(err)
	}

	q := []string{"body"}
	results := mfs.Search(q, 0)

	for _, result := range results {
		fmt.Printf("Path: %s\n", result.Path)
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
		watchOpts = WatchOptions{
			Delay: 200 * time.Millisecond,
		}

		mfs  *MemFS
		dw   *DirWatcher
		node *Node
		opts Options
		ns   NodeState
		err  error
	)

	opts.Root, err = ioutil.TempDir("", "memfs_watch")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = os.RemoveAll(opts.Root)
	}()

	mfs, err = New(&opts)
	if err != nil {
		log.Fatal(err)
	}

	dw, err = mfs.Watch(watchOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the goroutine on Watch run.
	time.Sleep(200 * time.Millisecond)

	testFile := filepath.Join(opts.Root, "file")
	err = os.WriteFile(testFile, []byte("dummy content"), 0700)
	if err != nil {
		log.Fatal(err)
	}

	ns = <-dw.C
	fmt.Printf("State: %s\n", ns.State)

	node, err = mfs.Get("/file")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Node: %s: %q\n", node.Path, node.Content)

	err = os.Remove(testFile)
	if err != nil {
		log.Fatal(err)
	}

	ns = <-dw.C
	fmt.Printf("State: %s\n", ns.State)

	node, _ = mfs.Get("/file")
	fmt.Printf("Node: %s: %v\n", ns.Node.Path, node)

	dw.Stop()

	//Output:
	//State: FileStateCreated
	//Node: /file: "dummy content"
	//State: FileStateDeleted
	//Node: /file: <nil>
}
