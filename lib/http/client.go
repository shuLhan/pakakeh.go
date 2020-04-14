// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shuLhan/share"
)

const (
	defUserAgent = "libhttp/" + share.Version +
		" (github.com/shuLhan/share/lib/http; m.shulhan@gmail.com)"
)

//
// Client is a wrapper for standard http.Client with simplified
// functionalities.
//
type Client struct {
	*http.Client
	serverURL  string
	defHeaders http.Header
}

//
// NewClient create and initialize new Client connection using serverURL to
// minimize repetition.
// The serverURL is any path that is static and will never changes during
// request to server.
// The headers parameter define default headers that will be set in any
// request to server.
//
func NewClient(serverURL string, headers http.Header) (client *Client) {
	client = &Client{
		serverURL:  serverURL,
		defHeaders: headers,
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}

	client.setUserAgent()

	return client
}

//
// Get send the GET request to server using path and params as query
// parameters.
// On success, it will return the response body.
//
func (client *Client) Get(path string, params url.Values) (
	resBody []byte, err error,
) {
	if params != nil {
		path += "?" + params.Encode()
	}

	url := client.serverURL + path
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Get: %w", err)
	}

	client.setHeaders(httpReq)

	httpRes, err := client.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Get: %w", err)
	}

	resBody, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("http: Get: %w", err)
	}

	err = httpRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Get: %w", err)
	}

	return resBody, nil
}

//
// PostForm send the POST request to path using
// "application/x-www-form-urlencoded".
//
func (client *Client) PostForm(path string, params url.Values) (
	resBody []byte, err error,
) {
	url := client.serverURL + path
	body := strings.NewReader(params.Encode())

	httpReq, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("Post: %w", err)
	}

	client.setHeaders(httpReq)
	httpReq.Header.Set(ContentType, ContentTypeForm)

	httpRes, err := client.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Post: %w", err)
	}

	resBody, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("Post: %w", err)
	}

	err = httpRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Post: %w", err)
	}

	return resBody, nil
}

//
// PostFormData send the POST request to path with all parameters is send
// using "multipart/form-data".
//
func (client *Client) PostFormData(path string, params map[string][]byte) (
	resBody []byte, err error,
) {
	url := client.serverURL + path

	contentType, strBody, err := generateFormData(params)
	if err != nil {
		return nil, fmt.Errorf("http: PostFormData: %w", err)
	}

	body := strings.NewReader(strBody)

	httpReq, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("Post: %w", err)
	}

	client.setHeaders(httpReq)
	httpReq.Header.Set(ContentType, contentType)

	httpRes, err := client.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http: PostFormData: %w", err)
	}

	resBody, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("http: PostFormData: %w", err)
	}

	err = httpRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("http: PostFormData: %w", err)
	}

	return resBody, nil
}

//
// PostJSON send the POST request with content type set to "application/json"
// and params encoded automatically to JSON.
//
func (client *Client) PostJSON(path string, params interface{}) (
	resBody []byte, err error,
) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("PostJSON: %w", err)
	}

	url := client.serverURL + path
	body := bytes.NewReader(paramsJSON)

	httpReq, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("PostJSON: %w", err)
	}

	client.setHeaders(httpReq)
	httpReq.Header.Set(ContentType, ContentTypeJSON)

	httpRes, err := client.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("PostJSON: %w", err)
	}

	resBody, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("PostJSON: %w", err)
	}

	err = httpRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("PostJSON: %w", err)
	}

	return resBody, nil
}

//
// setHeaders set the request headers to default headers.
// If the header's key contains more than one value, the last one will be
// used.
//
func (client *Client) setHeaders(req *http.Request) {
	for k, vv := range client.defHeaders {
		for _, v := range vv {
			if len(v) > 0 {
				req.Header.Set(k, v[0])
			}
		}
	}
}

//
// setUserAgent set the User-Agent header only if its not defined by user.
//
func (client *Client) setUserAgent() {
	v := client.defHeaders.Get(UserAgent)
	if len(v) > 0 {
		return
	}
	client.defHeaders.Set(UserAgent, defUserAgent)
}

func generateFormData(params map[string][]byte) (
	contentType, body string, err error,
) {
	sb := new(strings.Builder)
	w := multipart.NewWriter(sb)
	for k, v := range params {
		part, err := w.CreateFormField(k)
		if err != nil {
			return "", "", err
		}
		_, err = part.Write(v)
		if err != nil {
			return "", "", err
		}
	}

	err = w.Close()
	if err != nil {
		return "", "", err
	}

	contentType = w.FormDataContentType()

	return contentType, sb.String(), nil
}
