// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ClientRequest define the parameters for each Client methods.
type ClientRequest struct {
	// Header additional header to be send on request.
	// This field is optional.
	Header http.Header

	//
	// Params define parameter to be send on request.
	// This field is optional.
	//
	// For Method GET, CONNECT, DELETE, HEAD, OPTIONS, or TRACE; the
	// params value should be nil or url.Values.
	// If its url.Values, then the params will be encoded as query
	// parameters.
	//
	// For Method PATCH, POST, or PUT; the Params will converted based on
	// Type rules below,
	//
	// * If Type is RequestTypeQuery and Params is url.Values it will be
	// added as query parameters in the Path.
	//
	// * If Type is RequestTypeForm and Params is url.Values it will be
	// added as URL encoded in the body.
	//
	// * If Type is RequestTypeMultipartForm and Params is
	// map[string][]byte, then it will be converted as multipart form in
	// the body.
	//
	// * If Type is RequestTypeJSON and Params is not nil, the params will
	// be encoded as JSON in body using json.Encode().
	//
	Params any

	body io.Reader

	// The Path to resource on the server.
	// This field is required, if its empty default to "/".
	Path string

	contentType string

	// The HTTP method of request.
	// This field is optional, if its empty default to RequestMethodGet
	// (GET).
	Method RequestMethod

	// The Type of request.
	// This field is optional, it's affect how the Params field encoded in
	// the path or body.
	Type RequestType
}

// toHTTPRequest convert the ClientRequest into the standard http.Request.
func (creq *ClientRequest) toHTTPRequest(client *Client) (httpReq *http.Request, err error) {
	var (
		logp              = `toHTTPRequest`
		paramsAsURLValues url.Values
		paramsAsJSON      []byte
		contentType       = creq.Type.String()
		path              strings.Builder
		body              io.Reader
		strBody           string
		isParamsURLValues bool
	)

	if client != nil {
		path.WriteString(client.opts.ServerURL)
	}
	path.WriteString(creq.Path)
	paramsAsURLValues, isParamsURLValues = creq.Params.(url.Values)

	switch creq.Method {
	case RequestMethodGet,
		RequestMethodConnect,
		RequestMethodDelete,
		RequestMethodHead,
		RequestMethodOptions,
		RequestMethodTrace:

		if isParamsURLValues {
			path.WriteString("?")
			path.WriteString(paramsAsURLValues.Encode())
		}

	case RequestMethodPatch,
		RequestMethodPost,
		RequestMethodPut:

		switch creq.Type {
		case RequestTypeNone, RequestTypeXML:
			// NOOP.

		case RequestTypeQuery:
			if isParamsURLValues {
				path.WriteString("?")
				path.WriteString(paramsAsURLValues.Encode())
			}

		case RequestTypeForm:
			if isParamsURLValues {
				strBody = paramsAsURLValues.Encode()
				body = strings.NewReader(strBody)
			}

		case RequestTypeMultipartForm:
			paramsAsMultipart, ok := creq.Params.(map[string][]byte)
			if ok {
				contentType, strBody, err = GenerateFormData(paramsAsMultipart)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", logp, err)
				}
				body = strings.NewReader(strBody)
			}

		case RequestTypeJSON:
			if creq.Params != nil {
				paramsAsJSON, err = json.Marshal(creq.Params)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", logp, err)
				}
				body = bytes.NewReader(paramsAsJSON)
			}
		}
	}

	var ctx = context.Background()

	httpReq, err = http.NewRequestWithContext(ctx, creq.Method.String(), path.String(), body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if client != nil {
		setHeaders(httpReq, client.opts.Headers)
	}
	setHeaders(httpReq, creq.Header)

	if len(contentType) > 0 {
		httpReq.Header.Set(HeaderContentType, contentType)
	}

	return httpReq, nil
}

// paramsAsURLEncoded convert the Params as [url.Values] and return the
// [Encode]d value.
// If Params is nil or Params is not [url.Values], it will return an empty
// string.
func (creq *ClientRequest) paramsAsURLEncoded() string {
	if creq.Params == nil {
		return ``
	}

	var (
		params url.Values
		ok     bool
	)
	params, ok = creq.Params.(url.Values)
	if !ok {
		return ``
	}
	return params.Encode()
}

// paramsAsMultipart convert the Params as "map[string][]byte" and return the
// content type and body.
func (creq *ClientRequest) paramsAsMultipart() (params map[string][]byte) {
	if creq.Params == nil {
		return nil
	}

	var ok bool

	params, ok = creq.Params.(map[string][]byte)
	if !ok {
		return nil
	}

	return params
}

// paramsAsBytes convert the Params as []byte.
func (creq *ClientRequest) paramsAsBytes() (body []byte) {
	if creq.Params == nil {
		return nil
	}

	var ok bool

	body, ok = creq.Params.([]byte)
	if !ok {
		return nil
	}

	return body
}
