// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2025 M. Shulhan <ms@kilabit.info>

package git_test

import (
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/git"
)

func ExampleGitignore_IsIgnored() {
	var ign = git.Gitignore{}

	ign.Parse(`testdata/IsIgnored/`, []byte(`# comment
  # comment
  vendor/  # Ignore vendor directory, but not vendor file.
/hello.*    # Ignore hello at root, but not foo/hello.go.
!hello.go`))

	var listPath = []string{
		``,
		`vendor`,
		`vendor/dummy`,
		`hello.html`,
		`hello.go`,
		`foo/hello.go`,
		`foo/vendor`,
	}
	for _, path := range listPath {
		fmt.Printf("%q: %t\n", path, ign.IsIgnored(path))
	}
	// Output:
	// "": true
	// "vendor": true
	// "vendor/dummy": true
	// "hello.html": true
	// "hello.go": false
	// "foo/hello.go": false
	// "foo/vendor": false
}
