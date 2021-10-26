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
		Embed: EmbedOptions{
			PackageName: "embed",
			GoFileName:  "./embed_test/embed_test.go",
		},
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	err = mfs.GoEmbed()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemFS_GoEmbed_DisableModTime(t *testing.T) {
	opts := &Options{
		Root: "testdata",
		Excludes: []string{
			`^\..*`,
			".*/node_save$",
		},
		Embed: EmbedOptions{
			PackageName:    "embed",
			GoFileName:     "./internal/test/embed_disable_modtime/embed_test.go",
			WithoutModTime: true,
		},
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	err = mfs.GoEmbed()
	if err != nil {
		t.Fatal(err)
	}
}
