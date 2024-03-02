// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func ExampleEndpointResponse() {
	type myData struct {
		ID string
	}

	server, err := NewServer(&ServerOptions{
		Address: "127.0.0.1:7016",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Lest say we have an endpoint that echoing back the request
	// parameter "id" back to client inside the EndpointResponse.Data using
	// myData as JSON format.
	// If the parameter "id" is missing or empty it will return an HTTP
	// status code with message as defined in EndpointResponse.
	err = server.RegisterEndpoint(&Endpoint{
		Method:       RequestMethodGet,
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypeJSON,
		Call: func(epr *EndpointRequest) ([]byte, error) {
			res := &EndpointResponse{}
			id := epr.HTTPRequest.Form.Get(`id`)
			if len(id) == 0 {
				res.E.Code = http.StatusBadRequest
				res.E.Message = "empty parameter id"
				return nil, res
			}
			if id == "0" {
				// If the EndpointResponse.Code is 0, it will
				// default to http.StatusInternalServerError
				res.E.Message = "id value 0 cause internal server error"
				return nil, res
			}
			res.E.Code = http.StatusOK
			res.Data = &myData{
				ID: id,
			}
			return json.Marshal(res)
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var errStart = server.Start()
		if errStart != nil {
			log.Fatal(errStart)
		}
	}()
	time.Sleep(1 * time.Second)

	clientOpts := &ClientOptions{
		ServerURL: `http://127.0.0.1:7016`,
	}
	cl := NewClient(clientOpts)
	params := url.Values{}

	// Test call endpoint without "id" parameter.
	_, resBody, err := cl.Get("/", nil, params)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GET / => %s\n", resBody)

	// Test call endpoint with "id" parameter set to "0", it should return
	// HTTP status 500 with custom message.
	params.Set("id", "0")
	_, resBody, err = cl.Get("/", nil, params)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GET /?id=0 => %s\n", resBody)

	// Test with "id" parameter is set.
	params.Set("id", "1000")
	_, resBody, err = cl.Get("/", nil, params)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GET /?id=1000 => %s\n", resBody)

	// Output:
	// GET / => {"message":"empty parameter id","code":400}
	// GET /?id=0 => {"message":"id value 0 cause internal server error","code":500}
	// GET /?id=1000 => {"data":{"ID":"1000"},"code":200}
}
