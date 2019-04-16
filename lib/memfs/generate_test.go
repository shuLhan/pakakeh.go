// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"path/filepath"
	"testing"
)

func TestGoGenerate(t *testing.T) {
	mfs, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = mfs.Mount(filepath.Join(_testWD, "/testdata"))
	if err != nil {
		t.Fatal(err)
	}

	err = mfs.GoGenerate("test", "generate_test/memfs_generate.go")
	if err != nil {
		t.Fatal(err)
	}
}
