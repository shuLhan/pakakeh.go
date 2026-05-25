// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

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
	"strings"
	"time"

	pakakeh "kilabit.info/pakakeh.go"
)

var (
	defUserAgent = `libhttp/` + pakakeh.Version
)

// Client is a wrapper for standard [http.Client] with simplified
// usabilities, including setting default headers and uncompressing response
// body.
type Client struct {
	flateReader io.ReadCloser
	gzipReader  *gzip.Reader

	*http.Client

	opts ClientOptions
}

// NewClient returns initialized new [Client].
//
// The client have [net.Dialer.KeepAlive] set to 30 seconds,
// [http.Transport.MaxIdleConns] set to 1, and
// [http.Transport.IdleConnTimeout] set to 90 seconds.
func NewClient(opts ClientOptions) (client *Client) {
	opts.init()

	client = &Client{
		opts:   opts,
		Client: &http.Client{},
	}

	client.Client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   opts.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          1,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   opts.Timeout,
		ExpectContinueTimeout: opts.Timeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.AllowInsecure,
		},
	}

	client.setUserAgent()

	return client
}

// Delete returns the response after sending DELETE request to server.
// The request send to [ClientRequest.Path] with [ClientRequest.Params] as
// query parameters.
// The returned response body is uncompressed.
func (client *Client) Delete(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodDelete

	return client.doRequest(req)
}

// Do wraps the standard [http.Client.Do] for reading and decompressing the
// response body.
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

// Download downloads resource from remote server and write it into
// [DownloadRequest.Output].
//
// If the [DownloadRequest.Output] is nil, it will return an error
// [ErrClientDownloadNoOutput].
// If server return HTTP code other than 200, it will return non-nil
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

// GenerateHTTPRequest converts the [ClientRequest] to [*http.Request] with
// body set based on [ClientRequest.Type].
//
// For HTTP method GET, CONNECT, DELETE, HEAD, OPTIONS, or TRACE; the
// [ClientRequest.Params] value [should be nil] or [url.Values].
//
// For HTTP method PATCH, POST, or PUT; the [ClientRequest.Params] is
// converted based on [ClientRequest.Type] rules below,
//
//   - If Type is [RequestTypeQuery] and Params is [url.Values] it will set
//     the query parameters in Path.
//   - If Type is [RequestTypeForm] and Params is [url.Values] it
//     will be set the body as URL encoded.
//   - If Type is [RequestTypeMultipartForm] and Params type is
//     [*multipart.Form], then it will set the bdoy as multipart form.
//   - If Type is [RequestTypeJSON] and Params is not nil, it will set the
//     body by encoding the Params into JSON.
//
// [should be nil]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/GET
func (client *Client) GenerateHTTPRequest(req ClientRequest) (httpReq *http.Request, err error) {
	var (
		logp          = `GenerateHTTPRequest`
		contentType   = req.Type.String()
		paramsEncoded = req.paramsAsURLEncoded()

		body io.Reader
	)

	switch req.Type {
	case RequestTypeNone, RequestTypeHTML, RequestTypeXML:
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
			mpform *multipart.Form
			ok     bool
		)

		mpform, ok = req.Params.(*multipart.Form)
		if ok {
			var strBody string

			contentType, strBody, err = GenerateFormData(mpform)
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

// Get returns the response from sending GET request to server using
// [ClientRequest.Path] as target endpoint and [ClientRequest.Params] in query
// parameters.
// The returned response body is uncompressed.
func (client *Client) Get(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodGet

	return client.doRequest(req)
}

// Head returns the response from sending HEAD request to the server.
// The returned response body is nil.
func (client *Client) Head(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodHead

	return client.doRequest(req)
}

// Post returns the response from sending the POST request to server.
// The Params is send in query parameters.
func (client *Client) Post(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()
	if len(params) != 0 {
		req.Path += `?` + params
	}

	req.Method = RequestMethodPost

	return client.doRequest(req)
}

// PostForm returns the response from sending the POST request to server.
// The Params is send in body with content-type set to [ContentTypeForm].
func (client *Client) PostForm(req ClientRequest) (res *ClientResponse, err error) {
	var params = req.paramsAsURLEncoded()

	req.Method = RequestMethodPost
	req.contentType = ContentTypeForm
	req.body = strings.NewReader(params)

	return client.doRequest(req)
}

// PostFormData returns the response from sending the POST request to server.
// The Params is send in body with content-type set to
// [ContentTypeMultipartForm].
func (client *Client) PostFormData(req ClientRequest) (res *ClientResponse, err error) {
	logp := `PostFormData`

	req.contentType = req.Type.String()

	params := req.paramsAsMultipart()
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

// PostJSON returns the response from sending POST request to server.
// The Params is encoded to JSON and send in body with content-type set to
// [ContentTypeJSON].
// Params must be a type than can be marshalled with [json.Marshal] or
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

// Put returns the response from sending PUT request to server.
// The Params is send in body as raw bytes without setting content-type.
// The Content-Type can be set in the Header before calling this method.
func (client *Client) Put(req ClientRequest) (*ClientResponse, error) {
	var params = req.paramsAsBytes()

	req.Method = RequestMethodPut
	req.body = bytes.NewReader(params)

	return client.doRequest(req)
}

// PutForm returns the response from sending PUT request to server.
// The Params is send in body using content-type [ContentTypeForm].
func (client *Client) PutForm(req ClientRequest) (*ClientResponse, error) {
	var params = req.paramsAsURLEncoded()

	req.Method = RequestMethodPut
	req.contentType = ContentTypeForm
	req.body = strings.NewReader(params)

	return client.doRequest(req)
}

// PutFormData returns the response from sending PUT request to server.
// The Params is send in body with content-type [ContentTypeMultipartForm].
func (client *Client) PutFormData(req ClientRequest) (res *ClientResponse, err error) {
	var (
		logp = `PutFormData`

		params *multipart.Form
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

// PutJSON returns the response from sending PUT request to server.
// The Params is encoded to JSON and send in body with content-type set to
// [ContentTypeJSON].
func (client *Client) PutJSON(req ClientRequest) (res *ClientResponse, err error) {
	var (
		logp   = `PutJSON`
		params []byte
	)

	params, err = json.Marshal(req.Params)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	req.Method = RequestMethodPut
	req.contentType = ContentTypeJSON
	req.body = bytes.NewReader(params)

	return client.doRequest(req)
}

// Transport returns the [http.Transport] that Client use.
func (client *Client) Transport() (transport *http.Transport) {
	transport, _ = client.Client.Transport.(*http.Transport)
	return transport
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
