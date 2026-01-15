// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Bencmark route.Parse before replacing it.
//
// Result:
//
// $ benchstat bench_handleDelete_before.txt bench_handleDelete_after.txt
// goos: linux
// goarch: amd64
// pkg: git.sr.ht/~shulhan/pakakeh.go/lib/http
// cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
//
//	│ bench_handleDelete_before.txt │ bench_handleDelete_after.txt  │
//	│            sec/op             │   sec/op     vs base          │
//
// Server_handleDelete-4                     1.703µ ± 1%   1.702µ ± 2%  ~ (p=0.956 n=10)
//
//	│ bench_handleDelete_before.txt │   bench_handleDelete_after.txt   │
//	│             B/op              │     B/op      vs base            │
//
// Server_handleDelete-4                    1.125Ki ± 0%   1.125Ki ± 0%  ~ (p=1.000 n=10) ¹
// ¹ all samples are equal
//
//	│ bench_handleDelete_before.txt │  bench_handleDelete_after.txt  │
//	│           allocs/op           │ allocs/op   vs base            │
//
// Server_handleDelete-4                      9.000 ± 0%   9.000 ± 0%  ~ (p=1.000 n=10) ¹
// ¹ all samples are equal
func BenchmarkServer_handleDelete(b *testing.B) {
	var (
		srv = &Server{}
		err error
	)

	err = srv.RegisterEndpoint(Endpoint{
		Method: RequestMethodDelete,
		Path:   `/a/b/c/:d/e`,
		Call: func(_ *EndpointRequest) ([]byte, error) {
			return nil, nil
		},
	})
	if err != nil {
		b.Fatal(err)
	}

	var (
		httpWriter      = httptest.NewRecorder()
		httpReqMatched  *http.Request
		httpReqNotFound *http.Request
		body            bytes.Buffer
	)

	httpReqMatched, err = http.NewRequestWithContext(context.Background(),
		http.MethodDelete, `/a/b/c/dddd/e`, &body)
	if err != nil {
		b.Fatal(err)
	}

	httpReqNotFound, err = http.NewRequestWithContext(context.Background(),
		http.MethodDelete, `/a/b/c/dddd`, &body)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	var x int
	for ; x < b.N; x++ {
		srv.handleDelete(httpWriter, httpReqMatched)
		srv.handleDelete(httpWriter, httpReqNotFound)
	}
}

// Result:
//
// goos: linux
// goarch: amd64
// pkg: git.sr.ht/~shulhan/pakakeh.go/lib/http
// cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
//
//	│ bench_registerDelete_before.txt │   bench_registerDelete_after.txt    │
//	│             sec/op              │   sec/op     vs base                │
//
// Server_registerDelete-4                       2.411µ ± 2%   4.232µ ± 1%  +75.53% (p=0.000 n=10)
//
//	│ bench_registerDelete_before.txt │   bench_registerDelete_after.txt   │
//	│              B/op               │    B/op     vs base                │
//
// Server_registerDelete-4                        660.0 ± 0%   796.0 ± 0%  +20.61% (p=0.000 n=10)
//
//	│ bench_registerDelete_before.txt │    bench_registerDelete_after.txt    │
//	│            allocs/op            │  allocs/op   vs base                 │
//
// Server_registerDelete-4                        8.000 ± 0%   19.000 ± 0%  +137.50% (p=0.000 n=10)
//
// Summary: creating new route and matching the nodes result in increase
// ops and memory usage.
func BenchmarkServer_registerDelete(b *testing.B) {
	var (
		srv *Server
		err error
	)

	srv, err = NewServer(ServerOptions{})
	if err != nil {
		b.Fatal(err)
	}

	var ep = Endpoint{
		Method: RequestMethodDelete,
		Path:   `/a/b/c/:d/e`,
		Call: func(_ *EndpointRequest) ([]byte, error) {
			return nil, nil
		},
	}

	b.ResetTimer()

	var x int
	for ; x < b.N; x++ {
		_ = srv.RegisterEndpoint(ep)
	}
}
