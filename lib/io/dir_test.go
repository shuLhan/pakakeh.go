// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"os"
	"testing"
)

func TestRmdirEmptyAll(t *testing.T) {
	cases := []struct {
		desc        string
		createDir   string
		createFile  string
		path        string
		expExist    string
		expNotExist string
	}{{
		desc:       "With path as file",
		path:       "testdata/file",
		createFile: "testdata/file",
		expExist:   "testdata/file",
	}, {
		desc:      "With empty path",
		createDir: "testdata/a/b/c/d",
		expExist:  "testdata/a/b/c/d",
	}, {
		desc:        "With non empty at middle",
		createDir:   "testdata/a/b/c/d",
		createFile:  "testdata/a/b/file",
		path:        "testdata/a/b/c/d",
		expExist:    "testdata/a/b/file",
		expNotExist: "testdata/a/b/c",
	}, {
		desc:        "With first path not exist",
		createDir:   "testdata/a/b/c",
		path:        "testdata/a/b/c/d",
		expExist:    "testdata/a/b/file",
		expNotExist: "testdata/a/b/c",
	}, {
		desc:        "With non empty at parent",
		createDir:   "testdata/dirempty/a/b/c/d",
		path:        "testdata/dirempty/a/b/c/d",
		expExist:    "testdata",
		expNotExist: "testdata/dirempty",
	}}

	var (
		err error
		f   *os.File
	)
	for _, c := range cases {
		t.Log(c.desc)

		if len(c.createDir) > 0 {
			err = os.MkdirAll(c.createDir, 0700)
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(c.createFile) > 0 {
			f, err = os.Create(c.createFile)
			if err != nil {
				t.Fatal(err)
			}
			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}

		err = RmdirEmptyAll(c.path)
		if err != nil {
			t.Fatal(err)
		}

		if len(c.expExist) > 0 {
			_, err = os.Stat(c.expExist)
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(c.expNotExist) > 0 {
			_, err = os.Stat(c.expNotExist)
			if !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	}
}
