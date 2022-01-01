// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
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
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/shuLhan/share"
	"github.com/shuLhan/share/lib/debug"
)

const (
	defUserAgent = "libhttp/" + share.Version +
		" (github.com/shuLhan/share/lib/http; ms@kilabit.info)"
)

//
// Client is a wrapper for standard http.Client with simplified usabilities,
// including setting default headers, uncompressing response body.
//
type Client struct {
	flateReader io.ReadCloser
	gzipReader  *gzip.Reader

	opts *ClientOptions

	*http.Client
}

//
// NewClient create and initialize new Client.
//
// The client will have KeepAlive timeout set to 30 seconds, with 1 maximum
// idle connection, and 90 seconds IdleConnTimeout.
//
func NewClient(opts *ClientOptions) (client *Client) {
	opts.init()

	httpTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   opts.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          1,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   opts.Timeout,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client = &Client{
		opts:   opts,
		Client: &http.Client{},
	}
	if opts.AllowInsecure {
		httpTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: opts.AllowInsecure,
		}
	}
	client.Client.Transport = httpTransport

	client.setUserAgent()

	return client
}

//
// Delete send the DELETE request to server using path and params as query
// parameters.
// On success, it will return the uncompressed response body.
//
func (client *Client) Delete(path string, headers http.Header, params url.Values) (
	httpRes *http.Response, resBody []byte, err error,
) {
	if params != nil {
		path += "?" + params.Encode()
	}

	return client.doRequest(http.MethodDelete, headers, path, "", nil)
}

//
// Do overwrite the standard http Client.Do to allow debugging request and
// response, and to read and return the response body immediately.
//
func (client *Client) Do(httpRequest *http.Request) (
	httpRes *http.Response, resBody []byte, err error,
) {
	logp := "Do"

	if debug.Value >= 3 {
		dump, err := httputil.DumpRequestOut(httpRequest, true)
		if err != nil {
			log.Printf("%s: %s\n", logp, err)
		} else {
			fmt.Printf("%s\n", dump)
		}
	}

	httpRes, err = client.Client.Do(httpRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", logp, err)
	}

	if debug.Value >= 3 {
		dump, err := httputil.DumpResponse(httpRes, true)
		if err != nil {
			log.Printf("%s: %s", logp, err)
		} else {
			fmt.Printf("%s\n", dump)
		}
	}

	rawBody, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", logp, err)
	}

	err = httpRes.Body.Close()
	if err != nil {
		return httpRes, resBody, fmt.Errorf("%s: %w", logp, err)
	}

	resBody, err = client.uncompress(httpRes, rawBody)
	if err != nil {
		return httpRes, resBody, fmt.Errorf("%s: %w", logp, err)
	}

	// Recreate the body to prevent error on caller.
	httpRes.Body = io.NopCloser(bytes.NewReader(rawBody))

	return httpRes, resBody, nil
}

//
// GenerateHttpRequest generate http.Request from method, path, requestType,
// headers, and params.
//
// For HTTP method GET, CONNECT, DELETE, HEAD, OPTIONS, or TRACE; the params
// value should be nil or url.Values.
// If its url.Values, then the params will be encoded as query parameters.
//
// For HTTP method is PATCH, POST, or PUT; the params will converted based on
// requestType rules below,
//
// * If requestType is RequestTypeQuery and params is url.Values it will be
// added as query parameters in the path.
//
// * If requestType is RequestTypeForm and params is url.Values it will be
// added as URL encoded in the body.
//
// * If requestType is RequestTypeMultipartForm and params type is
// map[string][]byte, then it will be converted as multipart form in the
// body.
//
// * If requestType is RequestTypeJSON and params is not nil, the params will
// be encoded as JSON in body.
//
func (client *Client) GenerateHttpRequest(
	method RequestMethod,
	path string,
	requestType RequestType,
	headers http.Header,
	params interface{},
) (httpRequest *http.Request, err error) {
	var (
		logp              = "GenerateHttpRequest"
		paramsAsUrlValues url.Values
		isParamsUrlValues bool
		paramsAsJSON      []byte
		contentType       = requestType.String()
		strBody           string
		body              io.Reader
	)

	paramsAsUrlValues, isParamsUrlValues = params.(url.Values)

	switch method {
	case RequestMethodGet,
		RequestMethodConnect,
		RequestMethodDelete,
		RequestMethodHead,
		RequestMethodOptions,
		RequestMethodTrace:

		if isParamsUrlValues {
			path += "?" + paramsAsUrlValues.Encode()
		}

	case RequestMethodPatch,
		RequestMethodPost,
		RequestMethodPut:
		switch requestType {
		case RequestTypeQuery:
			if isParamsUrlValues {
				path += "?" + paramsAsUrlValues.Encode()
			}

		case RequestTypeForm:
			if isParamsUrlValues {
				strBody = paramsAsUrlValues.Encode()
				body = strings.NewReader(strBody)
			}

		case RequestTypeMultipartForm:
			paramsAsMultipart, ok := params.(map[string][]byte)
			if ok {
				contentType, strBody, err = generateFormData(paramsAsMultipart)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", logp, err)
				}

				body = strings.NewReader(strBody)
			}

		case RequestTypeJSON:
			if params != nil {
				paramsAsJSON, err = json.Marshal(params)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", logp, err)
				}
				body = bytes.NewReader(paramsAsJSON)
			}
		}
	}

	fullURL := client.opts.ServerUrl + path

	httpRequest, err = http.NewRequest(method.String(), fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	client.setHeaders(httpRequest, client.opts.Headers)
	client.setHeaders(httpRequest, headers)

	if len(contentType) > 0 {
		httpRequest.Header.Set(HeaderContentType, contentType)
	}

	return httpRequest, nil
}

