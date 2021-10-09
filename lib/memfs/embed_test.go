// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import "testing"

func TestMemFS_GoEmbed(t *testing.T) {
	opts := &Options{
		Root: "testdata",
		Excludes: []string{
			`^\..*`,
			".*/node_save$",
		},
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	err = mfs.GoEmbed("embed", "", "./embed_test/memfs_embed.go", "")
	if err != nil {
		t.Fatal(err)
	}
}
