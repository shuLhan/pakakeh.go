// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	libhttp "kilabit.info/pakakeh.go/lib/http"
)

// Example of writing the HTTP status code manully from [Endpoint.Call].
func ExampleServer_customHTTPStatusCode() {
	type CustomResponse struct {
		Message string `json:"message"`
	}

	epCustom := libhttp.Endpoint{
		Method:       `POST`,
		Path:         `/error/custom`,
		RequestType:  libhttp.RequestTypeJSON,
		ResponseType: libhttp.ResponseTypeJSON,
		Call: func(epr *libhttp.EndpointRequest) (resbody []byte, err error) {
			exp := CustomResponse{
				Message: `you have bad request`,
			}
			epr.HTTPWriter.WriteHeader(http.StatusBadRequest)
			return json.Marshal(exp)
		},
	}

	srv := libhttp.Server{}
	err := srv.RegisterEndpoint(epCustom)
	if err != nil {
		log.Fatal(err)
	}

	// Simulate request using httptest.

	req := httptest.NewRequest(`POST`, epCustom.Path, nil)
	req.Header.Set(libhttp.HeaderContentType, libhttp.ContentTypeJSON)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)
	resp := w.Result()

	respb, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d\n", resp.StatusCode)
	fmt.Printf("%s\n", respb)

	// Output:
	// 400
	// {"message":"you have bad request"}
}

func ExampleServer_RegisterHandleFunc() {
	server := libhttp.Server{}
	server.RegisterHandleFunc(`PUT /api/book/:id`,
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			fmt.Fprintf(w, "Request.URL: %s\n", r.URL)
			fmt.Fprintf(w, "Request.Form: %+v\n", r.Form)
			fmt.Fprintf(w, "Request.PostForm: %+v\n", r.PostForm)
		},
	)

	var respRec = httptest.NewRecorder()

	var body = []byte(`title=BahasaPemrogramanGo&author=Shulhan`)
	var req = httptest.NewRequest(`PUT`, `/api/book/123`, bytes.NewReader(body))
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)

	server.ServeHTTP(respRec, req)

	var resp = respRec.Result()

	body, _ = io.ReadAll(resp.Body)
	fmt.Println(resp.Status)
	fmt.Printf("%s", body)
	// Output:
	// 200 OK
	// Request.URL: /api/book/123
	// Request.Form: map[id:[123]]
	// Request.PostForm: map[author:[Shulhan] title:[BahasaPemrogramanGo]]
}
