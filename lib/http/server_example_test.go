// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
