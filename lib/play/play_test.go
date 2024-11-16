// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestMain(m *testing.M) {
	now = func() int64 {
		return 10_000_000_000
	}
	os.Exit(m.Run())
}

func TestFormat(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/format_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		req   Request
		name  string
		exp   string
		input []byte
		got   []byte
	)
	for name, input = range tdata.Input {
		req.Body = string(input)
		exp = string(tdata.Output[name])

		got, err = Format(req)
		if err != nil {
			test.Assert(t, name, exp, string(got))
			exp = string(tdata.Output[name+`:error`])
			test.Assert(t, name+`:error`, exp, err.Error())
			continue
		}

		test.Assert(t, name, exp, string(got))
	}
}

func TestHTTPHandleFormat(t *testing.T) {
	type testCase struct {
		tag         string
		contentType string
	}

	var (
		tdata *test.Data
		err   error
	)
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

	var (
		withBody = true

		req   Request
		tcase testCase
		rawb  []byte
		body  bytes.Buffer
	)
	for _, tcase = range listCase {
		req.Body = string(tdata.Input[tcase.tag])

		rawb, err = json.Marshal(&req)
		if err != nil {
			t.Fatal(err)
		}
		body.Reset()
		body.Write(rawb)

		var req *http.Request = httptest.NewRequest(`POST`, `/`, &body)
		req.Header.Set(libhttp.HeaderContentType, tcase.contentType)

		var writer *httptest.ResponseRecorder = httptest.NewRecorder()

		HTTPHandleFormat(writer, req)

		var result *http.Response = writer.Result()
		rawb, err = httputil.DumpResponse(result, withBody)
		if err != nil {
			t.Fatal(err)
		}
		rawb = bytes.ReplaceAll(rawb, []byte("\r"), []byte(""))

		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(rawb))
	}
}

func TestHTTPHandleRun(t *testing.T) {
	type testCase struct {
		tag         string
		contentType string
		req         Request
	}

	var (
		tdata *test.Data
		err   error
	)
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

	var (
		withBody = true

		tcase testCase
		rawb  []byte
		body  bytes.Buffer
	)
	for _, tcase = range listCase {
		tcase.req.Body = string(tdata.Input[tcase.tag])

		rawb, err = json.Marshal(&tcase.req)
		if err != nil {
			t.Fatal(err)
		}
		body.Reset()
		body.Write(rawb)

		var httpReq *http.Request = httptest.NewRequest(`POST`, `/`, &body)
		httpReq.Header.Set(libhttp.HeaderContentType, tcase.contentType)

		var writer *httptest.ResponseRecorder = httptest.NewRecorder()

		HTTPHandleRun(writer, httpReq)

		var result *http.Response = writer.Result()
		rawb, err = httputil.DumpResponse(result, withBody)
		if err != nil {
			t.Fatal(err)
		}
		rawb = bytes.ReplaceAll(rawb, []byte("\r"), []byte(""))

		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(rawb))
	}
}

func TestRun(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/run_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		req = &Request{
			cookieSid: &http.Cookie{},
		}
		sid   string
		exp   string
		input []byte
		got   []byte
	)
	for sid, input = range tdata.Input {
		req.cookieSid.Value = sid
		req.Body = string(input)
		got, err = Run(req)
		if err != nil {
			exp = string(tdata.Output[sid+`-error`])
			test.Assert(t, sid+`-error`, exp, err.Error())
		}
		exp = string(tdata.Output[sid])
		test.Assert(t, sid, exp, string(got))
	}
}

// TestRunOverlap execute Run multiple times.
// The first Run, run the code with infinite loop.
// The second Run, run normal code.
// On the second Run, the first Run should be cancelled or killed.
func TestRunOverlap(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/run_overlap_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	// First Run.
	var (
		sid   = `overlap`
		runwg sync.WaitGroup
	)

	runwg.Add(1)
	go testRunOverlap(t, &runwg, tdata, `run1`, sid)
	time.Sleep(200 * time.Millisecond)

	var cmd1 = runningCmd.get(sid)
	if cmd1 == nil {
		t.Fatal(`expecting cmd1, got nil`)
	}
	var cmd1Pid int = <-cmd1.pid

	// Second Run.

	runwg.Add(1)
	go testRunOverlap(t, &runwg, tdata, `run2`, sid)
	time.Sleep(200 * time.Millisecond)

	// The cmd1 Run should have been killed.
	var proc *os.Process
	proc, err = os.FindProcess(cmd1Pid)
	if err != nil {
		t.Fatalf(`find process: %s`, err)
	}

	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		var exp = os.ErrProcessDone.Error()
		test.Assert(t, `signal error`, exp, err.Error())
	}

	runwg.Wait()
}

func testRunOverlap(t *testing.T, runwg *sync.WaitGroup, tdata *test.Data,
	runName, sid string,
) {
	// In case the test hang, we found that moving [WaitGroup.Done] to
	// the top and call it using defer fix the issue.
	defer runwg.Done()

	var (
		req = &Request{
			cookieSid: &http.Cookie{
				Value: sid,
			},
			Body: string(tdata.Input[runName]),
		}
		exp string
		out []byte
		err error
	)

	out, err = Run(req)
	if err != nil {
		exp = string(tdata.Output[runName+`-error`])
		test.Assert(t, runName+` error`, exp, err.Error())
	}

	exp = string(tdata.Output[runName+`-output`])

	// On Inspiron PC, the test run and can be checked using
	// [test.Assert].
	// On Yoga laptop, the test output is only "signal: killed" and the
	// test hang after [test.Assert], so we replace it with
	// [strings.Contains] here.

	if !strings.Contains(string(out), exp) {
		t.Errorf("%s output: expecting:\n%s\ngot:\n%s", runName,
			exp, string(out))
	}
}
