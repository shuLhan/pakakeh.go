// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/memfs"
	"github.com/shuLhan/share/lib/mlog"
)

const (
	defRWTimeout = 30 * time.Second
	corsWildcard = "*"
)

// Server define HTTP server.
type Server struct {
	*http.Server

	// Options for server, set by calling NewServer.
	// This field is exported only for reference, for example logging in
	// the Options when server started.
	// Modifying the value of Options after server has been started may
	// cause undefined effects.
	Options *ServerOptions

	evals        []Evaluator
	routeDeletes []*route
	routeGets    []*route
	routePatches []*route
	routePosts   []*route
	routePuts    []*route
}

// NewServer create and initialize new HTTP server that serve root directory
// with custom connection.
func NewServer(opts *ServerOptions) (srv *Server, err error) {
	opts.init()

	srv = &Server{
		Options: opts,
	}

	srv.Server = opts.Conn
	srv.Server.Addr = opts.Address
	srv.Handler = srv

	if srv.ReadTimeout == 0 {
		srv.ReadTimeout = defRWTimeout
	}
	if srv.WriteTimeout == 0 {
		srv.WriteTimeout = defRWTimeout
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

// RegisterEndpoint register the Endpoint based on Method.
// If Method field is not set, it will default to GET.
// The Endpoint.Call field MUST be set, or it will return an error.
//
// Endpoint with method HEAD and OPTIONS does not have any effect because it
// already handled automatically by server.
//
// Endpoint with method CONNECT and TRACE will return an error because its not
// supported yet.
func (srv *Server) RegisterEndpoint(ep *Endpoint) (err error) {
	if ep == nil {
		return nil
	}
	if ep.Call == nil {
		return fmt.Errorf("http.RegisterEndpoint: empty Call field")
	}

	switch ep.Method {
	case RequestMethodConnect:
		return fmt.Errorf("http.RegisterEndpoint: can't handle CONNECT method yet")
	case RequestMethodDelete:
		err = srv.registerDelete(ep)
	case RequestMethodHead:
		return nil
	case RequestMethodOptions:
		return nil
	case RequestMethodPatch:
		err = srv.registerPatch(ep)
	case RequestMethodPost:
		err = srv.registerPost(ep)
	case RequestMethodPut:
		err = srv.registerPut(ep)
	case RequestMethodTrace:
		return fmt.Errorf("http.RegisterEndpoint: can't handle TRACE method yet")
	default:
		ep.Method = RequestMethodGet
		err = srv.registerGet(ep)
	}
	return err
}

// registerDelete register HTTP method DELETE with specific endpoint to handle
// it.
func (srv *Server) registerDelete(ep *Endpoint) (err error) {
	ep.RequestType = RequestTypeQuery

	// Check if the same route already registered.
	for _, rute := range srv.routeDeletes {
		_, ok := rute.parse(ep.Path)
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
// Endpoint evalutor and callback is called.
func (srv *Server) RegisterEvaluator(eval Evaluator) {
	srv.evals = append(srv.evals, eval)
}

// registerGet register HTTP method GET with callback to handle it.
func (srv *Server) registerGet(ep *Endpoint) (err error) {
	ep.RequestType = RequestTypeQuery

	// Check if the same route already registered.
	for _, rute := range srv.routeGets {
		_, ok := rute.parse(ep.Path)
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
		_, ok := rute.parse(ep.Path)
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
		_, ok := rute.parse(ep.Path)
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
		_, ok := rute.parse(ep.Path)
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
	var (
		logp = "ServeHTTP"
	)

	if debug.Value >= 3 {
		mlog.Outf("%s: %s %+v", logp, req.Method, req.URL)
	}

	switch req.Method {
	case http.MethodDelete:
		srv.handleDelete(res, req)

	case http.MethodGet:
		srv.handleCORS(res, req)
		srv.handleGet(res, req)

	case http.MethodHead:
		srv.handleCORS(res, req)
		srv.handleHead(res, req)

	case http.MethodOptions:
		srv.handleOptions(res, req)

	case http.MethodPatch:
		srv.handlePatch(res, req)

	case http.MethodPost:
		srv.handleCORS(res, req)
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
	if srv.TLSConfig == nil {
		err = srv.ListenAndServe()
	} else {
		err = srv.ListenAndServeTLS("", "")
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
	return srv.Shutdown(ctx)
}

func (srv *Server) getFSNode(reqPath string) (node *memfs.Node) {
	var (
		logp = "getFSNode"
		err  error
	)

	if srv.Options.Memfs == nil {
		return nil
	}

	node, err = srv.Options.Memfs.Get(reqPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if debug.Value >= 3 {
				mlog.Outf("%s: %q: %s", logp, reqPath, err)
			}
			return nil
		}

		asDir := path.Join(reqPath, "index.html")

		node, err = srv.Options.Memfs.Get(asDir)
		if err != nil {
			asHtml := reqPath + ".html"
			node, err = srv.Options.Memfs.Get(asHtml)
			if err != nil {
				if debug.Value >= 3 {
					mlog.Outf("%s: %q: %s", logp, reqPath, err)
				}
				return nil
			}
			return node
		}
	}

	if node.IsDir() {
		indexHTML := path.Join(reqPath, "index.html")
		node, err = srv.Options.Memfs.Get(indexHTML)
		if err != nil {
			return nil
		}
	}

	return node
}

// handleCORS handle the CORS request.
//
// Reference: https://www.html5rocks.com/static/images/cors_server_flowchart.png
func (srv *Server) handleCORS(res http.ResponseWriter, req *http.Request) {
	var found bool
	preflightOrigin := req.Header.Get(HeaderOrigin)
	if len(preflightOrigin) == 0 {
		return
	}

	for _, origin := range srv.Options.CORS.AllowOrigins {
		if origin == corsWildcard {
			res.Header().Set(HeaderACAllowOrigin, preflightOrigin)
			found = true
			break
		}
		if origin == preflightOrigin {
			res.Header().Set(HeaderACAllowOrigin, preflightOrigin)
			found = true
			break
		}
	}
	if !found {
		return
	}

	preflightMethod := req.Header.Get(HeaderACRequestMethod)
	if len(preflightMethod) == 0 {
		if len(srv.Options.CORS.exposeHeaders) > 0 {
			res.Header().Set(
				HeaderACExposeHeaders,
				srv.Options.CORS.exposeHeaders,
			)
		}
		return
	}

	switch preflightMethod {
	case http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete:
		res.Header().Set(HeaderACAllowMethod, preflightMethod)
	default:
		return
	}

	srv.handleCORSRequestHeaders(res, req)

	if len(srv.Options.CORS.maxAge) > 0 {
		res.Header().Set(HeaderACMaxAge, srv.Options.CORS.maxAge)
	}
	if srv.Options.CORS.AllowCredentials {
		res.Header().Set(HeaderACAllowCredentials, "true")
	}
}

func (srv *Server) handleCORSRequestHeaders(
	res http.ResponseWriter, req *http.Request,
) {
	preflightHeaders := req.Header.Get(HeaderACRequestHeaders)
	if len(preflightHeaders) == 0 {
		return
	}

	reqHeaders := strings.Split(preflightHeaders, ",")
	for x := 0; x < len(reqHeaders); x++ {
		reqHeaders[x] = strings.ToLower(strings.TrimSpace(reqHeaders[x]))
	}

	allowHeaders := make([]string, 0, len(reqHeaders))

	for _, reqHeader := range reqHeaders {
		for _, allowHeader := range srv.Options.CORS.AllowHeaders {
			if allowHeader == corsWildcard {
				allowHeaders = append(allowHeaders, reqHeader)
				break
			}
			if reqHeader == allowHeader {
				allowHeaders = append(allowHeaders, reqHeader)
				break
			}
		}
	}
	if len(allowHeaders) == 0 {
		return
	}

	res.Header().Set(HeaderACAllowHeaders, strings.Join(allowHeaders, ","))
}

// handleDelete handle the DELETE request by searching the registered route
// and calling the endpoint.
func (srv *Server) handleDelete(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routeDeletes {
		vals, ok := rute.parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}
	res.WriteHeader(http.StatusNotFound)
}

// HandleFS handle the request as resource in the memory file system.
// This method only works if the Server.Options.Memfs is not nil.
//
// If the request Path exists and Server Options FSHandler is set and
// returning false, it will return immediately.
//
// If the request Path exists in file system, it will return 200 OK with the
// header Content-Type set accordingly to the detected file type and the
// response body set to the content of file.
// If the request Method is HEAD, only the header will be sent back to client.
//
// If the request Path is not exist it will return 404 Not Found.
func (srv *Server) HandleFS(res http.ResponseWriter, req *http.Request) {
	var (
		logp = "HandleFS"

		node         *memfs.Node
		responseETag string
		requestETag  string
		body         []byte
		size         int64
		err          error
		ok           bool
	)

	node = srv.getFSNode(req.URL.Path)
	if node == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if srv.Options.HandleFS != nil {
		ok = srv.Options.HandleFS(node, res, req)
		if !ok {
			return
		}
	}

	res.Header().Set(HeaderContentType, node.ContentType)

	responseETag = strconv.FormatInt(node.ModTime().Unix(), 10)
	requestETag = req.Header.Get(HeaderIfNoneMatch)
	if requestETag == responseETag {
		res.WriteHeader(http.StatusNotModified)
		return
	}

	if len(node.Content) > 0 {
		body = node.Content
		size = node.Size()
	} else {
		body, err = os.ReadFile(node.SysPath)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		size = int64(len(body))
	}

	res.Header().Set(HeaderContentLength, strconv.FormatInt(size, 10))
	res.Header().Set(HeaderETag, responseETag)

	if req.Method == http.MethodHead {
		res.WriteHeader(http.StatusOK)
		return
	}

	res.WriteHeader(http.StatusOK)
	_, err = res.Write(body)
	if err != nil {
		mlog.Errf("%s: %s %s: %s", logp, req.Method, req.URL.Path, err)
	}
}

// handleGet handle the GET request by searching the registered route and
// calling the endpoint.
func (srv *Server) handleGet(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routeGets {
		vals, ok := rute.parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}

	srv.HandleFS(res, req)
}

func (srv *Server) handleHead(res http.ResponseWriter, req *http.Request) {
	var (
		rute *route
		ok   bool
	)

	for _, rute = range srv.routeGets {
		_, ok = rute.parse(req.URL.Path)
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

	node := srv.getFSNode(req.URL.Path)
	if node != nil {
		methods[http.MethodGet] = true
		methods[http.MethodHead] = true
	}

	for _, rute := range srv.routeDeletes {
		_, ok := rute.parse(req.URL.Path)
		if ok {
			methods[http.MethodDelete] = true
			break
		}
	}

	for _, rute := range srv.routeGets {
		_, ok := rute.parse(req.URL.Path)
		if ok {
			methods[http.MethodGet] = true
			break
		}
	}

	for _, rute := range srv.routePatches {
		_, ok := rute.parse(req.URL.Path)
		if ok {
			methods[http.MethodPatch] = true
			break
		}
	}

	for _, rute := range srv.routePosts {
		_, ok := rute.parse(req.URL.Path)
		if ok {
			methods[http.MethodPost] = true
			break
		}
	}

	for _, rute := range srv.routePuts {
		_, ok := rute.parse(req.URL.Path)
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

	srv.handleCORS(res, req)

	res.WriteHeader(http.StatusOK)
}

// handlePatch handle the PATCH request by searching the registered route and
// calling the endpoint.
func (srv *Server) handlePatch(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routePatches {
		vals, ok := rute.parse(req.URL.Path)
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
		vals, ok := rute.parse(req.URL.Path)
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
		vals, ok := rute.parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}
	res.WriteHeader(http.StatusNotFound)
}
