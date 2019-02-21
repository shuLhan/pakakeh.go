// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdata

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestGet(t *testing.T) {
	cases := []struct {
		path string
		exp  *Node
	}{{
		path: "/memfs_generate.go",
	}, {
		path: "/",
		exp: &Node{
			SysPath:     "/home/ms/src/github.com/shuLhan/share/lib/memfs/testdata",
			Path:        "/",
			Name:        "/",
			ContentType: "",
			Mode:        2147484141,
			Size:        4096,
			V:           []byte{},
		},
	}, {
		path: "/exclude/index.html",
		exp: &Node{
			SysPath:     "/home/ms/src/github.com/shuLhan/share/lib/memfs/testdata/exclude/index.html",
			Path:        "/exclude/index.html",
			Name:        "index.html",
			ContentType: "text/html; charset=utf-8",
			Mode:        420,
			Size:        14,
			V: []byte{
				60, 104, 116, 109, 108, 62, 60, 47, 104, 116, 109, 108, 62, 10,
			},
		},
	}}

	for _, c := range cases {
		t.Log(c.path)

		got := Get(c.path)

		test.Assert(t, "Node", c.exp, got, true)
	}
}
