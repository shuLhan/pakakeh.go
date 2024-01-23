// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package http implement custom HTTP server with memory file system and
// simplified routing handler.
//
// # Features
//
// The following enhancements are added to Server and Client,
//
//   - Simplify registering routing with key binding in Server
//   - Add support for handling CORS in Server
//   - Serving files using [memfs.MemFS] in Server
//   - Simplify sending body with "application/x-www-form-urlencoded",
//     "multipart/form-data", "application/json" with POST or PUT methods in
//     Client.
//   - Add support for [HTTP Range] in Server and Client
//   - Add support for [Server-Sent Events] (SSE) in Server.
//     For client see the sub package [sseclient].
//
// # Problems
//
// There are two problems that this library try to handle.
// First, optimizing serving local file system; second, complexity of routing
// regarding to their method, request type, and response type.
//
// Assuming that we want to serve file system and API using [http.ServeMux],
// the simplest registered handler are,
//
//	mux.HandleFunc("/", handleFileSystem)
//	mux.HandleFunc("/api", handleAPI)
//
// The first problem is regarding to [http.ServeFile].
// Everytime the request hit "handleFileSystem" the [http.ServeFile] try to
// locate the file regarding to request path in system, read the content of
// file, parse its content type, and finally write the content-type,
// content-length, and body as response.
// This is time consuming.
// Of course, on modern OS, they may caching readed file descriptor in memory
// to minimize disk lookup, so the next call to the same file path may not
// touch the hard storage back again.
//
// The second problem is regarding to handling API.
// We must check the request method, checking content-type, parsing query
// parameter or POST form in every sub-handler of API.
// Assume that we have an API with method POST and query parameter, the method
// to handle it would be like these,
//
//	handleAPILogin(res, req) {
//		// (1) Check if method is POST
//		// (2) Parse query parameter
//		// (3) Process request
//		// (4) Write response
//	}
//
// The step number 1, 2, 4 needs to be written for every handler of our API.
//
// # Solutions
//
// The solution to the first problem is by mapping all content of files to be
// served into memory.
// This cause more memory to be consumed on server side but we minimize path
// lookup, and cache-miss on OS level.
//
// Serving file system is handled by package [memfs], which can be set on
// [ServerOptions].
// For example, to serve all content in directory "www", we can set the
// [ServerOptions] to,
//
//	opts := &http.ServerOptions{
//		Memfs: &memfs.MemFS{
//			Opts: &memfs.Options{
//				Root:        `./www`,
//				TryDirect:   true,
//			},
//		},
//		Address: ":8080",
//	}
//	httpServer, err := NewServer(opts)
//
// There is a limit on size of file to be mapped on memory.
// See the package "lib/memfs" for more information.
//
// The solution to the second problem is by mapping the registered request per
// method and by path.
// User just need to focus on step 3, handling on how to process request, all
// of process on step 1, 2, and 4 will be handled by our library.
//
//	import (
//		libhttp "github.com/shuLhan/share/lib/http"
//	)
//
//	...
//
//	epAPILogin := &libhttp.Endpoint{
//		Method: libhttp.RequestMethodPost,
//		Path: "/api/login",
//		RequestType: libhttp.RequestTypeQuery,
//		ResponseType: libhttp.ResponseTypeJSON,
//		Call: handleLogin,
//	}
//	server.RegisterEndpoint(epAPILogin)
//
//	...
//
// Upon receiving request to "POST /api/login", the library will call
// [http.HttpRequest.ParseForm], read the content of body and pass them to
// "handleLogin",
//
//	func handleLogin(epr *EndpointRequest) (resBody []byte, err error) {
//		// Process login input from epr.HttpRequest.Form,
//		// epr.HttpRequest.PostForm, and/or epr.RequestBody.
//		// Return response body and error.
//	}
//
// # Routing
//
// The [Endpoint] allow binding the unique key into path using colon ":" as
// the first character.
//
// For example, after registering the following [Endpoint],
//
//	epBinding := &libhttp.Endpoint{
//		Method: libhttp.RequestMethodGet,
//		Path: "/category/:name",
//		RequestType: libhttp.RequestTypeQuery,
//		ResponseType: libhttp.ResponseTypeJSON,
//		Call: handleCategory,
//	}
//	server.RegisterEndpoint(epBinding)
//
// when the server receiving GET request using path "/category/book?limit=10",
// it will put the "book" and "10" into [http.Request.Form] with key is
// "name" and "limit"
//
//	fmt.Println("request.Form:", req.Form)
//	// request.Form: map[name:[book] limit:[10]]
//
// The key binding must be unique between path and query.  If query has the
// same key then it will be overridden by value in path.  For example, using
// the above endpoint, request with "/category/book?name=Hitchiker" will
// result in [http.Request.Form]:
//
//	map[name:[book]]
//
// not
//
//	map[name:[book Hitchiker]]
//
// # Callback error handling
//
// Each [Endpoint] can have their own error handler.
// If its nil, it will default to [DefaultErrorHandler], which return the
// error as JSON with the following format,
//
//	{"code":<HTTP_STATUS_CODE>,"message":<err.Error()>}
//
// # Range request
//
// The standard http package provide [http.ServeContent] function that
// support serving resources with Range request, except that it sometime it
// has an issue.
//
// When server receive,
//
//	GET /big
//	Range: bytes=0-
//
// and the requested resources is quite larger, where writing all content of
// file result in i/o timeout, it is [best] [practice] if the server
// write only partial content and let the client continue with the
// subsequent Range request.
//
// In the above case, the server should response with,
//
//	HTTP/1.1 206 Partial content
//	Content-Range: bytes 0-<limit>/<size>
//	Content-Length: <limit>
//
// Where limit is maximum packet that is [reasonable] for most of the
// client.
// In this server we choose 8MB as limit, see [DefRangeLimit].
//
// # Summary
//
// The pseudocode below illustrate how [Endpoint], [Callback], and
// [CallbackErrorHandler] works when the [Server] receive HTTP request,
//
//	func (server *Server) (w http.ResponseWriter, req *http.Request) {
//		for _, endpoint := range server.endpoints {
//			if endpoint.Method.String() != req.Method {
//				continue
//			}
//
//			epr := &EndpointRequest{
//				Endpoint: endpoint,
//				HttpWriter: w,
//				HttpRequest: req,
//			}
//			epr.RequestBody, _ = io.ReadAll(req.Body)
//
//			// Check request type, and call ParseForm or
//			// ParseMultipartForm if required.
//
//			var resBody []byte
//			resBody, epr.Error = endpoint.Call(epr)
//			if err != nil {
//				endpoint.ErrorHandler(epr)
//				return
//			}
//			// Set content-type based on endpoint.ResponseType,
//			// and write the response body,
//			w.Write(resBody)
//			return
//		}
//		// If request is HTTP GET, check if Path exist as static
//		// contents in Memfs.
//	}
//
// # Bugs and Limitations
//
//   - The server does not handle CONNECT method
//
//   - Missing test for request with content-type multipart-form
//
//   - Server can not register path with ambigous route.  For example, "/:x" and
//     "/y" are ambiguous because one is dynamic path using key binding "x" and
//     the last one is static path to "y".
//
// [best]: https://stackoverflow.com/questions/63614008/how-best-to-respond-to-an-open-http-range-request
// [practice]: https://bugzilla.mozilla.org/show_bug.cgi?id=570755
// [reasonable]: https://docs.aws.amazon.com/whitepapers/latest/s3-optimizing-performance-best-practices/use-byte-range-fetches.html
// [HTTP Range]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests
// [Server-Sent Events]: https://html.spec.whatwg.org/multipage/server-sent-events.html
package http

