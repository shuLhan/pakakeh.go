package httptest_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	libhttptest "git.sr.ht/~shulhan/pakakeh.go/lib/test/httptest"
)

func ExampleSimulateResult_DumpRequest() {
	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequest(http.MethodGet, `/a/b/c`, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set(`h1`, `v1`)
	req.Header.Set(`h2`, `v2`)
	req.Header.Set(`h3`, `v3`)

	var sim = libhttptest.SimulateResult{
		Request: req,
	}

	var got []byte

	got, err = sim.DumpRequest([]string{`h1`})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("DumpRequest:\n%s", got)

	sim.RequestDump = nil

	got, err = sim.DumpRequest([]string{`h3`})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("DumpRequest:\n%s", got)

	// Output:
	// DumpRequest:
	// GET /a/b/c HTTP/1.1
	// H2: v2
	// H3: v3
	//
	//
	// DumpRequest:
	// GET /a/b/c HTTP/1.1
	// H1: v1
	// H2: v2
}

func ExampleSimulateResult_DumpResponse() {
	var handler = func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(`h1`, `v1`)
		w.Header().Set(`h2`, `v2`)
		w.Header().Set(`h3`, `v3`)
		_, _ = io.WriteString(w, `Hello world!`)
	}

	var (
		ctx = context.Background()

		req *http.Request
		err error
	)

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, `/a/b/c`, nil)
	if err != nil {
		log.Fatal(err)
	}

	var recorder = httptest.NewRecorder()

	handler(recorder, req)

	var result = libhttptest.SimulateResult{
		Response: recorder.Result(),
	}

	var got []byte

	got, err = result.DumpResponse([]string{`h1`})
	if err != nil {
		log.Fatal(`DumpResponse #1:`, err)
	}
	fmt.Printf("<<< DumpResponse:\n%s\n", got)

	result.ResponseDump = nil

	got, err = result.DumpResponse([]string{`h3`})
	if err != nil {
		log.Fatal(`DumpResponse #2:`, err)
	}
	fmt.Printf("<<< DumpResponse:\n%s", got)

	// Output:
	// <<< DumpResponse:
	// HTTP/1.1 200 OK
	// Connection: close
	// Content-Type: text/plain; charset=utf-8
	// H2: v2
	// H3: v3
	//
	// Hello world!
	// <<< DumpResponse:
	// HTTP/1.1 200 OK
	// Connection: close
	// Content-Type: text/plain; charset=utf-8
	// H1: v1
	// H2: v2
	//
	// Hello world!
}
