// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package memfs

import (
	"testing"
)

func TestMemFS_GoEmbed_WithoutModTime(t *testing.T) {
	opts := &Options{
		Root: "testdata",
		Excludes: []string{
			`^\..*`,
			".*/node_save$",
		},
		Embed: EmbedOptions{
			CommentHeader: `// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>

`,
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
