package httptest_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test/httptest"
)

func ExampleSimulate() {
	http.HandleFunc(`/a/b/c`, func(w http.ResponseWriter, req *http.Request) {
		_ = req.ParseForm()
		var (
			rawjson []byte
			err     error
		)
		rawjson, err = json.Marshal(req.PostForm)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set(`Content-Type`, `application/json`)
		_, _ = w.Write(rawjson)
	})

	var simreq = &httptest.SimulateRequest{
		Method: http.MethodPost,
		Path:   `/a/b/c`,
		Header: http.Header{
			`Content-Type`: []string{`application/x-www-form-urlencoded`},
		},
		Body:               []byte(`id=1&name=go`),
		JSONIndentResponse: `  `,
	}

	var (
		result *httptest.SimulateResult
		err    error
	)

	result, err = httptest.Simulate(http.DefaultServeMux.ServeHTTP, simreq)
	if err != nil {
		log.Fatal(err)
	}

	var dump []byte

	dump, err = result.DumpRequest(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("<<< RequestDump:\n%s\n\n", dump)

	dump, err = result.DumpResponse(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("<<< ResponseDump:\n%s\n\n", dump)
	fmt.Printf("<<< ResponseBody:\n%s\n", result.ResponseBody)

	// Output:
	// <<< RequestDump:
	// POST /a/b/c HTTP/1.1
	// Host: example.com
	// Content-Type: application/x-www-form-urlencoded
	//
	// id=1&name=go
	//
	// <<< ResponseDump:
	// HTTP/1.1 200 OK
	// Connection: close
	// Content-Type: application/json
	//
	// {
	//   "id": [
	//     "1"
	//   ],
	//   "name": [
	//     "go"
	//   ]
	// }
	//
	// <<< ResponseBody:
	// {
	//   "id": [
	//     "1"
	//   ],
	//   "name": [
	//     "go"
	//   ]
	// }
}
