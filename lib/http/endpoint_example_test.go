// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ExampleEndpoint_errorHandler() {
	serverOpts := &ServerOptions{
		Address: "127.0.0.1:8123",
	}
	server, _ := NewServer(serverOpts)

	endpointError := &Endpoint{
		Method:       RequestMethodGet,
		Path:         "/",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call: func(epr *EndpointRequest) ([]byte, error) {
			return nil, errors.New(epr.HTTPRequest.Form.Get(`error`))
		},
		ErrorHandler: func(epr *EndpointRequest) {
			epr.HTTPWriter.Header().Set(HeaderContentType, ContentTypePlain)

			codeMsg := strings.Split(epr.Error.Error(), ":")
			if len(codeMsg) != 2 {
				epr.HTTPWriter.WriteHeader(http.StatusInternalServerError)
				_, _ = epr.HTTPWriter.Write([]byte(epr.Error.Error()))
			} else {
				code, _ := strconv.Atoi(codeMsg[0])
				epr.HTTPWriter.WriteHeader(code)
				_, _ = epr.HTTPWriter.Write([]byte(codeMsg[1]))
			}
		},
	}
	_ = server.RegisterEndpoint(endpointError)

	go func() {
		_ = server.Start()
	}()
	defer func() {
		_ = server.Stop(1 * time.Second)
	}()
	time.Sleep(1 * time.Second)

	clientOpts := &ClientOptions{
		ServerURL: `http://` + serverOpts.Address,
	}
	client := NewClient(clientOpts)

	params := url.Values{}
	params.Set("error", "400:error with status code")
	httpres, resbody, err := client.Get(`/`, nil, params) //nolint: bodyclose
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%d: %s\n", httpres.StatusCode, resbody)

	params.Set("error", "error without status code")
	httpres, resbody, err = client.Get(`/`, nil, params) //nolint: bodyclose
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%d: %s\n", httpres.StatusCode, resbody)

	// Output:
	// 400: error with status code
	// 500: error without status code
}
