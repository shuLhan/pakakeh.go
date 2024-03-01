// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"fmt"
	"os"
	"path"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParseHexDump(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/ParseHexDump_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tag string
		in  []byte
		exp []byte
		got []byte
	)
	for tag, in = range tdata.Input {
		exp = tdata.Output[tag]

		got, err = ParseHexDump(in)
		if err != nil {
			test.Assert(t, tag, string(exp), err.Error())
			continue
		}

		test.Assert(t, tag, string(exp), string(got))
	}
}

func TestParseHexDumpExpDirTar(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/ParseHexDump_exp_dir_tar_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tag     = `exp_dir.tar`
		expFile = path.Join(`testdata`, tag)

		exp []byte
		got []byte
	)

	got, err = ParseHexDump(tdata.Input[tag])
	if err != nil {
		t.Fatal(err)
	}

	exp, err = os.ReadFile(expFile)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, tag, exp, got)
}

func TestTrimNull(t *testing.T) {
	type testCase struct {
		in  []byte
		exp []byte
	}

	var (
		cases = []testCase{{
			in: []byte{0},
		}, {
			in:  []byte{0, 'H'},
			exp: []byte{'H'},
		}, {
			in:  []byte{'H', 0},
			exp: []byte{'H'},
		}, {
			in:  []byte{'H'},
			exp: []byte{'H'},
		}}

		x   int
		c   testCase
		got []byte
	)

	for x, c = range cases {
		got = TrimNull(c.in)
		test.Assert(t, fmt.Sprintf(`TrimNull #%d`, x), c.exp, got)
	}
}
