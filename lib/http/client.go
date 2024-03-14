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
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"path"
	"sort"
	"strings"
	"time"

	pakakeh "git.sr.ht/~shulhan/pakakeh.go"
)

var (
	defUserAgent = `libhttp/` + pakakeh.Version
)

// Client is a wrapper for standard [http.Client] with simplified
// usabilities, including setting default headers, uncompressing response
// body.
type Client struct {
	flateReader io.ReadCloser
	gzipReader  *gzip.Reader

	*http.Client

	opts ClientOptions
}

// NewClient create and initialize new [Client].
//
// The client will have [net.Dialer.KeepAlive] set to 30 seconds, with one
// [http.Transport.MaxIdleConns], and 90 seconds
// [http.Transport.IdleConnTimeout].
func NewClient(opts ClientOptions) (client *Client) {
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
			InsecureSkipVerify: opts.AllowInsecure, //nolint:gosec
		}
	}
	client.Client.Transport = httpTransport

	client.setUserAgent()

	return client
}

// Delete send the DELETE request to server using rpath as target endpoint
// and params as query parameters.
// On success, it will return the uncompressed response body.
func (client *Client) Delete(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodDelete

	return client.doRequest(req)
}

// Do overwrite the standard [http.Client.Do] to allow debugging request and
// response, and to read and return the response body immediately.
func (client *Client) Do(req *http.Request) (res *ClientResponse, err error) {
	var logp = `Do`

	res = &ClientResponse{}

	res.HTTPResponse, err = client.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var resBody []byte

	resBody, err = io.ReadAll(res.HTTPResponse.Body)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	err = res.HTTPResponse.Body.Close()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	res.Body, err = client.uncompress(res.HTTPResponse, resBody)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	// Recreate the body to prevent error on caller.
	res.HTTPResponse.Body = io.NopCloser(bytes.NewReader(res.Body))

	return res, nil
}

// Download a resource from remote server and write it into
// [DownloadRequest.Output].
//
// If the [DownloadRequest.Output] is nil, it will return an error
// [ErrClientDownloadNoOutput].
// If server return HTTP code beside 200, it will return non-nil
// [http.Response] with an error.
func (client *Client) Download(req DownloadRequest) (res *http.Response, err error) {
	var (
		logp     = "Download"
		httpReq  *http.Request
		errClose error
	)

	if req.Output == nil {
		return nil, fmt.Errorf("%s: %w", logp, ErrClientDownloadNoOutput)
	}

	httpReq, err = req.toHTTPRequest(client)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	res, err = client.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s: %s", logp, res.Status)
		goto out
	}

	_, err = io.Copy(req.Output, res.Body)
	if err != nil {
		err = fmt.Errorf("%s: %w", logp, err)
	}
out:
	errClose = res.Body.Close()
	if errClose != nil {
		if err == nil {
			err = fmt.Errorf("%s: %w", logp, errClose)
		} else {
			err = fmt.Errorf(`%w: %w`, err, errClose)
		}
	}

	return res, err
}

