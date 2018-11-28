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

	mfs, err := New(incs, excs)
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
