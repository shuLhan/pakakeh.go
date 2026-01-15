// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

package httptest

import (
	"bytes"
	"fmt"
	"maps"
	"net/http"
	"net/http/httputil"
)

// SimulateResult contains [http.Request] and [http.Response] from calling
// [http.ServeHTTP].
type SimulateResult struct {
	Request  *http.Request  `json:"-"`
	Response *http.Response `json:"-"`

	RequestDump  []byte
	ResponseDump []byte
	ResponseBody []byte
}

// DumpRequest convert [SimulateResult.Request] with its body to stream of
// bytes using [httputil.DumpRequest].
//
// The returned bytes have CRLF ("\r\n") replaced with single LF ("\n").
//
// Any request headers that match with excludeHeaders will be deleted before
// dumping.
func (result *SimulateResult) DumpRequest(excludeHeaders []string) (raw []byte, err error) {
	if result.RequestDump != nil {
		return result.RequestDump, nil
	}

	var (
		logp      = `DumpRequest`
		orgHeader = maps.Clone(result.Request.Header)
		header    string
	)

	for _, header = range excludeHeaders {
		result.Request.Header.Del(header)
	}

	raw, err = httputil.DumpRequest(result.Request, true)
	result.Request.Header = orgHeader
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	result.RequestDump = bytes.ReplaceAll(raw, []byte{'\r', '\n'}, []byte{'\n'})

	return result.RequestDump, nil
}

// DumpResponse convert [SimulateResult.Response] with its body to stream of
// bytes using [httputil.DumpResponse].
//
// The returned bytes have CRLF ("\r\n") replaced with single LF ("\n").
//
// Any response headers that match with excludeHeaders will be deleted
// before dumping.
func (result *SimulateResult) DumpResponse(excludeHeaders []string) (raw []byte, err error) {
	if result.ResponseDump != nil {
		return result.ResponseDump, nil
	}

	var (
		logp      = `DumpResponse`
		orgHeader = maps.Clone(result.Response.Header)
		header    string
	)
	for _, header = range excludeHeaders {
		result.Response.Header.Del(header)
	}

	raw, err = httputil.DumpResponse(result.Response, true)
	result.Response.Header = orgHeader
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	result.ResponseDump = bytes.ReplaceAll(raw, []byte{'\r', '\n'}, []byte{'\n'})

	return result.ResponseDump, nil
}
