// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package http_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	libhttp "kilabit.info/pakakeh.go/lib/http"
)

func ExampleEndpointResponse() {
	type myData struct {
		ID string
	}

	server := libhttp.Server{}

	// Lest say we have an endpoint that echoing back the request
	// parameter "id" back to client inside the EndpointResponse.Data using
	// myData as JSON format.
	// If the parameter "id" is missing or empty it will return an HTTP
	// status code with message as defined in EndpointResponse.
	err := server.RegisterEndpoint(libhttp.Endpoint{
		Method:       libhttp.RequestMethodGet,
		RequestType:  libhttp.RequestTypeQuery,
		ResponseType: libhttp.ResponseTypeJSON,
		Call: func(epr *libhttp.EndpointRequest) ([]byte, error) {
			res := &libhttp.EndpointResponse{}
			id := epr.HTTPRequest.Form.Get(`id`)
			if len(id) == 0 {
				res.E.Code = http.StatusBadRequest
				res.E.Message = `empty parameter id`
				return nil, res
			}
			if id == `0` {
				// If the EndpointResponse.Code is 0, it will
				// default to http.StatusInternalServerError
				res.E.Message = `id value 0 cause internal server error`
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

	// Simulate request using httptest.

	testServe := func(q string) (resp *http.Response, respBody string) {
		req := httptest.NewRequest(`GET`, `/`, nil)
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

	params := url.Values{}

	// Test call endpoint without "id" parameter.

	_, resBody := testServe(params.Encode())
	fmt.Printf("GET / => %s\n", resBody)

	// Test call endpoint with "id" parameter set to "0", it should return
	// HTTP status 500 with custom message.

	params.Set(`id`, `0`)
	_, resBody = testServe(params.Encode())
	fmt.Printf("GET /?id=0 => %s\n", resBody)

	// Test with "id" parameter is set.

	params.Set(`id`, `1000`)
	_, resBody = testServe(params.Encode())
	fmt.Printf("GET /?id=1000 => %s\n", resBody)

	// Output:
	// GET / => {"message":"empty parameter id","code":400}
	// GET /?id=0 => {"message":"id value 0 cause internal server error","code":500}
	// GET /?id=1000 => {"data":{"ID":"1000"},"code":200}
}
