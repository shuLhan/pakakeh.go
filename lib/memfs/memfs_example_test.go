package memfs

import (
	"fmt"
	"log"
)

func ExampleNew() {
	incs := []string{
		`.*/include`,
		`.*\.(css|html|js)$`,
	}
	excs := []string{
		`.*/exclude`,
	}

	mfs, err := New(incs, excs, true)
	if err != nil {
		log.Fatal(err)
	}

	err = mfs.Mount("./testdata")
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
	mfs, err := New(nil, nil, true)
	if err != nil {
		log.Fatal(err)
	}

	err = mfs.Mount("./testdata")
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
