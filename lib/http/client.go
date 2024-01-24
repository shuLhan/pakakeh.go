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
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/shuLhan/share"
)

var (
	defUserAgent = `libhttp/` + share.Version
)

// Client is a wrapper for standard [http.Client] with simplified
// usabilities, including setting default headers, uncompressing response
// body.
type Client struct {
	flateReader io.ReadCloser
	gzipReader  *gzip.Reader

	opts *ClientOptions

	*http.Client
}

// NewClient create and initialize new [Client].
//
// The client will have [net.Dialer.KeepAlive] set to 30 seconds, with one
// [http.Transport.MaxIdleConns], and 90 seconds
// [http.Transport.IdleConnTimeout].
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

// Delete send the DELETE request to server using rpath as target endpoint
// and params as query parameters.
// On success, it will return the uncompressed response body.
func (client *Client) Delete(rpath string, hdr http.Header, params url.Values) (
	res *http.Response, resBody []byte, err error,
) {
	if params != nil {
		rpath += `?` + params.Encode()
	}

	return client.doRequest(http.MethodDelete, hdr, rpath, ``, nil)
}

// Do overwrite the standard [http.Client.Do] to allow debugging request and
// response, and to read and return the response body immediately.
func (client *Client) Do(req *http.Request) (res *http.Response, resBody []byte, err error) {
	logp := "Do"

	res, err = client.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", logp, err)
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", logp, err)
	}

	err = res.Body.Close()
	if err != nil {
		return res, resBody, fmt.Errorf("%s: %w", logp, err)
	}

	resBody, err = client.uncompress(res, rawBody)
	if err != nil {
		return res, resBody, fmt.Errorf("%s: %w", logp, err)
	}

	// Recreate the body to prevent error on caller.
	res.Body = io.NopCloser(bytes.NewReader(rawBody))

	return res, resBody, nil
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
		return nil, fmt.Errorf("%s: %s", logp, err)
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
			err = fmt.Errorf("%w: %s", err, errClose)
		}
	}

	return res, err
}