// GenerateHTTPRequest generate [http.Request] from [ClientRequest].
//
// For HTTP method GET, CONNECT, DELETE, HEAD, OPTIONS, or TRACE; the
// [ClientRequest.Params] value should be nil or [url.Values].
// If its [url.Values], then the params will be encoded as query parameters.
//
// For HTTP method PATCH, POST, or PUT; the [ClientRequest.Params] will
// converted based on [ClientRequest.Type] rules below,
//
//   - If Type is [RequestTypeQuery] and Params is [url.Values] it
//     will be send as query parameters in the Path.
//   - If Type is [RequestTypeForm] and Params is [url.Values] it
//     will be send as URL encoded in the body.
//   - If Type is [RequestTypeMultipartForm] and Params type is
//     map[string][]byte, then it will send as multipart form in the
//     body.
//   - If Type is [RequestTypeJSON] and Params is not nil, the Params
//     will be encoded as JSON in the body.
func (client *Client) GenerateHTTPRequest(req ClientRequest) (httpReq *http.Request, err error) {
	var (
		logp          = `GenerateHTTPRequest`
		contentType   = req.Type.String()
		paramsEncoded = req.paramsAsURLEncoded()

		body io.Reader
	)

	switch req.Method {
	case RequestMethodGet,
		RequestMethodConnect,
		RequestMethodDelete,
		RequestMethodHead,
		RequestMethodOptions,
		RequestMethodTrace:

		if len(paramsEncoded) != 0 {
			req.Path += `?` + paramsEncoded
		}

	case RequestMethodPatch,
		RequestMethodPost,
		RequestMethodPut:

		switch req.Type {
		case RequestTypeNone, RequestTypeXML:
			// NOOP.

		case RequestTypeQuery:
			if len(paramsEncoded) != 0 {
				req.Path += `?` + paramsEncoded
			}

		case RequestTypeForm:
			if len(paramsEncoded) != 0 {
				body = strings.NewReader(paramsEncoded)
			}

		case RequestTypeMultipartForm:
			var (
				paramsAsMultipart map[string][]byte
				ok                bool
			)

			paramsAsMultipart, ok = req.Params.(map[string][]byte)
			if ok {
				var strBody string

				contentType, strBody, err = GenerateFormData(paramsAsMultipart)
				if err != nil {
					return nil, fmt.Errorf(`%s: %w`, logp, err)
				}

				body = strings.NewReader(strBody)
			}

		case RequestTypeJSON:
			if req.Params != nil {
				var paramsAsJSON []byte
				paramsAsJSON, err = json.Marshal(req.Params)
				if err != nil {
					return nil, fmt.Errorf(`%s: %w`, logp, err)
				}
				body = bytes.NewReader(paramsAsJSON)
			}
		}
	}

	req.Path = path.Join(`/`, req.Path)
	var (
		fullURL = client.opts.ServerURL + req.Path
		ctx     = context.Background()
	)

	httpReq, err = http.NewRequestWithContext(ctx, string(req.Method), fullURL, body)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	setHeaders(httpReq, client.opts.Headers)
	setHeaders(httpReq, req.Header)

	if len(contentType) > 0 {
		httpReq.Header.Set(HeaderContentType, contentType)
	}

	return httpReq, nil
}

// Get send the GET request to server using [ClientRequest.Path] as target
// endpoint and [ClientRequest.Params] as query parameters.
// On success, it will return the uncompressed response body.
func (client *Client) Get(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodGet

	return client.doRequest(req)
}

// Head send the HEAD request to rpath endpoint, with optional hdr and
// params in query parameters.
// The returned resBody shoule be always nil.
func (client *Client) Head(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodHead

	return client.doRequest(req)
}

// Post send the POST request to rpath without setting "Content-Type".
// If the params is not nil, it will send as query parameters in the rpath.
func (client *Client) Post(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodPost

	return client.doRequest(req)
}

// PostForm send the POST request to rpath using
// "application/x-www-form-urlencoded".
func (client *Client) PostForm(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()

	req.Method = RequestMethodPost
	req.contentType = ContentTypeForm
	req.body = strings.NewReader(params)

	return client.doRequest(req)
}

// PostFormData send the POST request to Path with all parameters is send
// using "multipart/form-data".
func (client *Client) PostFormData(req ClientRequest) (res *ClientResponse, err error) {
	var (
		logp = `PostFormData`

		params map[string][]byte
	)

	req.contentType = req.Type.String()

	params = req.paramsAsMultipart()
	if params == nil {
		req.body = strings.NewReader(``)
	} else {
		var body string

		req.contentType, body, err = GenerateFormData(params)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		req.body = strings.NewReader(body)
	}

	req.Method = RequestMethodPost

	return client.doRequest(req)
}

