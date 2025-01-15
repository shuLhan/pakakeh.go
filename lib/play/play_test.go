// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestMain(m *testing.M) {
	now = func() int64 {
		return 10_000_000_000
	}
	os.Exit(m.Run())
}

func TestGo_Format(t *testing.T) {
	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/format_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var playgo *Go
	playgo, err = NewGo(GoOptions{
		Root: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}

	var name string
	var input []byte
	for name, input = range tdata.Input {
		var req = Request{
			Body: string(input),
		}

		var exp = string(tdata.Output[name])
		var got []byte

		got, err = playgo.Format(req)
		if err != nil {
			test.Assert(t, name, exp, string(got))
			exp = string(tdata.Output[name+`:error`])
			test.Assert(t, name+`:error`, exp, err.Error())
			continue
		}

		test.Assert(t, name, exp, string(got))
	}
}

func TestGo_Run(t *testing.T) {
	type testCase struct {
		tag string
		req Request
	}

	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/run_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var listCase = []testCase{{
		tag: `nopackage`,
	}, {
		tag: `noimport`,
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

		var got []byte
		got, err = playgo.Run(&tcase.req)
		if err != nil {
			var tagError = tcase.tag + `Error`
			var exp = string(tdata.Output[tagError])
			test.Assert(t, tagError, exp, err.Error())
		}
		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(got))
	}
}

// TestGo_Run_Overlap execute Run multiple times.
// The first Run, run the code with infinite loop.
// The second Run, run normal code.
// On the second Run, the first Run should be cancelled or killed.
func TestGo_Run_Overlap(t *testing.T) {
	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/run_overlap_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	// First Run.
	var sid = `overlap`
	var runwg sync.WaitGroup

	runwg.Add(1)
	go testRunOverlap(t, &runwg, tdata, `run1`, sid)
	time.Sleep(200 * time.Millisecond)

	var cmd1 = runningCmd.get(sid)
	if cmd1 == nil {
		t.Fatal(`expecting cmd1, got nil`)
	}

	// Second Run.

	runwg.Add(1)
	go testRunOverlap(t, &runwg, tdata, `run2`, sid)

	runwg.Wait()

	// The cmd1 Run should have been killed.
	var proc *os.Process
	proc, err = os.FindProcess(cmd1.execGoRun.Process.Pid)
	if err != nil {
		t.Fatalf(`find process: %s`, err)
	}

	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		var exp = os.ErrProcessDone.Error()
		test.Assert(t, `signal error`, exp, err.Error())
	}
}

func testRunOverlap(t *testing.T, runwg *sync.WaitGroup, tdata *test.Data,
	runName, sid string,
) {
	// In case the test hang, we found that moving [WaitGroup.Done] to
	// the top and call it using defer fix the issue.
	defer runwg.Done()

	var playgo *Go
	var err error
	playgo, err = NewGo(GoOptions{
		Root: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}

	var req = &Request{
		cookieSid: &http.Cookie{
			Value: sid,
		},
		Body: string(tdata.Input[runName]),
	}
	var out []byte
	out, err = playgo.Run(req)
	if err != nil {
		var exp = string(tdata.Output[runName+`-error`])
		test.Assert(t, runName+` error`, exp, err.Error())
	}

	var exp = string(tdata.Output[runName+`-output`])

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

func TestGo_Run_UnsafeRun(t *testing.T) {
	var req = &Request{
		UnsafeRun: `/unsafe_run/cmd/forum`,
	}

	var playgo *Go
	var err error
	playgo, err = NewGo(GoOptions{
		Root: `testdata/`,
	})
	if err != nil {
		t.Fatal(err)
	}

	var out []byte
	out, err = playgo.Run(req)
	if err != nil {
		t.Fatal(err)
	}

	var exp = "Hello...\n"
	test.Assert(t, `unsafeRun`, exp, string(out))
}

func TestGo_Test(t *testing.T) {
	type testCase struct {
		tag      string
		exp      string
		expError string
		treq     Request
	}

	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/test_test.txt`)
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
		tag: `ok`,
		treq: Request{
			File: `/test_test.go`,
		},
	}, {
		tag: `fail`,
		treq: Request{
			File: `/test_test.go`,
		},
	}, {
		tag: `buildFailed`,
		treq: Request{
			File: `/test_test.go`,
		},
	}, {
		tag:      `emptyFile`,
		expError: ErrEmptyFile.Error(),
	}, {
		tag: `ErrPermission`,
		treq: Request{
			File: `../../etc/outside`,
		},
		expError: `Test: File "../../etc/outside" is outside Root: permission denied`,
	}}

	var rexDuration = regexp.MustCompile(`(?m)\s+(\d+\.\d+)s$`)
	var tcase testCase
	for _, tcase = range listCase {
		tcase.treq.Body = string(tdata.Input[tcase.tag])
		tcase.treq.init(playgo.opts)

		var got []byte
		got, err = playgo.Test(&tcase.treq)
		if err != nil {
			test.Assert(t, tcase.tag, tcase.expError, err.Error())
		}
		got = rexDuration.ReplaceAll(got, []byte(" Xs"))

		var exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(got))
	}
}
