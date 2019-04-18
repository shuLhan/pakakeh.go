// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/memfs"
	"github.com/shuLhan/share/lib/test"
)

func TestGeneratePathNode(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	wd = strings.TrimSuffix(wd, "generate_test")

	cases := []struct {
		path     string
		exp      *memfs.Node
		expError string
	}{{
		path:     "/memfs_generate.go",
		expError: "file does not exist",
	}, {
		path: "/",
		exp: &memfs.Node{
			SysPath:     filepath.Join(wd, "testdata"),
			Path:        "/",
			Name:        "/",
			ContentType: "",
			Mode:        2147484141,
			Size:        4096,
			V:           []byte{},
		},
	}, {
		path: "/exclude/index.html",
		exp: &memfs.Node{
			SysPath:     filepath.Join(wd, "testdata", "exclude", "index.html"),
			Path:        "/exclude/index.html",
			Name:        "index.html",
			ContentType: "text/html; charset=utf-8",
			Mode:        420,
			Size:        14,
			V:           []byte("<html></html>\n"),
		},
	}}

	mfs, err := memfs.New(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c.path)

		got, err := mfs.Get(c.path)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}

		test.Assert(t, "Node", c.exp, got, true)
	}
}
