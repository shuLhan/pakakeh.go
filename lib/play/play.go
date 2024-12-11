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
// HTTP APIs for formatting and running Go code accept JSON content type,
// with the following request format,
//
//	{
//		"goversion": <string>, // For run only.
//		"without_race": <boolean>, // For run only.
//		"body": <string>
//	}
//
// The "goversion" field define the Go tools and toolchain version to be
// used to compile the code.
// The default "goversion" is defined as global variable [GoVersion] in this
// package.
// If "without_race" is true, the Run command will not run with "-race"
// option.
// The "body" field contains the Go code to be formatted or run.
//
// Both return the following JSON response format,
//
//	{
//		"code": <integer, HTTP status code>,
//		"name": <string, error type>,
//		"message": <string, optional message>,
//		"data": <string>
//	}
//
// For the [HTTPHandleFormat], the response "data" contains the formatted Go
// code.
// For the [HTTPHandleRun], the response "data" contains the output from
// running the Go code, the "message" contains an error pre-Run, like bad
// request or file system related error.
//
// # Unsafe run
//
// As exceptional, the [Run] and [HTTPHandleRun] accept the following
// request for running program inside custom "go.mod",
//
//	{
//		"unsafe_run": <path>
//	}
//
// The "unsafe_run" define the path to directory relative to HTTP server
// working directory.
// Once request accepted it will change the directory into "unsafe_run" first
// and then run "go run ." directly.
// Go code that executed inside "unsafe_run" should be not modifiable and
// safe from mallicious execution.
//
// # Testing
//
// For testing, since the test must run inside the directory that contains
// the Go file to be tested, the [HTTPHandleTest] API accept the following
// request format,
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
// The "body" field contains the Go code that will be saved to
// "file".
// The test will run, by default, with "go test -count=1 -race $dirname"
// where "$dirname" is the path directory to the "file" relative to where
// the program is running.
// If "without_race" is true, the test command will not run with "-race"
// option.
package play

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/tools/imports"

	liberrors "git.sr.ht/~shulhan/pakakeh.go/lib/errors"
	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
)

// ErrEmptyFile error when running [Test] with empty File field in the
// [Request].
var ErrEmptyFile = errors.New(`empty File`)

// GoVersion define the Go tool version for go.mod to be used to run the
// code.
var GoVersion = `1.23.2`

// Timeout define the maximum time the program can be run until it get
// terminated.
var Timeout = 10 * time.Second

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

// Format the Go code in the [Request.Body] and return the result to out.
// Any syntax error on the code will be returned as error.
func Format(req Request) (out []byte, err error) {
	var logp = `Format`
	var fmtbody []byte

	fmtbody, err = imports.Process(`main.go`, []byte(req.Body), nil)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return fmtbody, nil
}

// HTTPHandleFormat define the HTTP handler for formating Go code.
func HTTPHandleFormat(httpresw http.ResponseWriter, httpreq *http.Request) {
	var (
		logp = `HTTPHandleFormat`
		resp = libhttp.EndpointResponse{}

		req     Request
		rawbody []byte
		err     error
	)

	httpresw.Header().Set(libhttp.HeaderContentType, libhttp.ContentTypeJSON)

	var contentType = httpreq.Header.Get(libhttp.HeaderContentType)
	if contentType != libhttp.ContentTypeJSON {
		resp.Code = http.StatusUnsupportedMediaType
		resp.Name = `ERR_CONTENT_TYPE`
		goto out
	}

	rawbody, err = io.ReadAll(httpreq.Body)
	if err != nil {
		resp.Code = http.StatusInternalServerError
		resp.Name = `ERR_INTERNAL`
		resp.Message = err.Error()
		goto out
	}

	err = json.Unmarshal(rawbody, &req)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Name = `ERR_BAD_REQUEST`
		resp.Message = err.Error()
		goto out
	}

	rawbody, err = Format(req)
	if err != nil {
		resp.Code = http.StatusUnprocessableEntity
		resp.Name = `ERR_CODE`
		resp.Message = err.Error()
		goto out
	}

	resp.Code = http.StatusOK
	resp.Data = string(rawbody)
out:
	rawbody, err = json.Marshal(resp)
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
		resp.Code = http.StatusInternalServerError
	}

	httpresw.WriteHeader(resp.Code)
	httpresw.Write(rawbody)
}

