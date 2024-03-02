// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maildir

import (
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestCreateFolder(t *testing.T) {
	var (
		dir  = t.TempDir()
		name = "  \r\n"

		folder *Folder
		err    error
	)

	// Case: with empty maildir.

	_, err = CreateFolder(``, `.name`)
	var expError = `CreateFolder: invalid maildir "."`
	test.Assert(t, `With empty maildir`, expError, err.Error())

	// Case: with empty name.

	_, err = CreateFolder(dir, name)
	expError = `CreateFolder: folder name is empty`
	test.Assert(t, `With empty name`, expError, err.Error())

	// Case: with no permission.

	name = `.folder`

	_, err = CreateFolder(`/`, name)
	expError = `CreateFolder: mkdir /.folder: read-only file system`
	test.Assert(t, `With empty name`, expError, err.Error())

	// Case: with valid dir and name.

	folder, err = CreateFolder(dir, name)
	if err != nil {
		t.Fatal(err)
	}
	var expDir = filepath.Join(dir, name)
	test.Assert(t, `Folder.dir`, expDir, folder.dir)
	assertFolder(t, folder)

	// Case: creating folder on existing non empty directory should not
	// failed.

	folder, err = CreateFolder(dir, name)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `Folder.dir`, expDir, folder.dir)
	assertFolder(t, folder)
}
