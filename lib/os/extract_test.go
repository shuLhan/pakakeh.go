// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestExtract(t *testing.T) {
	var (
		logp   = "TestExtract"
		tmpDir string
		err    error
	)

	// Directory that store all test output.
	tmpDir, err = os.MkdirTemp("testdata", "extract_")
	if err != nil {
		t.Fatalf("%s: %s", logp, err)
	}

	t.Cleanup(func() {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatalf("%s: %s", logp, err)
		}
	})

	// The expected file content for single file compression.
	expFile, err := os.ReadFile("testdata/exp")
	if err != nil {
		t.Fatalf("%s: %s", logp, err)
	}

	// The expected directory content for archive.
	mfsOpts := memfs.Options{
		Root: "testdata/exp_dir",
	}

	mfsExp, err := memfs.New(&mfsOpts)
	if err != nil {
		t.Fatalf("%s: %s", logp, err)
	}

	expDirContent := map[string]struct{}{
		"/dir_x":        struct{}{},
		"/dir_x/file_x": struct{}{},
		"/file_y":       struct{}{},
	}

	cases := []struct {
		desc      string
		fileInput string
		dirOutput string
		isArchive bool
	}{{
		desc:      "With .bz2",
		fileInput: "exp.bz2",
		dirOutput: "exp.bz2_",
	}, {
		desc:      "With .gz",
		fileInput: "exp.gz",
		dirOutput: "exp.gz_",
	}, {
		desc:      "With .zip file",
		fileInput: "exp.zip",
		dirOutput: "exp.zip_",
	}, {
		desc:      "With .tar",
		fileInput: "exp_dir.tar",
		dirOutput: "exp_dir.tar_",
		isArchive: true,
	}, {
		desc:      "With .zip dir",
		fileInput: "exp_dir.zip",
		dirOutput: "exp_dir.zip_",
		isArchive: true,
	}, {
		desc:      "With .tar.bz2",
		fileInput: "exp_dir.tar.bz2",
		dirOutput: "exp_dir.tar.bz2_",
		isArchive: true,
	}, {
		desc:      "With .tar.gz",
		fileInput: "exp_dir.tar.gz",
		dirOutput: "exp_dir.tar.gz_",
		isArchive: true,
	}}

	for _, c := range cases {
		// Create symlinks for fileInput into "testdata/input/" to prevent the
		// original file being removed after successful Extract.
		linkOrg := filepath.Join("..", c.fileInput)
		linkInput := filepath.Join("testdata", "input", c.fileInput)
		err = os.Symlink(linkOrg, linkInput)
		if err != nil {
			if !os.IsExist(err) {
				t.Fatalf("%s: %s: %s", logp, c.desc, err)
			}
		}

		c.dirOutput, err = os.MkdirTemp(tmpDir, c.dirOutput)
		if err != nil {
			t.Fatalf("%s: %s: %s", logp, c.desc, err)
		}

		err = Extract(linkInput, c.dirOutput)
		if err != nil {
			t.Fatalf("%s: %s: %s", logp, c.desc, err)
		}

		if c.isArchive {
			mfsGotOpts := memfs.Options{
				Root: filepath.Join(c.dirOutput, "exp_dir"),
			}
			mfsGot, err := memfs.New(&mfsGotOpts)
			if err != nil {
				t.Fatalf("%s: %s", logp, err)
			}

			for path := range expDirContent {
				expNode, err := mfsExp.Get(path)
				if err != nil {
					t.Fatalf("%s: %s: %s", logp, path, err)
				}

				gotNode, err := mfsGot.Get(path)
				if err != nil {
					t.Fatalf("%s: %s: %s", logp, path, err)
				}

				// Compare each fs.FileInfo.
				test.Assert(t, fmt.Sprintf("%s: %s: Size", logp, path),
					expNode.Size(), gotNode.Size())
				test.Assert(t, fmt.Sprintf("%s: %s: Mode", logp, path),
					expNode.Mode(), gotNode.Mode())

				// Do not compare the ModTime, because the
				// file modification time will be changes when
				// the repository is cloned.

				test.Assert(t, fmt.Sprintf("%s: %s: IsDir", logp, path),
					expNode.IsDir(), gotNode.IsDir())

				test.Assert(t, fmt.Sprintf("%s: %s: Content", logp, path),
					expNode.Content, gotNode.Content)

			}
		} else {
			gotFile, err := os.ReadFile(filepath.Join(c.dirOutput, "exp"))
			if err != nil {
				t.Fatalf("%s: %s: %s", logp, c.desc, err)
			}

			test.Assert(t, c.desc, expFile, gotFile)
		}
	}
}
