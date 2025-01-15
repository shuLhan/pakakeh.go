// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package play provides callable APIs and HTTP handlers to format, run, and
// test Go code, similar to Go playground but using HTTP instead of
// WebSocket.
//
// For HTTP API, this package expose handlers: [HTTPHandleFormat],
// [HTTPHandleRun], and [HTTPHandleTest].
//
// # Formatting and running Go code
//
// The HTTP APIs for formatting and running Go code accept the following JSON
// request format,
//
//	{
//		"goversion": <string>, // For run only.
//		"without_race": <boolean>, // For run only.
//		"body": <string>
//	}
//
// The "goversion" field define the Go tools and toolchain version to be
// used to compile the code.
// The default "goversion" is defined in global variable [GoVersion] in this
// package.
// If "without_race" is true, the Run command will not run with "-race"
// option.
// The "body" field contains the Go code to be formatted or run.
//
// Both request return the following JSON response format,
//
//	{
//		"code": <integer, HTTP status code>,
//		"name": <string, error type>,
//		"message": <string, optional message>,
//		"data": <string>
//	}
//
// For the [Go.HTTPHandleFormat], the response "data" contains the formatted Go
// code.
// For the [Go.HTTPHandleRun], the response "data" contains the output from
// running the Go code.
// The "message" field contains an error on pre-Run, like bad request or file
// system related error.
//
// # Unsafe run
//
// As exceptional, the [Go.Run] and [Go.HTTPHandleRun] accept the following
// request for running program inside custom "go.mod",
//
//	{
//		"unsafe_run": <path>
//	}
//
// The "unsafe_run" define the path to directory relative to
// [play.GoOptions.Root] working directory.
// Once the request is accepted it will change the directory into "unsafe_run"
// first and then execute "go run ." directly.
// Go code that executed inside "unsafe_run" should be not modifiable and
// safe from mallicious execution.
//
// # Testing
//
// The [Go.Test] or [Go.HTTPHandleTest] must run inside the directory that
// contains the Go code file to be tested.
// The [Go.HTTPHandleTest] API accept the following request format,
//
//	{
//		"goversion": <string>,
//		"file": <string>,
//		"body": <string>,
//		"without_race": <boolean>
//	}
//
// The "file" field define the path to the "_test.go" file, default to
// "test_test.go" if its empty.
// The "body" field contains the Go code that will be saved to that "file".
// The test will run, by default, with "go test -count=1 -race $dirname"
// where "$dirname" is the path to directory where "file" located, must be
// under the [play.GoOptions.Root] directory.
// If "without_race" is true, the test command will not run with "-race"
// option.
package play

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/imports"
)

// ErrEmptyFile error when running [Go.Test] with empty [play.Request.File].
var ErrEmptyFile = errors.New(`empty File`)

// GoVersion define the Go tool version for go.mod to be used to run the
// code.
const GoVersion = `1.23.2`

// Timeout define the maximum time the program can be run until it get
// terminated.
const Timeout = 10 * time.Second

var now = func() int64 {
	return time.Now().Unix()
}

// runningCmd contains list of running Go code with [Request.SID] as the
// key.
var runningCmd = runManager{
	sidCommand: make(map[string]*command),
}

var userHomeDir string
var userCacheDir string

func init() {
	var err error
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		userHomeDir = os.TempDir()
	}

	userCacheDir, err = os.UserCacheDir()
	if err != nil {
		userCacheDir = os.TempDir()
	}
}

// Format the Go code in the [play.Request.Body] and return the result to out.
// Any syntax error on the code will be returned as error.
func (playgo *Go) Format(req Request) (out []byte, err error) {
	var logp = `Format`
	var fmtbody []byte

	fmtbody, err = imports.Process(`main.go`, []byte(req.Body), nil)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return fmtbody, nil
}

// Run the Go code in the [play.Request.Body].
func (playgo *Go) Run(req *Request) (out []byte, err error) {
	var logp = `Run`

	req.init(playgo.opts)

	var cmd *command = runningCmd.get(req.cookieSid.Value)
	if cmd != nil {
		cmd.ctxCancel()
		runningCmd.delete(req.cookieSid.Value)
	}

	if len(req.UnsafeRun) == 0 {
		if len(req.Body) == 0 {
			return nil, nil
		}
		err = req.writes(playgo.opts)
	} else {
		err = req.initUnsafe(playgo.opts)
	}
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	cmd = newCommand(playgo.opts, req)
	runningCmd.store(req.cookieSid.Value, cmd)
	out = cmd.run()

	return out, nil
}

// Test the Go code in the [play.Request.Body].
func (playgo *Go) Test(req *Request) (out []byte, err error) {
	var logp = `Test`

	req.init(playgo.opts)

	var cmd *command = runningCmd.get(req.cookieSid.Value)
	if cmd != nil {
		cmd.ctxCancel()
		runningCmd.delete(req.cookieSid.Value)
	}

	if len(req.File) == 0 {
		return nil, ErrEmptyFile
	}

	var absPathFile = filepath.Join(playgo.opts.absRoot, req.File)
	if !strings.HasPrefix(absPathFile, playgo.opts.absRoot) {
		return nil, fmt.Errorf(`%s: File %q is outside Root: %w`,
			logp, req.File, os.ErrPermission)
	}

	if len(req.UnsafeRun) == 0 {
		req.UnsafeRun = filepath.Dir(absPathFile)
	} else {
		err = req.initUnsafe(playgo.opts)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	err = os.WriteFile(absPathFile, []byte(req.Body), 0600)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	cmd = newTestCommand(playgo.opts, req)
	runningCmd.store(req.cookieSid.Value, cmd)
	out = cmd.run()

	return out, nil
}
