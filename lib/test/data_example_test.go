// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2025 M. Shulhan <ms@kilabit.info>

package test_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func ExampleData_ExtractInput() {
	var tempDir string
	var err error
	tempDir, err = os.MkdirTemp(``, ``)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	var data = &test.Data{
		Input: map[string][]byte{
			`dir/a.txt`:       []byte(`Content of dir/a.txt.`),
			`dir/sub/b.txt`:   []byte(`Content of dir/sub/b.txt.`),
			`c.txt`:           []byte(`Content of c.txt.`),
			`dir/../../d.txt`: []byte(`Content of d.txt.`),
		},
	}

	// The ExtractInput will create the following directory structures,
	// including their files,
	//
	// ├── c.txt
	// ├── dir
	// │   ├── a.txt
	// │   └── sub
	// │       └── b.txt
	// └── d.txt
	//
	err = data.ExtractInput(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	var listExtractedInput = []string{
		filepath.Join(tempDir, `dir/a.txt`),
		filepath.Join(tempDir, `dir/sub/b.txt`),
		filepath.Join(tempDir, `c.txt`),
		// Since the path of "dir/../../d.txt" is outside of the
		// tempDir, the file will be created on root of tempDir.
		filepath.Join(tempDir, `d.txt`),
	}
	var got []byte
	for _, path := range listExtractedInput {
		got, err = os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", got)
	}
	// Output:
	// Content of dir/a.txt.
	// Content of dir/sub/b.txt.
	// Content of c.txt.
	// Content of d.txt.
}
