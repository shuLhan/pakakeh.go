// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestFileAttrs_Mode(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	cases := []struct {
		path string
	}{{
		path: "/etc",
	}, {
		path: "/etc/hosts",
	}}

	for _, c := range cases {
		exp, err := fs.Stat(os.DirFS(filepath.Dir(c.path)), filepath.Base(c.path))
		if err != nil {
			t.Fatal(err)
		}

		got, err := testClient.Stat(c.path)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Stat", exp.Mode(), got.Mode())
	}
}
