// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maildir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewFilename(t *testing.T) {
	var fname = createFilename(1000, 2, `localhost`)

	var expNameTmp = `1684640949.M875494_P1000_Q2.localhost`
	test.Assert(t, `nameTmp`, expNameTmp, fname.nameTmp)
}

func TestFilename_generateNewName(t *testing.T) {
	var fname = createFilename(1000, 2, `localhost`)

	// Case: with tmp file not exists.
	var err error

	_, err = fname.generateNameNew(``, 0)
	var expError = `generateNameNew: file does not exist`
	test.Assert(t, `With file not exists`, expError, err.Error())

	// Case: with tmp file exists.
	var (
		content = []byte(`content of file`)
		pathTmp = filepath.Join(t.TempDir(), fname.nameTmp)
		nameNew string
	)

	err = os.WriteFile(pathTmp, content, 0600)
	if err != nil {
		t.Fatal(err)
	}

	nameNew, err = fname.generateNameNew(pathTmp, 0)
	if err != nil {
		t.Fatal(err)
	}

	var expNameNew = `1684640949.M875494_P1000_V36_I170430_Q2.localhost,S=15:2`
	test.Assert(t, `generateNameNew`, expNameNew, nameNew)
}
