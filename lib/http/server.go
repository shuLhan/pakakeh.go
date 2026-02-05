// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mlog"
)

const (
	defRWTimeout = 30 * time.Second
	corsWildcard = "*"
)

// Server define HTTP server.
type Server struct {
	*http.Server

	evals        []Evaluator
	routeDeletes []*route
	routeGets    []*route
	routePatches []*route
	routePosts   []*route
	routePuts    []*route

	// Options for server, set by calling NewServer.
	// This field is exported only for reference, for example logging in
	// the Options when server started.
	// Modifying the value of Options after server has been started may
	// cause undefined effects.
	Options ServerOptions
}

// NewServer create and initialize new HTTP server that serve root directory
// with custom connection.
func NewServer(opts ServerOptions) (srv *Server, err error) {
	opts.init()

	srv = &Server{
		Options: opts,
	}

	srv.Server = opts.Conn
	srv.Server.Addr = opts.Address
	srv.Server.Handler = srv

	if srv.Server.ReadTimeout == 0 {
		srv.Server.ReadTimeout = defRWTimeout
	}
	if srv.Server.WriteTimeout == 0 {
		srv.Server.WriteTimeout = defRWTimeout
	}
	if srv.Options.Memfs != nil {
		err = srv.Options.Memfs.Init()
		if err != nil {
			return nil, fmt.Errorf("NewServer: %w", err)
		}
	}

	return srv, nil
}

