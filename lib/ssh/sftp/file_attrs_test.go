// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

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
