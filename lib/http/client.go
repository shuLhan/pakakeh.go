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
)

//
// Client is a wrapper for standard http.Client with simplified
// functionalities.
//
type Client struct {
	*http.Client
	serverURL string
}

//
// NewClient create and initialize new Client connection using serverURL to
// minimize repetition.
// The serverURL is any path that is static and will never changes during
// request to server.
//
func NewClient(serverURL string) (client *Client) {
	client = &Client{
		serverURL: serverURL,
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

	httpRes, err := client.Client.Get(client.serverURL + path)
	if err != nil {
		return nil, fmt.Errorf("http: Get: %w", err)
	}

	resBody, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("http: Get: %w", err)
	}

	err = httpRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("http: Get: %w", err)
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
	body := strings.NewReader(params.Encode())

	url := client.serverURL + path
	httpRes, err := client.Client.Post(url, ContentTypeForm, body)
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

	contentType, body, err := generateFormData(params)
	if err != nil {
		return nil, fmt.Errorf("http: PostFormData: %w", err)
	}

	httpRes, err := client.Client.Post(url, contentType,
		strings.NewReader(body))
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

	httpRes, err := client.Client.Post(url, ContentTypeJSON,
		bytes.NewReader(paramsJSON))
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