// GenerateHttpRequest generate [http.Request] from method, rpath,
// rtype, hdr, and params.
//
// For HTTP method GET, CONNECT, DELETE, HEAD, OPTIONS, or TRACE; the params
// value should be nil or [url.Values].
// If its [url.Values], then the params will be encoded as query parameters.
//
// For HTTP method is PATCH, POST, or PUT; the params will converted based
// on rtype rules below,
//
//   - If rtype is [RequestTypeQuery] and params is [url.Values] it
//     will be added as query parameters in the rpath.
//   - If rtype is [RequestTypeForm] and params is [url.Values] it
//     will be added as URL encoded in the body.
//   - If rtype is [RequestTypeMultipartForm] and params type is
//     map[string][]byte, then it will be converted as multipart form in the
//     body.
//   - If rtype is [RequestTypeJSON] and params is not nil, the params
//     will be encoded as JSON in body.
//
//revive:disable-next-line
func (client *Client) GenerateHttpRequest(
	method RequestMethod,
	rpath string,
	rtype RequestType,
	hdr http.Header,
	params interface{},
) (req *http.Request, err error) {
	var (
		logp              = "GenerateHttpRequest"
		paramsAsURLValues url.Values
		isParamsURLValues bool
		paramsAsJSON      []byte
		contentType       = rtype.String()
		strBody           string
		body              io.Reader
	)

	paramsAsURLValues, isParamsURLValues = params.(url.Values)

	switch method {
	case RequestMethodGet,
		RequestMethodConnect,
		RequestMethodDelete,
		RequestMethodHead,
		RequestMethodOptions,
		RequestMethodTrace:

		if isParamsURLValues {
			rpath += `?` + paramsAsURLValues.Encode()
		}

	case RequestMethodPatch,
		RequestMethodPost,
		RequestMethodPut:
		switch rtype {
		case RequestTypeQuery:
			if isParamsURLValues {
				rpath += `?` + paramsAsURLValues.Encode()
			}

		case RequestTypeForm:
			if isParamsURLValues {
				strBody = paramsAsURLValues.Encode()
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

	rpath = path.Join(`/`, rpath)
	fullURL := client.opts.ServerUrl + rpath

	req, err = http.NewRequest(method.String(), fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	setHeaders(req, client.opts.Headers)
	setHeaders(req, hdr)

	if len(contentType) > 0 {
		req.Header.Set(HeaderContentType, contentType)
	}

	return req, nil
}

// Get send the GET request to server using rpath as target endpoint and
// params as query parameters.
// On success, it will return the uncompressed response body.
func (client *Client) Get(rpath string, hdr http.Header, params url.Values) (
	res *http.Response, resBody []byte, err error,
) {
	if params != nil {
		rpath += `?` + params.Encode()
	}

	return client.doRequest(http.MethodGet, hdr, rpath, ``, nil)
}

// Head send the HEAD request to rpath endpoint, with optional hdr and
// params in query parameters.
// The returned resBody shoule be always nil.
func (client *Client) Head(rpath string, hdr http.Header, params url.Values) (
	res *http.Response, resBody []byte, err error,
) {
	if params != nil {
		rpath += `?` + params.Encode()
	}
	return client.doRequest(http.MethodHead, hdr, rpath, ``, nil)
}

// Post send the POST request to rpath without setting "Content-Type".
// If the params is not nil, it will send as query parameters in the rpath.
func (client *Client) Post(rpath string, hdr http.Header, params url.Values) (
	res *http.Response, resBody []byte, err error,
) {
	if params != nil {
		rpath += `?` + params.Encode()
	}

	return client.doRequest(http.MethodPost, hdr, rpath, ``, nil)
}

// PostForm send the POST request to rpath using
// "application/x-www-form-urlencoded".
func (client *Client) PostForm(rpath string, hdr http.Header, params url.Values) (
	res *http.Response, resBody []byte, err error,
) {
	body := strings.NewReader(params.Encode())

	return client.doRequest(http.MethodPost, hdr, rpath, ContentTypeForm, body)
}

// PostFormData send the POST request to rpath with all parameters is send
// using "multipart/form-data".
func (client *Client) PostFormData(
	rpath string,
	hdr http.Header,
	params map[string][]byte,
) (
	res *http.Response, resBody []byte, err error,
) {
	contentType, strBody, err := generateFormData(params)
	if err != nil {
		return nil, nil, fmt.Errorf("http: PostFormData: %w", err)
	}

	body := strings.NewReader(strBody)

	return client.doRequest(http.MethodPost, hdr, rpath, contentType, body)
}

// PostJSON send the POST request with content type set to "application/json"
// and params encoded automatically to JSON.
// The params must be a type than can be marshalled with [json.Marshal] or
// type that implement [json.Marshaler].
func (client *Client) PostJSON(rpath string, hdr http.Header, params interface{}) (
	res *http.Response, resBody []byte, err error,
) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, nil, fmt.Errorf("PostJSON: %w", err)
	}

	body := bytes.NewReader(paramsJSON)

	return client.doRequest(http.MethodPost, hdr, rpath, ContentTypeJSON, body)
}

// Put send the HTTP PUT request to rpath with optional, raw body.
// The Content-Type can be set in the hdr.
func (client *Client) Put(rpath string, hdr http.Header, body []byte) (
	*http.Response, []byte, error,
) {
	bodyReader := bytes.NewReader(body)
	return client.doRequest(http.MethodPut, hdr, rpath, ``, bodyReader)
}

// PutForm send the PUT request with params set in body using content type
// "application/x-www-form-urlencoded".
func (client *Client) PutForm(rpath string, hdr http.Header, params url.Values) (
	res *http.Response, resBody []byte, err error,
) {
	var body = strings.NewReader(params.Encode())

	return client.doRequest(http.MethodPut, hdr, rpath, ContentTypeForm, body)
}

// PutFormData send the PUT request with params set in body using content type
// "multipart/form-data".
func (client *Client) PutFormData(rpath string, hdr http.Header, params map[string][]byte) (
	res *http.Response, resBody []byte, err error,
) {
	var (
		contentType string
		strBody     string
		body        *strings.Reader
	)

	contentType, strBody, err = generateFormData(params)
	if err != nil {
		return nil, nil, fmt.Errorf(`http: PutFormData: %w`, err)
	}

	body = strings.NewReader(strBody)

	return client.doRequest(http.MethodPut, hdr, rpath, contentType, body)
}

// PutJSON send the PUT request with content type set to "application/json"
// and params encoded automatically to JSON.
func (client *Client) PutJSON(rpath string, hdr http.Header, params interface{}) (
	res *http.Response, resBody []byte, err error,
) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, nil, fmt.Errorf("PutJSON: %w", err)
	}

	body := bytes.NewReader(paramsJSON)

	return client.doRequest(http.MethodPut, hdr, rpath, ContentTypeJSON, body)
}

func (client *Client) doRequest(
	httpMethod string,
	hdr http.Header,
	rpath, contentType string,
	body io.Reader,
) (
	res *http.Response, resBody []byte, err error,
) {
	rpath = path.Join(`/`, rpath)
	fullURL := client.opts.ServerUrl + rpath

	httpReq, err := http.NewRequest(httpMethod, fullURL, body)
	if err != nil {
		return nil, nil, err
	}

	setHeaders(httpReq, client.opts.Headers)
	setHeaders(httpReq, hdr)

	if len(contentType) > 0 {
		httpReq.Header.Set(HeaderContentType, contentType)
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

// generateFormData generate multipart/form-data body from params.
func generateFormData(params map[string][]byte) (contentType, body string, err error) {
	var (
		sb = new(strings.Builder)
		w  = multipart.NewWriter(sb)

		part io.Writer
		k    string
		v    []byte
	)
	for k, v = range params {
		part, err = w.CreateFormField(k)
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
