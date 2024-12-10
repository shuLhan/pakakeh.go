// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/http/httputil"
	"regexp"
	"testing"

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestHTTPHandleTest(t *testing.T) {
	type testCase struct {
		tag         string
		contentType string
		req         Request
	}

	var (
		tdata *test.Data
		err   error
	)
	tdata, err = test.LoadData(`testdata/httpHandleTest_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var listCase = []testCase{{
		tag: `noContentType`,
	}, {
		tag:         `ok`,
		contentType: libhttp.ContentTypeJSON,
		req: Request{
			File: `testdata/test_test.go`,
		},
	}, {
		tag:         `invalidFile`,
		contentType: libhttp.ContentTypeJSON,
		req: Request{
			File: `testdata/notexist/test_test.go`,
		},
	}}

	var (
		rexDuration = regexp.MustCompile(`(?m)\\t(\d+\.\d+)s`)
		tcase       testCase
		rawb        []byte
	)
	for _, tcase = range listCase {
		tcase.req.Body = string(tdata.Input[tcase.tag])

		rawb, err = json.Marshal(&tcase.req)
		if err != nil {
			t.Fatal(err)
		}

		var httpReq = httptest.NewRequest(`POST`, `/`, bytes.NewReader(rawb))
		httpReq.Header.Set(libhttp.HeaderContentType, tcase.contentType)

		var httpWriter = httptest.NewRecorder()

		HTTPHandleTest(httpWriter, httpReq)

		var httpResp = httpWriter.Result()
		rawb, err = httputil.DumpResponse(httpResp, true)
		if err != nil {
			t.Fatal(err)
		}
		rawb = bytes.ReplaceAll(rawb, []byte("\r"), []byte(""))
		rawb = rexDuration.ReplaceAll(rawb, []byte(" Xs"))

		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(rawb))
	}
}
