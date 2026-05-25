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

	libhttp "kilabit.info/pakakeh.go/lib/http"
)

func ExampleEndpoint_requestType() {
	server := libhttp.Server{}

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
	err := server.RegisterEndpoint(epForm)
	if err != nil {
		log.Fatal(err)
	}

	testServe := func(contentType, body string) (resp *http.Response, respBody string) {
		reqBody := io.NopCloser(bytes.NewBufferString(body))
		req := httptest.NewRequest(`POST`, `/myendpoint`, reqBody)
		req.Header.Set(libhttp.HeaderContentType, contentType)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
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

// Example on using custom [Endpoint.ErrorHandler].
func ExampleEndpoint_errorHandler() {
	server := &libhttp.Server{}

	// This endpoint have custom [Endpoint.ErrorHandler] that read the
	// error code and message from query parameters.

	endpointError := libhttp.Endpoint{
		Method:       libhttp.RequestMethodGet,
		Path:         "/",
		RequestType:  libhttp.RequestTypeQuery,
		ResponseType: libhttp.ResponseTypePlain,
		Call: func(epr *libhttp.EndpointRequest) ([]byte, error) {
			// Returning an error will call the ErrorHandler.
			return nil, errors.New(epr.HTTPRequest.Form.Get(`error`))
		},
		ErrorHandler: func(epr *libhttp.EndpointRequest) {
			codeMsg := strings.Split(epr.Error.Error(), ":")
			var statusCode int
			var body string
			if len(codeMsg) != 2 {
				statusCode = http.StatusInternalServerError
				body = epr.Error.Error()
			} else {
				statusCode, _ = strconv.Atoi(codeMsg[0])
				body = codeMsg[1]
			}
			epr.HTTPWriter.WriteHeader(statusCode)
			_, _ = epr.HTTPWriter.Write([]byte(body))
		},
	}
	_ = server.RegisterEndpoint(endpointError)

	// Simulate request using httptest.

	testServe := func(q string) (resp *http.Response, respBody string) {
		req := httptest.NewRequest(`GET`, endpointError.Path, nil)
		req.URL.RawQuery = q
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp = w.Result()

		respb, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return resp, string(respb)
	}

	var params = url.Values{}
	params.Set(`error`, `400:error with status code`)
	resp, resBody := testServe(params.Encode())
	fmt.Printf("%d: %s\n", resp.StatusCode, resBody)

	// Set the error message without status code.

	params.Set(`error`, `error without status code`)
	resp, resBody = testServe(params.Encode())
	fmt.Printf("%d: %s\n", resp.StatusCode, resBody)

	// Output:
	// 400: error with status code
	// 500: error without status code
}
