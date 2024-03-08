// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httptest implement testing HTTP package.
package httptest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

// Simulate HTTP server handler by generating [http.Request] from
// fields in [SimulateRequest]; and then call [http.HandlerFunc].
// The HTTP response from serve along with its raw body and original HTTP
// request then returned in [*SimulateResult].
func Simulate(serve http.HandlerFunc, req *SimulateRequest) (result *SimulateResult, err error) {
	var logp = `Simulate`

	result = &SimulateResult{
		Request: req.toHTTPRequest(),
	}

	var httpWriter = httptest.NewRecorder()

	serve(httpWriter, result.Request)

	result.Response = httpWriter.Result() //nolint:bodyclose

	result.ResponseBody, err = io.ReadAll(result.Response.Body)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	err = result.Response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	if len(req.JSONIndentResponse) != 0 {
		var dst bytes.Buffer
		err = json.Indent(&dst, result.ResponseBody, ``, req.JSONIndentResponse)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		result.ResponseBody = dst.Bytes()
	}

	// Recreate the bodies to prevent panic when user trying to inspect
	// request or response body.

	result.Request.Body = io.NopCloser(bytes.NewReader(req.Body))
	result.Response.Body = io.NopCloser(bytes.NewReader(result.ResponseBody))

	return result, nil
}
