// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

package memfs

import (
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestOptionsIsIncluded(t *testing.T) {
	type testData struct {
		sysPath string
		exp     bool
	}

	var cases = []struct {
		data []testData
		desc string
		opts Options
	}{{
		desc: `With empty includes and excludes`,
		data: []testData{{
			sysPath: filepath.Join(_testWD, `/testdata`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/index.html`),
			exp:     true,
		}},
	}, {
		desc: `With excludes only`,
		opts: Options{
			Excludes: []string{
				`.*/exclude`,
				`.*\.html$`,
			},
		},
		data: []testData{{
			sysPath: filepath.Join(_testWD, `/testdata`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/exclude`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/exclude/dir`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/index.html`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/index.css`),
			exp:     true,
		}},
	}, {
		desc: `With includes only`,
		opts: Options{
			Includes: []string{
				`.*/include`,
				`.*\.html$`,
			},
		},
		data: []testData{{
			sysPath: filepath.Join(_testWD, `/testdata`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include/dir`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/index.html`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/index.css`),
			exp:     false,
		}},
	}, {
		desc: `With excludes and includes`,
		opts: Options{
			Excludes: []string{
				`.*/exclude`,
				`.*\.js$`,
			},
			Includes: []string{
				`.*/include`,
				`.*\.(css|html)$`,
			},
		},
		data: []testData{{
			sysPath: filepath.Join(_testWD, `/testdata`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/index.html`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/index.css`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/exclude`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/exclude/dir`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/exclude/index-link.css`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/exclude/index-link.html`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/exclude/index-link.js`),
			exp:     false,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include/dir`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include/index.css`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include/index.html`),
			exp:     true,
		}, {
			sysPath: filepath.Join(_testWD, `/testdata/include/index.js`),
			exp:     false,
		}},
	}}

	var (
		fi    os.FileInfo
		tdata testData
		err   error
		got   bool
	)
	for _, c := range cases {
		t.Log(c.desc)

		err = c.opts.init()
		if err != nil {
			t.Fatal(err)
		}

		for _, tdata = range c.data {
			fi, err = os.Stat(tdata.sysPath)
			if err != nil {
				t.Fatal(err)
			}

			got = c.opts.isExcluded(tdata.sysPath)
			if got {
				test.Assert(t, tdata.sysPath, !tdata.exp, got)
			} else {
				got = c.opts.isIncluded(tdata.sysPath, fi)
				test.Assert(t, tdata.sysPath, tdata.exp, got)
			}
		}
	}
}
