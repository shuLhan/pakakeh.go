// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestWatchDirOptions(t *testing.T) {
	listCase := []struct {
		expError string
		opts     DirWatcherOptions
	}{{
		opts: DirWatcherOptions{
			Includes: []string{`.(*\.adoc$`},
		},
		expError: "WatchDir: error parsing regexp: missing argument to repetition operator: `*`",
	}, {
		opts: DirWatcherOptions{
			Excludes: []string{`.(*\.adoc$`},
		},
		expError: "WatchDir: error parsing regexp: missing argument to repetition operator: `*`",
	}}

	for _, tc := range listCase {
		_, err := WatchDir(tc.opts)
		test.Assert(t, `error`, tc.expError, err.Error())
	}
}

func TestWatchDirInitialScan(t *testing.T) {
	listCase := []struct {
		expFiles map[string]os.FileInfo
		desc     string
		opts     DirWatcherOptions
		expIndex DirWatcher
	}{{
		desc: `With includes and excludes`,
		opts: DirWatcherOptions{
			FileWatcherOptions: FileWatcherOptions{
				File:     `testdata/rescan`,
				Interval: 50 * time.Millisecond,
			},
			Root:     `testdata/`,
			Includes: []string{`.*\.adoc$`},
			Excludes: []string{`exc$`, `.*\.html$`, `	`},
		},
		expIndex: DirWatcher{
			idxDir: map[string]node{
				`testdata/`: node{
					name: `testdata/`,
					size: 2,
					mode: fs.ModeDir | 0755,
				},
				`testdata/exc`: nodeExcluded,
				`testdata/inc`: node{
					name: `testdata/inc`,
					size: 3,
					mode: fs.ModeDir | 0755,
				},
			},
			idxFile: map[string]node{
				`testdata/inc/index.adoc`: node{
					name: `testdata/inc/index.adoc`,
					size: 7,
					mode: 0644,
				},
				`testdata/inc/index.css`:  nodeExcluded,
				`testdata/inc/index.html`: nodeExcluded,
			},
		},
		expFiles: map[string]os.FileInfo{
			`testdata/inc/index.adoc`: &node{
				name: `testdata/inc/index.adoc`,
				size: 7,
				mode: 0644,
			},
		},
	}, {
		desc: `With empty includes`,
		opts: DirWatcherOptions{
			FileWatcherOptions: FileWatcherOptions{
				File:     `testdata/rescan`,
				Interval: 50 * time.Millisecond,
			},
			Root:     `testdata/`,
			Includes: []string{`   `},
			Excludes: []string{`exc$`, `.*\.adoc$`},
		},
		expIndex: DirWatcher{
			idxDir: map[string]node{
				`testdata/`: node{
					name: `testdata/`,
					size: 2,
					mode: fs.ModeDir | 0755,
				},
				`testdata/exc`: nodeExcluded,
				`testdata/inc`: node{
					name: `testdata/inc`,
					size: 3,
					mode: fs.ModeDir | 0755,
				},
			},
			idxFile: map[string]node{
				`testdata/inc/index.adoc`: nodeExcluded,
				`testdata/inc/index.css`: node{
					name: `testdata/inc/index.css`,
					mode: 0644,
				},
				`testdata/inc/index.html`: node{
					name: `testdata/inc/index.html`,
					mode: 0644,
				},
			},
		},
		expFiles: map[string]os.FileInfo{
			`testdata/inc/index.css`: &node{
				name: `testdata/inc/index.css`,
				mode: 0644,
			},
			`testdata/inc/index.html`: &node{
				name: `testdata/inc/index.html`,
				mode: 0644,
			},
		},
	}}

	for _, tc := range listCase {
		dwatch, err := WatchDir(tc.opts)
		if err != nil {
			t.Fatal(err)
		}
		dwatch.Stop()
		test.Assert(t, tc.desc+`: idxDir`, tc.expIndex.idxDir,
			dwatch.idxDir)
		test.Assert(t, tc.desc+`: idxFile`, tc.expIndex.idxFile,
			dwatch.idxFile)
		gotFiles := dwatch.Files()
		test.Assert(t, tc.desc+`: Files`, tc.expFiles, gotFiles)
	}
}

func TestWatchDir(t *testing.T) {
	var (
		dirTemp     = t.TempDir()
		fileToWatch = `rescan`
		opts        = DirWatcherOptions{
			FileWatcherOptions: FileWatcherOptions{
				File:     filepath.Join(dirTemp, fileToWatch),
				Interval: 50 * time.Millisecond,
			},
			Root:     dirTemp,
			Includes: []string{`.*\.adoc$`},
			Excludes: []string{`exc$`, `.*\.html$`},
		}
		dwatch *DirWatcher
		err    error
	)

	dwatch, err = WatchDir(opts)
	if err != nil {
		t.Fatal(err)
	}

	var (
		fileAadoc = filepath.Join(opts.Root, `a.adoc`)
		fileBadoc = filepath.Join(opts.Root, `b.adoc`)
	)
	t.Run(`On included files created`, func(t *testing.T) {
		var expNames = []string{
			fileAadoc,
			fileBadoc,
		}
		for _, file := range expNames {
			err = os.WriteFile(file, nil, 0600)
			if err != nil {
				t.Fatal(err)
			}
		}

		dwatch.ForceRescan()

		var gotNames = listFileName(<-dwatch.C)
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On dir excluded created`, func(t *testing.T) {
		err = os.MkdirAll(filepath.Join(opts.Root, `exc`), 0700)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File,
			[]byte(`xx`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames []string
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On file excluded created`, func(t *testing.T) {
		var fileAhtml = filepath.Join(opts.Root, `a.html`)
		err = os.WriteFile(fileAhtml, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File,
			[]byte(`xxx`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames []string
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On sub directory created`, func(t *testing.T) {
		var dirInc = filepath.Join(opts.Root, `inc`)
		err = os.MkdirAll(dirInc, 0700)
		if err != nil {
			t.Fatal(err)
		}
		var fileC = filepath.Join(dirInc, `ccc.adoc`)
		err = os.WriteFile(fileC, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File,
			[]byte(`xxxx`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames = []string{
			fileC,
		}
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On file updated`, func(t *testing.T) {
		err = os.WriteFile(fileAadoc, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File,
			[]byte(`xxxxx`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames = []string{
			fileAadoc,
		}
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On file deleted`, func(t *testing.T) {
		err = os.Remove(fileAadoc)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File,
			[]byte(`xxxxx x`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames = []string{
			fileAadoc,
		}
		test.Assert(t, `changes`, expNames, gotNames)
	})
}

func listFileName(listfi []os.FileInfo) (listName []string) {
	var fi os.FileInfo
	for _, fi = range listfi {
		listName = append(listName, fi.Name())
	}
	sort.Strings(listName)
	return listName
}
