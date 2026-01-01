// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2025 M. Shulhan <ms@kilabit.info>

package git_test

import (
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/git"
)

func ExampleGit_IsIgnored() {
	var agit *git.Git
	var err error
	agit, err = git.New(`testdata/IsIgnored`)
	if err != nil {
		log.Fatal(err)
	}

	var listPath = []string{
		``,
		`vendor`,
		`vendor/dummy`,
		`hello.html`,
		`hello.go`,
		`foo/hello.go`,
	}
	var path string
	var got bool
	for _, path = range listPath {
		got = agit.IsIgnored(path)
		fmt.Printf("%q: %t\n", path, got)
	}
	// Output:
	// "": true
	// "vendor": true
	// "vendor/dummy": true
	// "hello.html": true
	// "hello.go": false
	// "foo/hello.go": false
}
