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

func TestGo_HTTPHandleFormat(t *testing.T) {
	type testCase struct {
		tag         string
		contentType string
	}

	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/httpHandleFormat_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var listCase = []testCase{{
		tag: `invalid_content_type`,
	}, {
		tag:         `no_package`,
		contentType: libhttp.ContentTypeJSON,
	}, {
		tag:         `indent_and_missing_import`,
		contentType: libhttp.ContentTypeJSON,
	}}

	var playgo *Go
	playgo, err = NewGo(GoOptions{
		Root: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}

	var tcase testCase
	for _, tcase = range listCase {
		var req Request
		req.Body = string(tdata.Input[tcase.tag])

		var rawb []byte
		rawb, err = json.Marshal(&req)
		if err != nil {
			t.Fatal(err)
		}

		var httpReq = httptest.NewRequest(`POST`, `/`,
			bytes.NewReader(rawb))
		httpReq.Header.Set(libhttp.HeaderContentType, tcase.contentType)

		var httpWriter = httptest.NewRecorder()

		playgo.HTTPHandleFormat(httpWriter, httpReq)

		var result = httpWriter.Result()
		rawb, err = httputil.DumpResponse(result, true)
		if err != nil {
			t.Fatal(err)
		}
		rawb = bytes.ReplaceAll(rawb, []byte("\r"), []byte(""))

		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(rawb))
	}
}

func TestGo_HTTPHandleRun(t *testing.T) {
	type testCase struct {
		tag         string
		contentType string
		req         Request
	}

	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/httpHandleRun_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var listCase = []testCase{{
		tag: `no-content-type`,
	}, {
		tag:         `helloworld`,
		contentType: libhttp.ContentTypeJSON,
	}, {
		tag:         `nopackage`,
		contentType: libhttp.ContentTypeJSON,
	}, {
		tag:         `nopackage`,
		contentType: libhttp.ContentTypeJSON,
	}, {
		tag:         `go121_for`,
		contentType: libhttp.ContentTypeJSON,
		req: Request{
			GoVersion:   `1.21.13`,
			WithoutRace: true,
		},
	}}

	var playgo *Go
	playgo, err = NewGo(GoOptions{
		Root: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}

	var tcase testCase
	for _, tcase = range listCase {
		tcase.req.Body = string(tdata.Input[tcase.tag])

		var rawb []byte
		rawb, err = json.Marshal(&tcase.req)
		if err != nil {
			t.Fatal(err)
		}

		var httpReq = httptest.NewRequest(`POST`, `/`,
			bytes.NewReader(rawb))
		httpReq.Header.Set(libhttp.HeaderContentType,
			tcase.contentType)

		var httpWriter = httptest.NewRecorder()

		playgo.HTTPHandleRun(httpWriter, httpReq)

		var result = httpWriter.Result()
		rawb, err = httputil.DumpResponse(result, true)
		if err != nil {
			t.Fatal(err)
		}
		rawb = bytes.ReplaceAll(rawb, []byte("\r"), []byte(""))

		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(rawb))
	}
}

func TestGo_HTTPHandleTest(t *testing.T) {
	type testCase struct {
		tag         string
		contentType string
		req         Request
	}

	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/httpHandleTest_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var playgo *Go
	playgo, err = NewGo(GoOptions{
		Root: `testdata/`,
	})
	if err != nil {
		t.Fatal(err)
	}

	var listCase = []testCase{{
		tag: `noContentType`,
	}, {
		tag:         `ok`,
		contentType: libhttp.ContentTypeJSON,
		req: Request{
			File: `/test_test.go`,
		},
	}, {
		tag:         `invalidFile`,
		contentType: libhttp.ContentTypeJSON,
		req: Request{
			File: `/notexist/test_test.go`,
		},
	}}

	var rexDuration = regexp.MustCompile(`(?m)\\t(\d+\.\d+)s`)
	var empty = []byte(``)
	var tcase testCase
	for _, tcase = range listCase {
		tcase.req.Body = string(tdata.Input[tcase.tag])

		var rawb []byte
		rawb, err = json.Marshal(&tcase.req)
		if err != nil {
			t.Fatal(err)
		}

		var httpReq = httptest.NewRequest(`POST`, `/`,
			bytes.NewReader(rawb))
		httpReq.Header.Set(libhttp.HeaderContentType,
			tcase.contentType)

		var httpWriter = httptest.NewRecorder()

		playgo.HTTPHandleTest(httpWriter, httpReq)

		var httpResp = httpWriter.Result()
		rawb, err = httputil.DumpResponse(httpResp, true)
		if err != nil {
			t.Fatal(err)
		}
		rawb = bytes.ReplaceAll(rawb, []byte("\r"), empty)
		rawb = bytes.ReplaceAll(rawb, []byte(playgo.opts.absRoot),
			empty)
		rawb = rexDuration.ReplaceAll(rawb, []byte(" Xs"))

		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(rawb))
	}
}
