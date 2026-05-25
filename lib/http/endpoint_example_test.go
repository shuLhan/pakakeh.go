// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package http_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"time"

	libhttp "kilabit.info/pakakeh.go/lib/http"
)

func ExampleEndpoint_requestType() {
	srv := libhttp.Server{}

	// Here we defined an endpoint that accept form with "message"
	// parameter
	//
	//	POST /myendpoint
	//	Content-Type: application/x-www-form-urlencoded
	//
	//	message=<string>
	//
	// This endpoint will echo the message in JSON format.
	// If we send "message=hello" it will return,
	//
	//	HTTP/1.1 200 OK
	//	Content-Type: application/json
	//
	//	{"message":"hello"}

	epForm := libhttp.Endpoint{
		Method:       `POST`,
		Path:         `/myendpoint`,
		RequestType:  libhttp.RequestTypeForm,
		ResponseType: libhttp.ResponseTypeJSON,
		Call: func(epr *libhttp.EndpointRequest) ([]byte, error) {
			gotMessage := epr.HTTPRequest.Form.Get(`message`)
			resp := fmt.Sprintf(`{"message":%q}`, gotMessage)
			return []byte(resp), nil
		},
	}
	err := srv.RegisterEndpoint(epForm)
	if err != nil {
		log.Fatal(err)
	}

	testServe := func(contentType, body string) (resp *http.Response, respBody string) {
		reqBody := io.NopCloser(bytes.NewBufferString(body))
		req := httptest.NewRequest(`POST`, `/myendpoint`, reqBody)
		req.Header.Set(libhttp.HeaderContentType, contentType)
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)
		resp = w.Result()

		respb, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return resp, string(respb)
	}

	// Sending request with content-type other than
	// "application/x-www-form-urlencoded" will return an error.

	resp, respBody := testServe(libhttp.ContentTypeJSON, ``)

	fmt.Printf("With Content-Type JSON: %s\n", resp.Status)
	fmt.Printf("Body: %s\n", respBody)

	// With valid content-type and message in body.

	resp, respBody = testServe(libhttp.ContentTypeForm, `message=hello`)

	fmt.Printf("With valid Content-Type: %s\n", resp.Status)
	fmt.Printf("Body: %s\n", respBody)

	// Output:
	// With Content-Type JSON: 400 Bad Request
	// Body: {"message":"invalid content type","code":400}
	// With valid Content-Type: 200 OK
	// Body: {"message":"hello"}
}

func ExampleEndpoint_errorHandler() {
	var (
		serverOpts = libhttp.ServerOptions{
			Address: `127.0.0.1:8123`,
		}
		server *libhttp.Server
	)

	server, _ = libhttp.NewServer(serverOpts)

	var endpointError = libhttp.Endpoint{
		Method:       libhttp.RequestMethodGet,
		Path:         "/",
		RequestType:  libhttp.RequestTypeQuery,
		ResponseType: libhttp.ResponseTypePlain,
		Call: func(epr *libhttp.EndpointRequest) ([]byte, error) {
			return nil, errors.New(epr.HTTPRequest.Form.Get(`error`))
		},
		ErrorHandler: func(epr *libhttp.EndpointRequest) {
			epr.HTTPWriter.Header().Set(libhttp.HeaderContentType,
				libhttp.ContentTypePlain)

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

	var (
		clientOpts = libhttp.ClientOptions{
			ServerURL: `http://` + serverOpts.Address,
		}
		client = libhttp.NewClient(clientOpts)
	)

	var params = url.Values{}
	params.Set("error", "400:error with status code")

	var req = libhttp.ClientRequest{
		Path:   `/`,
		Params: params,
	}

	var (
		res *libhttp.ClientResponse
		err error
	)

	res, err = client.Get(req)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%d: %s\n", res.HTTPResponse.StatusCode, res.Body)

	params.Set("error", "error without status code")

	res, err = client.Get(req)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%d: %s\n", res.HTTPResponse.StatusCode, res.Body)

	// Output:
	// 400: error with status code
	// 500: error without status code
}