//
// Get send the GET request to server using path and params as query
// parameters.
// On success, it will return the uncompressed response body.
//
func (client *Client) Get(path string, headers http.Header, params url.Values) (
	httpRes *http.Response, resBody []byte, err error,
) {
	if params != nil {
		path += "?" + params.Encode()
	}

	return client.doRequest(http.MethodGet, headers, path, "", nil)
}

//
// Post send the POST request to path without setting "Content-Type".
// If the params is not nil, it will send as query parameters in the path.
//
func (client *Client) Post(path string, headers http.Header, params url.Values) (
	httpRes *http.Response, resBody []byte, err error,
) {
	if params != nil {
		path += "?" + params.Encode()
	}

	return client.doRequest(http.MethodPost, headers, path, "", nil)
}

//
// PostForm send the POST request to path using
// "application/x-www-form-urlencoded".
//
func (client *Client) PostForm(path string, headers http.Header, params url.Values) (
	httpRes *http.Response, resBody []byte, err error,
) {
	body := strings.NewReader(params.Encode())

	return client.doRequest(http.MethodPost, headers, path, ContentTypeForm, body)
}

//
// PostFormData send the POST request to path with all parameters is send
// using "multipart/form-data".
//
func (client *Client) PostFormData(
	path string,
	headers http.Header,
	params map[string][]byte,
) (
	httpRes *http.Response, resBody []byte, err error,
) {
	contentType, strBody, err := generateFormData(params)
	if err != nil {
		return nil, nil, fmt.Errorf("http: PostFormData: %w", err)
	}

	body := strings.NewReader(strBody)

	return client.doRequest(http.MethodPost, headers, path, contentType, body)
}

//
// PostJSON send the POST request with content type set to "application/json"
// and params encoded automatically to JSON.
//
func (client *Client) PostJSON(path string, headers http.Header, params interface{}) (
	httpRes *http.Response, resBody []byte, err error,
) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, nil, fmt.Errorf("PostJSON: %w", err)
	}

	body := bytes.NewReader(paramsJSON)

	return client.doRequest(http.MethodPost, headers, path, ContentTypeJSON, body)
}

//
// Put send the HTTP PUT request with specific content type and body to
// specific path at the server.
//
func (client *Client) Put(path string, headers http.Header, body []byte) (
	*http.Response, []byte, error,
) {
	bodyReader := bytes.NewReader(body)
	return client.doRequest(http.MethodPut, headers, path, "", bodyReader)
}

//
// PutJSON send the PUT request with content type set to "application/json"
// and params encoded automatically to JSON.
//
func (client *Client) PutJSON(path string, headers http.Header, params interface{}) (
	httpRes *http.Response, resBody []byte, err error,
) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, nil, fmt.Errorf("PutJSON: %w", err)
	}

	body := bytes.NewReader(paramsJSON)

	return client.doRequest(http.MethodPut, headers, path, ContentTypeJSON, body)
}

func (client *Client) doRequest(
	httpMethod string,
	headers http.Header,
	path, contentType string,
	body io.Reader,
) (
	httpRes *http.Response, resBody []byte, err error,
) {
	fullURL := client.opts.ServerUrl + path

	httpReq, err := http.NewRequest(httpMethod, fullURL, body)
	if err != nil {
		return nil, nil, err
	}

	client.setHeaders(httpReq, client.opts.Headers)
	client.setHeaders(httpReq, headers)

	if len(contentType) > 0 {
		httpReq.Header.Set(HeaderContentType, contentType)
	}

	return client.Do(httpReq)
}

//
// setHeaders set the request headers.
//
func (client *Client) setHeaders(req *http.Request, headers http.Header) {
	for k, v := range headers {
		for x, hv := range v {
			if x == 0 {
				req.Header.Set(k, hv)
			} else {
				req.Header.Add(k, hv)
			}
		}
	}
}

//
// setUserAgent set the User-Agent header only if its not defined by user.
//
func (client *Client) setUserAgent() {
	v := client.opts.Headers.Get(HeaderUserAgent)
	if len(v) > 0 {
		return
	}
	client.opts.Headers.Set(HeaderUserAgent, defUserAgent)
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
		buf           = make([]byte, 1024)
	)

	switch res.Header.Get(HeaderContentEncoding) {
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
