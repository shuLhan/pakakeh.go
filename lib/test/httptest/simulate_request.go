// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

package httptest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
)

// SimulateRequest request to simulate [http.ServeHTTP].
type SimulateRequest struct {
	Method string
	Path   string

	// JSONIndentResponse if not empty, the response body will be
	// indented using [json.Indent].
	JSONIndentResponse string

	Header http.Header
	Body   []byte
}

func (simreq *SimulateRequest) toHTTPRequest() (req *http.Request) {
	var body bytes.Buffer

	body.Write(simreq.Body)

	req = httptest.NewRequest(simreq.Method, simreq.Path, &body)

	var (
		key  string
		val  string
		vals []string
	)
	for key, vals = range simreq.Header {
		for _, val = range vals {
			req.Header.Add(key, val)
		}
	}
	return req
}
