// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
// Client is a wrapper for standard http.Client with simplified usabilities,
// including setting default headers, uncompressing response body.
//
type Client struct {
	*http.Client

	flateReader io.ReadCloser
	gzipReader  *gzip.Reader

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
// The insecure parameter allow to connect to remote server with unknown
// certificate authority.
//
func NewClient(serverURL string, headers http.Header, insecure bool) (client *Client) {
	if headers == nil {
		headers = make(http.Header)
	}
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
				ForceAttemptHTTP2: true,
				MaxIdleConns:      100,
				IdleConnTimeout:   90 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecure,
				},
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
// On success, it will return the uncompressed response body.
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

	return client.uncompress(httpRes, resBody)
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
	httpReq.Header.Set(HeaderContentType, ContentTypeForm)

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

	return client.uncompress(httpRes, resBody)
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
	httpReq.Header.Set(HeaderContentType, contentType)

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

	return client.uncompress(httpRes, resBody)
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
	httpReq.Header.Set(HeaderContentType, ContentTypeJSON)

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

	return client.uncompress(httpRes, resBody)
}

//
// setHeaders set the request headers to default headers.
// If the header's key contains more than one value, the last one will be
// used.
//
func (client *Client) setHeaders(req *http.Request) {
	for k, v := range client.defHeaders {
		if len(v) > 0 {
			req.Header.Set(k, v[len(v)-1])
		}
	}
}

//
// setUserAgent set the User-Agent header only if its not defined by user.
//
func (client *Client) setUserAgent() {
	v := client.defHeaders.Get(HeaderUserAgent)
	if len(v) > 0 {
		return
	}
	client.defHeaders.Set(HeaderUserAgent, defUserAgent)
}

//
// uncompress the response body only if the response.Uncompressed is false or
// user's is not explicitly disable compression and the Content-Type is
// "text/*" or JSON.
//
func (client *Client) uncompress(res *http.Response, body []byte) (
	out []byte, err error,
) {
	trans := client.Client.Transport.(*http.Transport)
	if res.Uncompressed || trans.DisableCompression {
		return body, nil
	}

	contentType := res.Header.Get(HeaderContentType)
	switch {
	case strings.HasPrefix(contentType, "text/"):
	case strings.HasPrefix(contentType, ContentTypeJSON):
	default:
		return body, nil
	}

	var (
		n   int
		dec io.ReadCloser
		in  io.Reader = bytes.NewReader(body)
		buf []byte    = make([]byte, 1024)
	)

	switch res.Header.Get(ContentEncoding) {
	case ContentEncodingBzip2:
		dec = ioutil.NopCloser(bzip2.NewReader(in))

	case ContentEncodingCompress:
		dec = lzw.NewReader(in, lzw.MSB, 8)

	case ContentEncodingDeflate:
		if client.flateReader == nil {
			client.flateReader = flate.NewReader(in)
		} else {
			err = client.flateReader.(flate.Resetter).Reset(in, nil)
			if err != nil {
				return body, err
			}
		}
		dec = client.flateReader

	case ContentEncodingGzip:
		if client.gzipReader == nil {
			client.gzipReader, err = gzip.NewReader(in)
		} else {
			err = client.gzipReader.Reset(in)
		}
		if err != nil {
			return body, err
		}
		dec = client.gzipReader

	default:
		// Unknown encoding detected, return as is ...
		return body, nil
	}

	for {
		n, err = dec.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			break
		}
		if errors.Is(err, io.EOF) {
			out = append(out, buf[:n]...)
			err = nil
			break
		}
		if n == 0 {
			break
		}
		out = append(out, buf[:n]...)
	}

	errc := dec.Close()
	if errc != nil {
		log.Printf("http.Client: uncompress: %s", errc.Error())
		if err == nil {
			err = errc
		}
	}

	return out, err
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
