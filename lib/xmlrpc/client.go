// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/shuLhan/share/lib/debug"
	libhttp "github.com/shuLhan/share/lib/http"
	libnet "github.com/shuLhan/share/lib/net"
)

const (
	defaultTimeout = 10 * time.Second
)

//
// Client for XML-RPC.
//
type Client struct {
	conn    *libhttp.Client
	timeout time.Duration
	url     *url.URL
}

//
// NewClient create and initialize new connection to RPC server.
//
func NewClient(url *url.URL, timeout time.Duration) (client *Client, err error) {
	if url == nil {
		return nil, nil
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}

	host, ip, port := libnet.ParseIPPort(url.Host, 0)

	client = &Client{
		url:     url,
		timeout: timeout,
	}
	clientOpts := &libhttp.ClientOptions{
		Timeout: timeout,
	}

	if url.Scheme == schemeIsHTTPS {
		if ip != nil {
			clientOpts.AllowInsecure = true
		}
		if port == 0 {
			port = 443
		}
		clientOpts.ServerUrl = fmt.Sprintf("https://%s:%d", host, port)
	} else {
		if port == 0 {
			port = 80
		}
		clientOpts.ServerUrl = fmt.Sprintf("http://%s:%d", host, port)
	}

	client.conn = libhttp.NewClient(clientOpts)

	return client, nil
}

//
// Close the client connection.
//
func (cl *Client) Close() {
	cl.url = nil
	cl.conn = nil
}

//
// Send the RPC method with parameters to the server.
//
func (cl *Client) Send(req Request) (resp Response, err error) {
	var (
		logp = "Client.Send"
	)

	xmlbin, _ := req.MarshalText()
	reqBody := bytes.NewReader(xmlbin)

	if debug.Value >= 3 {
		fmt.Printf("<<< Send: request body:\n%s\n", xmlbin)
	}

	httpRequest, err := http.NewRequest("POST", cl.url.String(), reqBody)
	if err != nil {
		return resp, fmt.Errorf("%s: %w", logp, err)
	}

	httpRequest.Header.Set(libhttp.HeaderContentType, libhttp.ContentTypeXML)

	_, resBody, err := cl.conn.Do(httpRequest)
	if err != nil {
		return resp, fmt.Errorf("%s: %w", logp, err)
	}

	if debug.Value >= 3 {
		fmt.Printf(">>> Send: response:\n%s\n", resBody)
	}

	if len(resBody) > 0 {
		err = resp.UnmarshalText(resBody)
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}
