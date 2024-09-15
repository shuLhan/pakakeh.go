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
)

func ExampleFormat() {
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
		out []byte
		err error
	)
	out, err = Format(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)

	//Output:
	//package main
	//
	//import "fmt"
	//
	//func main() {
	//	fmt.Println("Hello, world")
	//}
}

func ExampleHTTPHandleFormat() {
	var mux = http.NewServeMux()
	mux.HandleFunc(`POST /api/play/format`, HTTPHandleFormat)

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
	var httpreq = httptest.NewRequest(`POST`, `/api/play/format`, bytes.NewReader(rawbody))
	httpreq.Header.Set(`Content-Type`, `application/json`)

	mux.ServeHTTP(resprec, httpreq)
	var resp = resprec.Result()

	rawbody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, rawbody)

	//Output:
	//{"data":"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, world\")\n}\n","code":200}
}

func ExampleHTTPHandleRun() {
	var mux = http.NewServeMux()
	mux.HandleFunc(`POST /api/play/run`, HTTPHandleRun)

	const codeRun = `
package main
import "fmt"
func main() {
	fmt.Println("Hello, world")
}
`
	var req = Request{
		Body: codeRun,
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
	var httpreq = httptest.NewRequest(`POST`, `/api/play/run`, bytes.NewReader(rawbody))
	httpreq.Header.Set(`Content-Type`, `application/json`)

	mux.ServeHTTP(resprec, httpreq)
	var resp = resprec.Result()

	rawbody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, rawbody)

	//Output:
	//{"data":"Hello, world\n","code":200}
}

func ExampleRun() {
	const codeRun = `
package main
import "fmt"
func main() {
	fmt.Println("Hello, world")
}`

	var req = Request{
		Body: codeRun,
	}
	var (
		out []byte
		err error
	)
	out, err = Run(&req)
	if err != nil {
		fmt.Printf(`error: %s`, err)
	}
	fmt.Printf(`%s`, out)

	//Output:
	//Hello, world
}
