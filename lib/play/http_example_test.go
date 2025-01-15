// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
)

func ExampleGo_HTTPHandleFormat() {
	const codeIndentMissingImport = `
package main
func main() {
  fmt.Println("Hello, world")
}
`
	var req = Request{
		Body: codeIndentMissingImport,
	}
	var (
		rawbody []byte
		err     error
	)
	rawbody, err = json.Marshal(&req)
	if err != nil {
		log.Fatal(err)
	}

	var resprec = httptest.NewRecorder()
	var httpreq = httptest.NewRequest(`POST`, `/api/play/format`,
		bytes.NewReader(rawbody))
	httpreq.Header.Set(`Content-Type`, `application/json`)

	var playgo = NewGo(GoOptions{})
	var mux = http.NewServeMux()
	mux.HandleFunc(`POST /api/play/format`, playgo.HTTPHandleFormat)
	mux.ServeHTTP(resprec, httpreq)

	var resp = resprec.Result()
	rawbody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, rawbody)

	// Output:
	// {"data":"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, world\")\n}\n","code":200}
}

func ExampleGo_HTTPHandleRun() {
	const code = `
package main
import "fmt"
func main() {
	fmt.Println("Hello, world")
}
`
	var req = Request{
		Body: code,
	}
	var (
		rawbody []byte
		err     error
	)
	rawbody, err = json.Marshal(&req)
	if err != nil {
		log.Fatal(err)
	}

	var resprec = httptest.NewRecorder()

	var httpreq = httptest.NewRequest(`POST`, `/api/play/run`,
		bytes.NewReader(rawbody))
	httpreq.Header.Set(`Content-Type`, `application/json`)

	var playgo = NewGo(GoOptions{})
	var mux = http.NewServeMux()

	mux.HandleFunc(`POST /api/play/run`, playgo.HTTPHandleRun)
	mux.ServeHTTP(resprec, httpreq)

	var resp = resprec.Result()
	rawbody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, rawbody)

	// Output:
	// {"data":"Hello, world\n","code":200}
}

func ExampleGo_HTTPHandleTest() {
	const code = `
package test
import "testing"
func TestSum(t *testing.T) {
	var total = sum(1, 2, 3)
	if total != 6 {
		t.Fatalf("got %d, want 6", total)
	}
}`
	var req = Request{
		Body: code,
		File: `testdata/test_test.go`,
	}
	var (
		rawbody []byte
		err     error
	)
	rawbody, err = json.Marshal(&req)
	if err != nil {
		log.Fatal(err)
	}

	var playgo = NewGo(GoOptions{})
	var mux = http.NewServeMux()

	mux.HandleFunc(`POST /api/play/test`, playgo.HTTPHandleTest)

	var resprec = httptest.NewRecorder()
	var httpreq = httptest.NewRequest(`POST`, `/api/play/test`,
		bytes.NewReader(rawbody))
	httpreq.Header.Set(`Content-Type`, `application/json`)

	mux.ServeHTTP(resprec, httpreq)
	var resp = resprec.Result()

	rawbody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var rexDuration = regexp.MustCompile(`(?m)\\t(\d+\.\d+)s`)
	rawbody = rexDuration.ReplaceAll(rawbody, []byte(`\tXs`))

	fmt.Printf(`%s`, rawbody)

	// Output:
	// {"data":"ok  \tgit.sr.ht/~shulhan/pakakeh.go/lib/play/testdata\tXs\n","code":200}
}
