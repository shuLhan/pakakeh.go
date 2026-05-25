// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// ClientRequest define the parameters for each Client methods.
type ClientRequest struct {
	// Additional header to be send on request.
	// If the same header exist in [ClientOptions.Header] it will replace
	// it.
	// This field is optional.
	Header http.Header

	// Parameter to be send on request.
	// This field is optional.
	// It will converted based on Type rules below,
	//
	// * If Type is [RequestTypeQuery] and Params is [url.Values] it
	//   will be added as query parameters in the Path.
	//
	// * If Type is [RequestTypeForm] and Params is [url.Values] it
	//   will be added as URL encoded in the body.
	//
	// * If Type is [RequestTypeMultipartForm] and Params is
	//   map[string][]byte, then it will be converted as multipart form
	//   in the body.
	//
	// * If Type is [RequestTypeJSON] and Params is not nil, the params
	//   will be encoded to JSON and set in body.
	Params any

	body io.Reader

	// The Path to resource on the server.
	// This field is required, default to "/".
	Path string

	contentType string

	// The HTTP method of request.
	// This field is optional, default to [RequestMethodGet].
	Method RequestMethod

	// The Type of request.
	// It's affect how the Params is encoded in the path or body.
	// This field is optional.
	Type RequestType
}

// toHTTPRequest converts the ClientRequest to the standard [http.Request].
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
		case RequestTypeNone, RequestTypeHTML, RequestTypeXML:
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
			paramsAsMultipart, ok := creq.Params.(*multipart.Form)
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

	httpReq, err = http.NewRequestWithContext(ctx, string(creq.Method), path.String(), body)
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

// paramsAsURLEncoded returns the Params as encoded [url.Values] string.
// If Params is not [url.Values], it will return an empty string.
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

// paramsAsMultipart returns the Params as [*multipart.Form].
func (creq *ClientRequest) paramsAsMultipart() (params *multipart.Form) {
	if creq.Params == nil {
		return nil
	}

	var ok bool

	params, ok = creq.Params.(*multipart.Form)
	if !ok {
		return nil
	}

	return params
}

// paramsAsBytes returns the Params as []byte.
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
