// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

// Opening the ApoFile should create the file if its not exist, and write
// the header and footer.
func TestOpenApo(t *testing.T) {
	var (
		dir  = t.TempDir()
		path = filepath.Join(dir, `OpenApo_test.bin`)

		apo *ApoFile
		err error
	)

	apo, err = OpenApo(path)
	if err != nil {
		t.Fatal(err)
	}
	err = apo.Close()
	if err != nil {
		t.Fatal(err)
	}

	var tdata *test.Data

	tdata, err = test.LoadData(`testdata/OpenApo_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var gotb []byte
	gotb, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var gotDump bytes.Buffer
	libbytes.DumpPrettyTable(&gotDump, `empty`, gotb)

	var exp = string(tdata.Output[`empty`])
	test.Assert(t, `empty`, exp, gotDump.String())

	// Test reading ...

	var apor *ApoFile
	apor, err = OpenApo(path)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, `ReadApo`, apo, apor)
}

type testCaseWrite struct {
	tag         string
	expHexdump  string
	expMetaData []ApoMetaData
	expFooter   apoFooter
	expHeader   apoHeader
}

type dataWrite struct {
	ID int64
}

func TestApoFileWrite(t *testing.T) {
	tdata, err := test.LoadData(`testdata/ApoFileWrite_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var dir = t.TempDir()
	var path = filepath.Join(dir, `ApoFileWrite_test.apo`)

	apo, err := OpenApo(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = apo.Close()
	})

	var listCase = []testCaseWrite{{
		tag:        `insert`,
		expHexdump: string(tdata.Output[`insert`]),
		expHeader: apoHeader{
			Version:   apoVersionOne,
			TotalData: 1,
			OffFoot:   38,
		},
		expFooter: apoFooter{
			idxMetaOff: []int64{
				17,
			},
		},
		expMetaData: []ApoMetaData{{
			Meta: ApoMeta{
				At: 1735179660000000000,
			},
			Data: &dataWrite{
				ID: 1,
			},
		}},
	}}

	for _, tcase := range listCase {
		t.Run(tcase.tag, func(t *testing.T) {
			testWrite(t, tcase, apo)
		})

		t.Run(tcase.tag+` read`, func(t *testing.T) {
			testRead(t, tcase, apo)
		})
	}
}

func testWrite(t *testing.T, tcase testCaseWrite, apow *ApoFile) {
	for _, md := range tcase.expMetaData {
		err := apow.Write(md.Meta, md.Data)
		if err != nil {
			t.Fatal(err)
		}
	}

	gotb, err := os.ReadFile(apow.name)
	if err != nil {
		t.Fatal(err)
	}

	var gotDump bytes.Buffer
	libbytes.DumpPrettyTable(&gotDump, tcase.tag, gotb)

	test.Assert(t, tcase.tag, tcase.expHexdump, gotDump.String())
}

func testRead(t *testing.T, tcase testCaseWrite, apow *ApoFile) {
	apor, err := OpenApo(apow.name)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = apor.Close()
	})

	test.Assert(t, `header`, tcase.expHeader, apor.head)
	test.Assert(t, `footer`, tcase.expFooter, apor.foot)

	var data dataWrite
	gotMetaData, err := apor.ReadAll(&data)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `meta-data`, tcase.expMetaData, gotMetaData)
}
