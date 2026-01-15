// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

package http

import (
	"crypto/rand"
	"mime/multipart"
	"strings"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test/mock"
)

func TestGenerateFormData(t *testing.T) {
	type testcase struct {
		form           multipart.Form
		field2filename map[string]string
		tagOutput      string
	}

	rand.Reader = mock.NewRandReader([]byte(`randomseed`))

	var (
		tdata *test.Data
		err   error
	)
	tdata, err = test.LoadData(`testdata/GenerateFormData_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var listcase = []testcase{{
		form: multipart.Form{
			Value: map[string][]string{
				`field1`: []string{`value1`, `value1.1`},
			},
			File: map[string][]*multipart.FileHeader{},
		},
		field2filename: map[string]string{
			`field0`: `file0`,
		},
		tagOutput: `file0`,
	}}

	var (
		tcase  testcase
		listFH []*multipart.FileHeader
		fh     *multipart.FileHeader

		fieldname      string
		filename       string
		gotContentType string
		gotBody        string
		tag            string
	)

	for _, tcase = range listcase {
		for fieldname, filename = range tcase.field2filename {
			fh, err = CreateMultipartFileHeader(filename, tdata.Input[filename])
			if err != nil {
				t.Fatal(err)
			}

			listFH = tcase.form.File[fieldname]
			listFH = append(listFH, fh)
			tcase.form.File[fieldname] = listFH
		}

		gotContentType, gotBody, err = GenerateFormData(&tcase.form)
		if err != nil {
			t.Fatal(err)
		}

		tag = tcase.tagOutput + `.ContentType`
		test.Assert(t, tag, string(tdata.Output[tag]), gotContentType)

		gotBody = strings.ReplaceAll(gotBody, "\r\n", "\n")
		tag = tcase.tagOutput + `.Body`
		test.Assert(t, tag, string(tdata.Output[tag]), gotBody)
	}
}
