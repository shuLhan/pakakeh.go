// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package hexdump_test

import (
	"os"
	"path"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/hexdump"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParse(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/Parse_test.txt`)
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

		got, err = hexdump.Parse(in, false)
		if err != nil {
			test.Assert(t, tag, string(exp), err.Error())
			continue
		}

		test.Assert(t, tag, string(exp), string(got))
	}
}

func TestParseExpDirTar(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/Parse_exp_dir_tar_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tag     = `exp_dir.tar`
		expFile = path.Join(`testdata`, tag)

		exp []byte
		got []byte
	)

	got, err = hexdump.Parse(tdata.Input[tag], false)
	if err != nil {
		t.Fatal(err)
	}

	exp, err = os.ReadFile(expFile)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, tag, exp, got)
}