// PostJSON send the POST request with content type set to "application/json"
// and Params encoded automatically to JSON.
// The Params must be a type than can be marshalled with [json.Marshal] or
// type that implement [json.Marshaler].
func (client *Client) PostJSON(req ClientRequest) (res *ClientResponse, err error) {
	var (
		logp = `PostJSON`

		params []byte
	)

	params, err = json.Marshal(req.Params)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	req.Method = RequestMethodPost
	req.contentType = ContentTypeJSON
	req.body = bytes.NewReader(params)

	return client.doRequest(req)
}

// Put send the HTTP PUT request to rpath with optional, raw body.
// The Content-Type can be set in the hdr.
func (client *Client) Put(req ClientRequest) (*ClientResponse, error) {
	var params = req.paramsAsBytes()

	req.Method = RequestMethodPut
	req.body = bytes.NewReader(params)

	return client.doRequest(req)
}

// PutForm send the PUT request with params set in body using content type
// "application/x-www-form-urlencoded".
func (client *Client) PutForm(req ClientRequest) (*ClientResponse, error) {
	var params = req.paramsAsURLEncoded()

	req.Method = RequestMethodPut
	req.contentType = ContentTypeForm
	req.body = strings.NewReader(params)

	return client.doRequest(req)
}

// PutFormData send the PUT request with params set in body using content type
// "multipart/form-data".
func (client *Client) PutFormData(req ClientRequest) (res *ClientResponse, err error) {
	var (
		logp   = `PutFormData`
		params map[string][]byte
	)

	req.contentType = req.Type.String()

	params = req.paramsAsMultipart()
	if params == nil {
		req.body = strings.NewReader(``)
	} else {
		var body string

		req.contentType, body, err = GenerateFormData(params)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		req.body = strings.NewReader(body)
	}

	req.Method = RequestMethodPut

	return client.doRequest(req)
}

// PutJSON send the PUT request with content type set to "application/json"
// and params encoded automatically to JSON.
func (client *Client) PutJSON(req ClientRequest) (res *ClientResponse, err error) {
	var (
		logp   = `PutJSON`
		params []byte
	)

	params, err = json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	req.Method = RequestMethodPut
	req.contentType = ContentTypeJSON
	req.body = bytes.NewReader(params)

	return client.doRequest(req)
}

func (client *Client) doRequest(req ClientRequest) (res *ClientResponse, err error) {
	req.Path = path.Join(`/`, req.Path)

	var (
		fullURL = client.opts.ServerURL + req.Path
		ctx     = context.Background()

		httpReq *http.Request
	)

	if len(req.Method) == 0 {
		req.Method = RequestMethodGet
	}

	httpReq, err = http.NewRequestWithContext(ctx, string(req.Method), fullURL, req.body)
	if err != nil {
		return nil, err
	}

	setHeaders(httpReq, client.opts.Headers)
	setHeaders(httpReq, req.Header)

	if len(req.contentType) > 0 {
		httpReq.Header.Set(HeaderContentType, req.contentType)
	}

	return client.Do(httpReq)
}

// setUserAgent set the User-Agent header only if its not defined by user.
func (client *Client) setUserAgent() {
	v := client.opts.Headers.Get(HeaderUserAgent)
	if len(v) > 0 {
		return
	}
	client.opts.Headers.Set(HeaderUserAgent, defUserAgent)
}

// uncompress the response body only if the response.Uncompressed is false or
// user's is not explicitly disable compression and the Content-Type is
// "text/*" or JSON.
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
		dec = io.NopCloser(bzip2.NewReader(in))

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

// GenerateFormData generate multipart/form-data body from params.
func GenerateFormData(params map[string][]byte) (contentType, body string, err error) {
	var (
		sb      = new(strings.Builder)
		w       = multipart.NewWriter(sb)
		listKey = make([]string, 0, len(params))

		k string
	)
	for k = range params {
		listKey = append(listKey, k)
	}
	sort.Strings(listKey)

	var (
		part io.Writer
		v    []byte
	)
	for _, k = range listKey {
		part, err = w.CreateFormField(k)
		if err != nil {
			return "", "", err
		}
		v = params[k]
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