import (
	"errors"
	"net/http"
	"strings"

	libnet "github.com/shuLhan/share/lib/net"
)

// List of header value for HTTP header Accept-Ranges.
const (
	AcceptRangesBytes = `bytes`
	AcceptRangesNone  = `none`
)

// List of known HTTP header keys and values.
const (
	ContentEncodingBzip2    = "bzip2"
	ContentEncodingCompress = "compress" // Using LZW.
	ContentEncodingGzip     = "gzip"
	ContentEncodingDeflate  = "deflate" // Using zlib.

	ContentTypeBinary              = "application/octet-stream"
	ContentTypeEventStream         = `text/event-stream`
	ContentTypeForm                = "application/x-www-form-urlencoded"
	ContentTypeMultipartByteRanges = `multipart/byteranges`
	ContentTypeMultipartForm       = "multipart/form-data"
	ContentTypeHTML                = "text/html; charset=utf-8"
	ContentTypeJSON                = "application/json"
	ContentTypePlain               = "text/plain; charset=utf-8"
	ContentTypeXML                 = "text/xml; charset=utf-8"

	HeaderACAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderACAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderACAllowMethod      = "Access-Control-Allow-Method"
	HeaderACAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderACExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderACMaxAge           = "Access-Control-Max-Age"
	HeaderACRequestHeaders   = "Access-Control-Request-Headers"
	HeaderACRequestMethod    = "Access-Control-Request-Method"
	HeaderAccept             = `Accept`
	HeaderAcceptEncoding     = "Accept-Encoding"
	HeaderAcceptRanges       = `Accept-Ranges`
	HeaderAllow              = "Allow"
	HeaderAuthKeyBearer      = "Bearer"
	HeaderAuthorization      = "Authorization"
	HeaderCacheControl       = `Cache-Control`
	HeaderContentEncoding    = "Content-Encoding"
	HeaderContentLength      = "Content-Length"
	HeaderContentRange       = `Content-Range`
	HeaderContentType        = "Content-Type"
	HeaderCookie             = "Cookie"
	HeaderDate               = `Date`
	HeaderETag               = "Etag"
	HeaderHost               = "Host"
	HeaderIfNoneMatch        = "If-None-Match"
	HeaderLastEventID        = `Last-Event-ID`
	HeaderLocation           = "Location"
	HeaderOrigin             = "Origin"
	HeaderRange              = `Range`
	HeaderUserAgent          = "User-Agent"
	HeaderXForwardedFor      = "X-Forwarded-For" // https://en.wikipedia.org/wiki/X-Forwarded-For
	HeaderXRealIp            = `X-Real-Ip`       //revive:disable-line
)

