// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

func ExampleServer_customHTTPStatusCode() {
	type CustomResponse struct {
		Status int `json:"status"`
	}

	var (
		exp = CustomResponse{
			Status: http.StatusBadRequest,
		}
		opts = ServerOptions{
			Address: "127.0.0.1:8123",
		}

		srv *Server
		err error
	)

	srv, err = NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = srv.Start()
		if err != nil {
			log.Println(err)
		}
	}()

	defer func() {
		_ = srv.Stop(5 * time.Second)
	}()

	epCustom := &Endpoint{
		Path:         "/error/custom",
		RequestType:  RequestTypeJSON,
		ResponseType: ResponseTypeJSON,
		Call: func(epr *EndpointRequest) (
			resbody []byte, err error,
		) {
			epr.HTTPWriter.WriteHeader(exp.Status)
			return json.Marshal(exp)
		},
	}

	err = srv.registerPost(epCustom)
	if err != nil {
		log.Println(err)
		return
	}

	// Wait for the server fully started.
	time.Sleep(1 * time.Second)

	var (
		clientOpts = ClientOptions{
			ServerURL: `http://127.0.0.1:8123`,
		}
		client = NewClient(clientOpts)
		req    = ClientRequest{
			Path: epCustom.Path,
		}

		res *ClientResponse
	)

	res, err = client.PostJSON(req)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%d\n", res.HTTPResponse.StatusCode)
	fmt.Printf("%s\n", res.Body)

	// Output:
	// 400
	// {"status":400}
}

func ExampleServer_RegisterHandleFunc() {
	var serverOpts = ServerOptions{}
	server, _ := NewServer(serverOpts)
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