// RedirectTemp make the request to temporary redirect (307) to new URL.
func (srv *Server) RedirectTemp(res http.ResponseWriter, redirectURL string) {
	if len(redirectURL) == 0 {
		redirectURL = "/"
	}
	res.Header().Set(HeaderLocation, redirectURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// RegisterEndpoint register the [Endpoint] based on Method.
// If [Endpoint.Method] field is not set, it will default to GET.
// The [Endpoint.Call] field MUST be set, or it will return an error.
//
// Endpoint with Method HEAD or OPTIONS does not have any effect because it
// already handled automatically by server.
//
// Endpoint with Method CONNECT or TRACE will return an error because its
// not supported, yet.
func (srv *Server) RegisterEndpoint(ep Endpoint) (err error) {
	var logp = `RegisterEndpoint`

	if ep.Call == nil {
		return errors.New(logp + `: empty Call field`)
	}
	if len(ep.Method) == 0 {
		ep.Method = RequestMethodGet
	}

	switch ep.Method {
	case RequestMethodConnect:
		return errors.New(logp + `: can't handle CONNECT method yet`)
	case RequestMethodDelete:
		err = srv.registerDelete(&ep)
	case RequestMethodHead:
		return nil
	case RequestMethodOptions:
		return nil
	case RequestMethodPatch:
		err = srv.registerPatch(&ep)
	case RequestMethodPost:
		err = srv.registerPost(&ep)
	case RequestMethodPut:
		err = srv.registerPut(&ep)
	case RequestMethodTrace:
		return errors.New(logp + `: can't handle TRACE method yet`)
	case RequestMethodGet:
		err = srv.registerGet(&ep)
	}
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// RegisterHandlerFunc register a pattern with a handler, similar to
// [http.ServeMux.HandleFunc].
// The pattern follow the Go 1.22 format:
//
//	[METHOD] PATH
//
// The METHOD is optional, default to GET.
// The PATH must not contains the domain name and space.
// Unlike standard library, variable in PATH is read using ":var" not
// "{var}".
// This endpoint will accept any content type and return the body as is;
// it is up to the handler to read and set the content type and the response
// headers.
//
// If the METHOD and/or PATH is already registered it will panic.
func (srv *Server) RegisterHandleFunc(
	pattern string,
	handler func(http.ResponseWriter, *http.Request),
) {
	var (
		logp       = `RegisterHandleFunc`
		methodPath = strings.Fields(pattern)
		ep         = Endpoint{
			Call: func(epr *EndpointRequest) (resp []byte, err error) {
				handler(epr.HTTPWriter, epr.HTTPRequest)
				return nil, nil
			},
		}
	)
	if len(methodPath) == 1 {
		ep.Path = methodPath[0]
	} else if len(methodPath) > 1 {
		ep.Method = RequestMethod(strings.ToUpper(methodPath[0]))
		ep.Path = methodPath[1]
	}
	var err = srv.RegisterEndpoint(ep)
	if err != nil {
		log.Panicf(`%s: %s %q`, logp, err, ep.Path)
	}
}

// RegisterSSE register Server-Sent Events endpoint.
// It will return an error if the [SSEEndpoint.Call] field is not set or
// [ErrEndpointAmbiguous] if the same path is already registered.
func (srv *Server) RegisterSSE(ep SSEEndpoint) (err error) {
	var logp = `RegisterSSE`

	if ep.Call == nil {
		return fmt.Errorf(`%s: Call field not set`, logp)
	}

	// Check if the same GET path already registered.
	var (
		rute  *route
		exist bool
	)
	for _, rute = range srv.routeGets {
		_, exist = rute.Parse(ep.Path)
		if exist {
			return fmt.Errorf(`%s: %w`, logp, ErrEndpointAmbiguous)
		}
	}

	rute, err = newRouteSSE(&ep)
	if err != nil {
		return err
	}

	srv.routeGets = append(srv.routeGets, rute)

	return nil
}

// registerDelete register HTTP method DELETE with specific endpoint to handle
// it.
func (srv *Server) registerDelete(ep *Endpoint) (err error) {
	ep.RequestType = RequestTypeQuery

	// Check if the same route already registered.
	for _, rute := range srv.routeDeletes {
		_, ok := rute.Parse(ep.Path)
		if ok {
			return ErrEndpointAmbiguous
		}
	}

	rute, err := newRoute(ep)
	if err != nil {
		return err
	}

	srv.routeDeletes = append(srv.routeDeletes, rute)

	return nil
}

// RegisterEvaluator register HTTP middleware that will be called before
// [Endpoint.Eval] and [Endpoint.Call] is called.
func (srv *Server) RegisterEvaluator(eval Evaluator) {
	srv.evals = append(srv.evals, eval)
}

// registerGet register HTTP method GET with callback to handle it.
func (srv *Server) registerGet(ep *Endpoint) (err error) {
	ep.Method = RequestMethodGet
	ep.RequestType = RequestTypeQuery

	// Check if the same route already registered.
	for _, rute := range srv.routeGets {
		_, ok := rute.Parse(ep.Path)
		if ok {
			return ErrEndpointAmbiguous
		}
	}

	rute, err := newRoute(ep)
	if err != nil {
		return err
	}

	srv.routeGets = append(srv.routeGets, rute)

	return nil
}

// registerPatch register HTTP method PATCH with callback to handle it.
func (srv *Server) registerPatch(ep *Endpoint) (err error) {
	// Check if the same route already registered.
	for _, rute := range srv.routePatches {
		_, ok := rute.Parse(ep.Path)
		if ok {
			return ErrEndpointAmbiguous
		}
	}

	rute, err := newRoute(ep)
	if err != nil {
		return err
	}

	srv.routePatches = append(srv.routePatches, rute)

	return nil
}

// registerPost register HTTP method POST with callback to handle it.
func (srv *Server) registerPost(ep *Endpoint) (err error) {
	// Check if the same route already registered.
	for _, rute := range srv.routePosts {
		_, ok := rute.Parse(ep.Path)
		if ok {
			return ErrEndpointAmbiguous
		}
	}

	rute, err := newRoute(ep)
	if err != nil {
		return err
	}

	srv.routePosts = append(srv.routePosts, rute)

	return nil
}

// registerPut register HTTP method PUT with callback to handle it.
func (srv *Server) registerPut(ep *Endpoint) (err error) {
	// Check if the same route already registered.
	for _, rute := range srv.routePuts {
		_, ok := rute.Parse(ep.Path)
		if ok {
			return ErrEndpointAmbiguous
		}
	}

	rute, err := newRoute(ep)
	if err != nil {
		return err
	}

	srv.routePuts = append(srv.routePuts, rute)

	return nil
}

// ServeHTTP handle mapping of client request to registered endpoints.
func (srv *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		srv.handleDelete(res, req)

	case http.MethodGet:
		srv.Options.CORS.handle(res, req)
		srv.handleGet(res, req)

	case http.MethodHead:
		srv.Options.CORS.handle(res, req)
		srv.handleHead(res, req)

	case http.MethodOptions:
		srv.handleOptions(res, req)

	case http.MethodPatch:
		srv.handlePatch(res, req)

	case http.MethodPost:
		srv.Options.CORS.handle(res, req)
		srv.handlePost(res, req)

	case http.MethodPut:
		srv.handlePut(res, req)

	default:
		res.WriteHeader(http.StatusNotImplemented)
		return
	}
}

// Start the HTTP server.
func (srv *Server) Start() (err error) {
	if srv.Server.TLSConfig == nil {
		if srv.Options.Listener == nil {
			err = srv.Server.ListenAndServe()
		} else {
			err = srv.Server.Serve(srv.Options.Listener)
		}
	} else {
		if srv.Options.Listener == nil {
			err = srv.Server.ListenAndServeTLS(``, ``)
		} else {
			err = srv.Server.ServeTLS(srv.Options.Listener, ``, ``)
		}
	}
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

// Stop the server using Shutdown method. The wait is set default and minimum
// to five seconds.
func (srv *Server) Stop(wait time.Duration) (err error) {
	var defWait = 5 * time.Second
	if wait <= defWait {
		wait = defWait
	}
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	return srv.Server.Shutdown(ctx)
}

// getFSNode get the memfs Node based on the request path.
//
// If the path is not exist, try path with ".html".
//
// If the path is directory and contains index.html, the node content for
// index.html will be returned.
//
// If the path is directory and does not contains index.html and
// [ServerOptions.EnableIndexHTML] is true, server will generate list of
// content for index.html.
//
// The redirectPath return whether the server should redirect first before
// serving the content to provide consistent canonical path that end with "/"
// for directory that contains "index.html" or EnableIndexHTML is on.
func (srv *Server) getFSNode(reqURL url.URL) (node *memfs.Node, redirectURL *url.URL) {
	if srv.Options.Memfs == nil {
		return nil, nil
	}

	reqPath := reqURL.Path
	endWithSlash := reqPath[len(reqPath)-1] == '/'
	var err error

	// The [MemFS.Get] method check reqPath with trailing `/` removed, so
	// "path" or "path/" will return the same node.
	node, err = srv.Options.Memfs.Get(reqPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if !endWithSlash {
			pathHTML := reqPath + `.html`
			node, err = srv.Options.Memfs.Get(pathHTML)
			if err != nil {
				return nil, nil
			}
			return node, nil
		}
		return nil, nil
	}
	if !node.IsDir() {
		return node, nil
	}

	pathIndexHTML := path.Join(reqPath, `index.html`)
	nodeIndexHTML, err := srv.Options.Memfs.Get(pathIndexHTML)
	if err != nil {
		if srv.Options.EnableIndexHTML {
			if !endWithSlash {
				redirectURL = reqURL.JoinPath(`/`)
				return nil, redirectURL
			}
			node.GenerateIndexHTML()
			return node, nil
		}
		return nil, nil
	}
	if !endWithSlash {
		redirectURL = reqURL.JoinPath(`/`)
		return nil, redirectURL
	}

	return nodeIndexHTML, nil
}

// handleDelete handle the DELETE request by searching the registered route
// and calling the endpoint.
func (srv *Server) handleDelete(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routeDeletes {
		vals, ok := rute.Parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}
	res.WriteHeader(http.StatusNotFound)
}

// HandleFS handle the request as resource in the memory file system.
// This method only works if the [ServerOptions.Memfs] is not nil.
//
// If the request Path exists and [ServerOptions.HandleFS] is set and
// returning [*memfs.Node] with non-zero HTTP `statusCode`, server will return
// immediately.
//
// If the request Path exists in file system, it will return 200 OK with the
// header Content-Type set accordingly to the detected file type and the
// response body set to the content of file.
//
// If the request Method is HEAD, only the header will be sent back to client.
//
// If the request Path is not exist it will return 404 Not Found.
func (srv *Server) HandleFS(res http.ResponseWriter, req *http.Request) {
	node, redirectURL := srv.getFSNode(*req.URL)
	if redirectURL != nil {
		// If request path is a directory and it is not end with
		// slash, redirect request to location with slash to allow
		// relative links works inside the HTML content.
		http.Redirect(res, req, redirectURL.String(), http.StatusFound)
		return
	}
	if srv.Options.HandleFS != nil {
		var statusCode int
		node, statusCode = srv.Options.HandleFS(node, res, req)
		if statusCode != 0 {
			// HandleFS already response to the request.
			return
		}
	}
	if node == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set(HeaderContentType, node.ContentType)

	var nodeModtime = node.ModTime().Unix()

	var (
		responseETag = strconv.FormatInt(nodeModtime, 10)
		requestETag  = req.Header.Get(HeaderIfNoneMatch)
	)
	if !node.IsDir() && requestETag == responseETag {
		res.WriteHeader(http.StatusNotModified)
		return
	}

	var ifModifiedSince = req.Header.Get(HeaderIfModifiedSince)
	if len(ifModifiedSince) != 0 {
		timeModsince, err := time.Parse(time.RFC1123, ifModifiedSince)
		if err == nil {
			if nodeModtime <= timeModsince.Unix() {
				res.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	var (
		bodyReader io.ReadSeeker
		size       int64
	)

	if len(node.Content) > 0 {
		bodyReader = bytes.NewReader(node.Content)
		size = node.Size()
	} else {
		f, err := os.Open(node.SysPath)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		var fstat os.FileInfo
		fstat, err = f.Stat()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		bodyReader = f
		size = fstat.Size()
	}

	res.Header().Set(HeaderETag, responseETag)

	if req.Method == http.MethodHead {
		var sizeStr = strconv.FormatInt(size, 10)
		res.Header().Set(HeaderContentLength, sizeStr)
		res.Header().Set(HeaderAcceptRanges, AcceptRangesBytes)
		res.WriteHeader(http.StatusOK)
		return
	}

	var reqRange = req.Header.Get(HeaderRange)
	if len(reqRange) != 0 {
		handleRange(res, req, bodyReader, ``, reqRange)
		return
	}

	responseWrite(`HandleFS`, res, req, bodyReader)
}

// handleGet handle the GET request by searching the registered route and
// calling the endpoint.
func (srv *Server) handleGet(res http.ResponseWriter, req *http.Request) {
	var (
		rute *route
		vals map[string]string
		ok   bool
	)
	for _, rute = range srv.routeGets {
		vals, ok = rute.Parse(req.URL.Path)
		if !ok {
			continue
		}
		if rute.kind == routeKindHTTP {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
		if rute.kind == routeKindSSE {
			rute.endpointSSE.call(res, req, srv.evals, vals)
			return
		}
		// Unknown kind will be handled by HandleFS.
	}

	srv.HandleFS(res, req)
}

// handleHead handle HTTP method [HEAD] request.
// The HEAD request only applicable to GET endpoint.
//
// [HEAD]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD
func (srv *Server) handleHead(res http.ResponseWriter, req *http.Request) {
	var (
		rute *route
		ok   bool
	)

	for _, rute = range srv.routeGets {
		_, ok = rute.Parse(req.URL.Path)
		if ok {
			break
		}
	}
	if !ok {
		srv.HandleFS(res, req)
		return
	}

	switch rute.endpoint.ResponseType {
	case ResponseTypeNone:
		res.WriteHeader(http.StatusNoContent)
		return
	case ResponseTypeBinary:
		res.Header().Set(HeaderContentType, ContentTypeBinary)
	case ResponseTypeJSON:
		res.Header().Set(HeaderContentType, ContentTypeJSON)
	case ResponseTypeHTML:
		res.Header().Set(HeaderContentType, ContentTypeHTML)
	case ResponseTypePlain:
		res.Header().Set(HeaderContentType, ContentTypePlain)
	case ResponseTypeXML:
		res.Header().Set(HeaderContentType, ContentTypeXML)
	}

	res.WriteHeader(http.StatusOK)
}

// handleOptions return list of allowed methods on requested path in HTTP
// response header "Allow".
// If no path found, it will return 404.
func (srv *Server) handleOptions(res http.ResponseWriter, req *http.Request) {
	methods := make(map[string]bool)

	node, redirectURL := srv.getFSNode(*req.URL)
	if node != nil || redirectURL != nil {
		methods[http.MethodGet] = true
		methods[http.MethodHead] = true
	}

	for _, rute := range srv.routeDeletes {
		_, ok := rute.Parse(req.URL.Path)
		if ok {
			methods[http.MethodDelete] = true
			break
		}
	}

	for _, rute := range srv.routeGets {
		_, ok := rute.Parse(req.URL.Path)
		if ok {
			methods[http.MethodGet] = true
			break
		}
	}

	for _, rute := range srv.routePatches {
		_, ok := rute.Parse(req.URL.Path)
		if ok {
			methods[http.MethodPatch] = true
			break
		}
	}

	for _, rute := range srv.routePosts {
		_, ok := rute.Parse(req.URL.Path)
		if ok {
			methods[http.MethodPost] = true
			break
		}
	}

	for _, rute := range srv.routePuts {
		_, ok := rute.Parse(req.URL.Path)
		if ok {
			methods[http.MethodPut] = true
			break
		}
	}

	// The OPTIONS method request to non existen path.
	if len(methods) == 0 {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	methods[http.MethodOptions] = true

	var x int
	allows := make([]string, len(methods))
	for k, v := range methods {
		if v {
			allows[x] = k
			x++
		}
	}

	sort.Strings(allows)

	res.Header().Set(HeaderAllow, strings.Join(allows, ", "))

	srv.Options.CORS.handle(res, req)

	res.WriteHeader(http.StatusOK)
}

// handlePatch handle the PATCH request by searching the registered route and
// calling the endpoint.
func (srv *Server) handlePatch(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routePatches {
		vals, ok := rute.Parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}
	res.WriteHeader(http.StatusNotFound)
}

// handlePost handle the POST request by searching the registered route and
// calling the endpoint.
func (srv *Server) handlePost(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routePosts {
		vals, ok := rute.Parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}
	res.WriteHeader(http.StatusNotFound)
}

// handlePut handle the PUT request by searching the registered route and
// calling the endpoint.
func (srv *Server) handlePut(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routePuts {
		vals, ok := rute.Parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}
	res.WriteHeader(http.StatusNotFound)
}

// HandleRange handle [HTTP Range] request using "bytes" unit.
//
// The body parameter contains the content of resource being requested that
// implement Reader and Seeker.
//
// If the [http.Request.Method] is not GET, or no "Range" in
// [http.Request.Header], it will return all the body [RFC7233 S-3.1].
//
// The contentType is optional, if its empty, it will detected by
// [http.ResponseWriter] during Write.
//
// It will return HTTP Code,
//   - 406 [http.StatusNotAcceptable], if the [Range] unit is not "bytes".
//   - 416 [http.StatusRequestedRangeNotSatisfiable], if the request Range
//     start position is greater than resource size.
//
// [HTTP Range]: https://datatracker.ietf.org/doc/html/rfc7233
// [RFC7233 S-3.1]: https://datatracker.ietf.org/doc/html/rfc7233#section-3.1
func HandleRange(res http.ResponseWriter, req *http.Request, bodyReader io.ReadSeeker, contentType string) {
	var (
		logp     = `HandleRange`
		reqRange = req.Header.Get(HeaderRange)
	)

	if req.Method != http.MethodGet || len(reqRange) == 0 {
		if len(contentType) > 0 {
			res.Header().Set(HeaderContentType, contentType)
		}
		responseWrite(logp, res, req, bodyReader)
		return
	}

	handleRange(res, req, bodyReader, contentType, reqRange)
}

func handleRange(res http.ResponseWriter, req *http.Request, bodyReader io.ReadSeeker, contentType, reqRange string) {
	var (
		logp = `handleRange`
		r    = ParseRange(reqRange)
	)
	if r.IsEmpty() {
		// No range specified, write the full body.
		responseWrite(logp, res, req, bodyReader)
		return
	}
	if r.unit != AcceptRangesBytes {
		res.WriteHeader(http.StatusNotAcceptable)
		return
	}

	if len(contentType) == 0 {
		contentType = rangeContentType(bodyReader)
	}

	var (
		size  int64
		nread int64
		err   error
	)

	size, err = bodyReader.Seek(0, io.SeekEnd)
	if err != nil {
		// An error here assume that the size is unknown ('*').
		log.Printf(`%s: seek body size: %s`, logp, err)
		res.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	var (
		header   = res.Header()
		listBody = make([][]byte, 0, len(r.positions))

		pos *RangePosition
	)
	for _, pos = range r.positions {
		// Refill the position if its nil, for response later,
		// calculate the number of bytes to read, and move the file
		// position for read.
		if pos.start == nil {
			pos.start = new(int64)
			if *pos.end > size {
				*pos.start = 0
			} else {
				*pos.start = size - *pos.end
			}
			*pos.end = size - 1
		} else if pos.end == nil {
			if *pos.start > size {
				// rfc7233#section-4.4
				// the first-byte-pos of all of the
				// byte-range-spec values were greater than
				// the current length of the selected
				// representation.
				pos.start = nil
				header.Set(HeaderContentRange, pos.ContentRange(r.unit, size))
				res.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			pos.end = new(int64)
			*pos.end = size - 1
		}

		_, err = bodyReader.Seek(*pos.start, io.SeekStart)
		if err != nil {
			log.Printf(`%s: seek %s: %s`, logp, pos, err)
			res.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}

		nread = (*pos.end - *pos.start) + 1

		if nread > DefRangeLimit {
			nread = DefRangeLimit
			*pos.end = *pos.start + nread
		}

		var (
			body = make([]byte, nread)
			n    int
		)

		n, err = bodyReader.Read(body)
		if n == 0 || err != nil {
			log.Printf(`%s: range %s/%d: %s`, logp, pos, size, err)
			header.Set(HeaderContentRange, pos.ContentRange(r.unit, size))
			res.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		body = body[:n]
		listBody = append(listBody, body)
	}

	if len(listBody) == 1 {
		var (
			body  = listBody[0]
			nbody = strconv.FormatInt(int64(len(body)), 10)
		)
		pos = r.positions[0]
		header.Set(HeaderContentLength, nbody)
		header.Set(HeaderContentRange, pos.ContentRange(r.unit, size))
		header.Set(HeaderContentType, contentType)
		res.WriteHeader(http.StatusPartialContent)
		_, err = res.Write(listBody[0])
		if err != nil {
			mlog.Errf(`%s: %s`, logp, err)
		}
		return
	}

	var (
		boundary = ascii.Random([]byte(ascii.Hexaletters), 16)

		bb bytes.Buffer
		x  int
	)

	for x, pos = range r.positions {
		fmt.Fprintf(&bb, "--%s\r\n", boundary)
		fmt.Fprintf(&bb, "%s: %s\r\n", HeaderContentType, contentType)
		fmt.Fprintf(&bb, "%s: %s\r\n\r\n", HeaderContentRange, pos.ContentRange(r.unit, size))
		bb.Write(listBody[x])
		bb.WriteString("\r\n")
	}
	fmt.Fprintf(&bb, "--%s--\r\n", boundary)

	var v = fmt.Sprintf(`%s; boundary=%s`, ContentTypeMultipartByteRanges, boundary)
	header.Set(HeaderContentType, v)

	v = strconv.FormatInt(int64(bb.Len()), 10)
	header.Set(HeaderContentLength, v)

	res.WriteHeader(http.StatusPartialContent)
	_, err = res.Write(bb.Bytes())
	if err != nil {
		mlog.Errf(`%s: %s`, logp, err)
	}
}

// rangeContentType detect the body content type for range reply.
func rangeContentType(bodyReader io.ReadSeeker) (contentType string) {
	var (
		part = make([]byte, 512)
		err  error
	)
	_, err = bodyReader.Read(part)
	if err != nil {
		return ContentTypeBinary
	}
	contentType = http.DetectContentType(part)
	return contentType
}

func responseWrite(logp string, res http.ResponseWriter, req *http.Request, bodyReader io.ReadSeeker) {
	var (
		body []byte
		err  error
	)

	body, err = io.ReadAll(bodyReader)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = res.Write(body)
	if err != nil {
		mlog.Errf(`%s: %s %s: %s`, logp, req.Method, req.URL.Path, err)
	}
}