// HTTPHandleRun define the HTTP handler for running Go code.
// Each client is identified by unique cookie, so if two Run requests come
// from the same client, the previous Run will be cancelled.
func HTTPHandleRun(httpresw http.ResponseWriter, httpreq *http.Request) {
	var (
		logp = `HTTPHandleRun`

		req  *Request
		resp *libhttp.EndpointResponse
		rawb []byte
		err  error
	)

	httpresw.Header().Set(libhttp.HeaderContentType, libhttp.ContentTypeJSON)

	req, resp = readRequest(httpreq)
	if resp != nil {
		goto out
	}

	rawb, err = Run(req)
	if err != nil {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: err.Error(),
				Name:    `ERR_INTERNAL`,
				Code:    http.StatusInternalServerError,
			},
		}
		goto out
	}

	http.SetCookie(httpresw, req.cookieSid)
	resp = &libhttp.EndpointResponse{}
	resp.Code = http.StatusOK
	resp.Data = string(rawb)
out:
	rawb, err = json.Marshal(resp)
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
		resp.Code = http.StatusInternalServerError
	}
	httpresw.WriteHeader(resp.Code)
	httpresw.Write(rawb)
}

func readRequest(httpreq *http.Request) (req *Request, resp *libhttp.EndpointResponse) {
	var contentType = httpreq.Header.Get(libhttp.HeaderContentType)
	if contentType != libhttp.ContentTypeJSON {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: `invalid content type`,
				Name:    `ERR_CONTENT_TYPE`,
				Code:    http.StatusUnsupportedMediaType,
			},
		}
		return nil, resp
	}

	var (
		rawbody []byte
		err     error
	)

	rawbody, err = io.ReadAll(httpreq.Body)
	if err != nil {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: err.Error(),
				Name:    `ERR_INTERNAL`,
				Code:    http.StatusInternalServerError,
			},
		}
		return nil, resp
	}

	err = json.Unmarshal(rawbody, &req)
	if err != nil {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: err.Error(),
				Name:    `ERR_BAD_REQUEST`,
				Code:    http.StatusBadRequest,
			},
		}
		return nil, resp
	}

	req.cookieSid, err = httpreq.Cookie(cookieNameSid)
	if err != nil {
		// Ignore the error if cookie is not exist, we wiil generate
		// one later.
	}

	return req, nil
}

// Run the Go code in the [Request.Body].
func Run(req *Request) (out []byte, err error) {
	var logp = `Run`

	req.init()

	var cmd *command = runningCmd.get(req.cookieSid.Value)
	if cmd != nil {
		cmd.ctxCancel()
		runningCmd.delete(req.cookieSid.Value)
	}

	if len(req.UnsafeRun) != 0 {
		return unsafeRun(req)
	}
	if len(req.Body) == 0 {
		return nil, nil
	}

	var tempdir = filepath.Join(userCacheDir, `goplay`, req.cookieSid.Value)

	err = os.MkdirAll(tempdir, 0700)
	if err != nil {
		return nil, fmt.Errorf(`%s: MkdirAll %q: %w`, logp, tempdir, err)
	}

	var gomod = filepath.Join(tempdir, `go.mod`)

	var gomodTemplate = "module play.local\n\ngo " + req.GoVersion + "\n"

	err = os.WriteFile(gomod, []byte(gomodTemplate), 0600)
	if err != nil {
		return nil, fmt.Errorf(`%s: WriteFile %q: %w`, logp, gomod, err)
	}

	var maingo = filepath.Join(tempdir, `main.go`)

	err = os.WriteFile(maingo, []byte(req.Body), 0600)
	if err != nil {
		return nil, fmt.Errorf(`%s: WriteFile %q: %w`, logp, maingo, err)
	}

	cmd = newCommand(req, tempdir)
	runningCmd.store(req.cookieSid.Value, cmd)
	out = cmd.run()

	return out, nil
}

func unsafeRun(req *Request) (out []byte, err error) {
	var cmd = newCommand(req, req.UnsafeRun)
	runningCmd.store(req.cookieSid.Value, cmd)

	out = cmd.run()

	return out, nil
}

// Test the Go code in the [Request.Body].
func Test(req *Request) (out []byte, err error) {
	var logp = `Test`

	req.init()

	var cmd *command = runningCmd.get(req.cookieSid.Value)
	if cmd != nil {
		cmd.ctxCancel()
		runningCmd.delete(req.cookieSid.Value)
	}

	if len(req.File) == 0 {
		return nil, ErrEmptyFile
	}
	if len(req.UnsafeRun) == 0 {
		req.UnsafeRun = filepath.Dir(req.File)
	}

	err = os.WriteFile(req.File, []byte(req.Body), 0600)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	cmd = newTestCommand(req)
	runningCmd.store(req.cookieSid.Value, cmd)
	out = cmd.run()

	return out, nil
}
