package memfs

import (
	"fmt"
	"log"
)

func ExampleNew() {
	opts := &Options{
		Root: "./testdata",
		Includes: []string{
			`.*/include`,
			`.*\.(css|html|js)$`,
		},
		Excludes: []string{
			`.*/exclude`,
		},
		WithContent: true,
	}
	mfs, err := New(opts)
	if err != nil {
		log.Fatal(err)
	}

	node, err := mfs.Get("/index.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", node.V)
	// Output:
	// <html></html>
}

func ExampleMemFS_Search() {
	opts := &Options{
		Root:        "./testdata",
		WithContent: true,
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
	// Path: /exclude/index.css
	// Snippets: ["body {\n}\n"]
	// Path: /index.css
	// Snippets: ["body {\n}\n"]
}
