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
			Excludes: []string{`exc$`, `.*\.html$`},
		},
		expIndex: DirWatcher{
			idxDir: map[string]node{
				`testdata/`: node{
					name: `testdata/`,
					size: 12,
					mode: fs.ModeDir | 0755,
				},
				`testdata/exc`: nodeExcluded,
				`testdata/inc`: node{
					name: `testdata/inc`,
					size: 58,
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
	}, {
		desc: `With empty includes`,
		opts: DirWatcherOptions{
			FileWatcherOptions: FileWatcherOptions{
				File:     `testdata/rescan`,
				Interval: 50 * time.Millisecond,
			},
			Root:     `testdata/`,
			Excludes: []string{`exc$`, `.*\.adoc$`},
		},
		expIndex: DirWatcher{
			idxDir: map[string]node{
				`testdata/`: node{
					name: `testdata/`,
					size: 12,
					mode: fs.ModeDir | 0755,
				},
				`testdata/exc`: nodeExcluded,
				`testdata/inc`: node{
					name: `testdata/inc`,
					size: 58,
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
	}}

	for _, tc := range listCase {
		dwatch, err := WatchDir(tc.opts)
		if err != nil {
			t.Fatal(err)
		}
		dwatch.Stop()
		test.Assert(t, tc.desc+`: idxDir`, tc.expIndex.idxDir, dwatch.idxDir)
		test.Assert(t, tc.desc+`: idxFile`, tc.expIndex.idxFile, dwatch.idxFile)
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
		fileAaa = filepath.Join(opts.Root, `aaa.adoc`)
		fileBbb = filepath.Join(opts.Root, `bbb.adoc`)
	)
	t.Run(`On file created`, func(t *testing.T) {
		err = os.WriteFile(fileAaa, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fileBbb, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File, []byte(`created`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames = []string{
			fileAaa,
			fileBbb,
		}
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On dir excluded created`, func(t *testing.T) {
		err = os.MkdirAll(filepath.Join(opts.Root, `exc`), 0700)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames []string
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On file excluded created`, func(t *testing.T) {
		err = os.WriteFile(filepath.Join(opts.Root, `aaa.html`), nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File, nil, 0600)
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
		var fileCcc = filepath.Join(dirInc, `ccc.adoc`)
		err = os.WriteFile(fileCcc, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames = []string{
			fileCcc,
		}
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On file updated`, func(t *testing.T) {
		err = os.WriteFile(fileAaa, nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File, []byte(`updated`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames = []string{
			fileAaa,
		}
		test.Assert(t, `changes`, expNames, gotNames)
	})

	t.Run(`On file deleted`, func(t *testing.T) {
		err = os.Remove(fileAaa)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(opts.FileWatcherOptions.File, []byte(`updated`), 0600)
		if err != nil {
			t.Fatal(err)
		}

		var gotNames []string = listFileName(<-dwatch.C)
		var expNames = []string{
			fileAaa,
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