var (
	// ErrClientDownloadNoOutput define an error when Client's
	// DownloadRequest does not define the Output.
	ErrClientDownloadNoOutput = errors.New("invalid or empty client download output")

	// ErrEndpointAmbiguous define an error when registering path that
	// already exist.  For example, after registering "/:x", registering
	// "/:y" or "/z" on the same HTTP method will result in ambiguous.
	ErrEndpointAmbiguous = errors.New("ambigous endpoint")

	// ErrEndpointKeyDuplicate define an error when registering path with
	// the same keys, for example "/:x/:x".
	ErrEndpointKeyDuplicate = errors.New("duplicate key in route")

	// ErrEndpointKeyEmpty define an error when path contains an empty
	// key, for example "/:/y".
	ErrEndpointKeyEmpty = errors.New("empty route's key")
)

// IPAddressOfRequest get the client IP address from HTTP request header
// "X-Real-IP" or "X-Forwarded-For", which ever non-empty first.
// If no headers present, use the default address.
func IPAddressOfRequest(headers http.Header, defAddr string) (addr string) {
	addr = headers.Get(HeaderXRealIp)
	if len(addr) == 0 {
		addr, _ = ParseXForwardedFor(headers.Get(HeaderXForwardedFor))
		if len(addr) == 0 {
			addr = defAddr
		}
	}
	addr, _, _ = libnet.ParseIPPort(addr, 0)
	return addr
}

// ParseXForwardedFor parse the HTTP header "X-Forwarded-For" value from the
// following format "client, proxy1, proxy2" into client address and list of
// proxy addressess.
func ParseXForwardedFor(val string) (clientAddr string, proxyAddrs []string) {
	if len(val) == 0 {
		return "", nil
	}
	addrs := strings.Split(val, ",")
	for x, addr := range addrs {
		if x == 0 {
			clientAddr = strings.TrimSpace(addr)
		} else {
			proxyAddrs = append(proxyAddrs, strings.TrimSpace(addr))
		}
	}
	return clientAddr, proxyAddrs
}

// setHeaders set the request headers.
func setHeaders(req *http.Request, headers http.Header) {
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
