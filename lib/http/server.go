// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/memfs"
)

const (
	defRWTimeout = 30 * time.Second
)

//
// Server define HTTP server.
//
type Server struct {
	// Memfs contains the content of file systems to be served in memory.
	// It will be initialized only if ServerOptions's Root is not empty or
	// if the current directory contains generated Go file from
	// memfs.GoGenerate.
	Memfs *memfs.MemFS

	evals        []Evaluator
	conn         *http.Server
	routeDeletes []*route
	routeGets    []*route
	routePatches []*route
	routePosts   []*route
	routePuts    []*route
}

//
// NewServer create and initialize new HTTP server that serve root directory
// with custom connection.
//
func NewServer(opts *ServerOptions) (srv *Server, err error) {
	srv = &Server{}

	if len(opts.Address) == 0 {
		opts.Address = ":80"
	}

	if opts.Conn == nil {
		srv.conn = &http.Server{
			ReadTimeout:    defRWTimeout,
			WriteTimeout:   defRWTimeout,
			MaxHeaderBytes: 1 << 20,
		}
	} else {
		srv.conn = opts.Conn
	}

	srv.conn.Addr = opts.Address
	srv.conn.Handler = srv

	if srv.conn.ReadTimeout == 0 {
		srv.conn.ReadTimeout = defRWTimeout
	}
	if srv.conn.WriteTimeout == 0 {
		srv.conn.WriteTimeout = defRWTimeout
	}

	memfs.Development = opts.Development

	if len(opts.Root) > 0 {
		srv.Memfs, err = memfs.New(opts.Includes, opts.Excludes, true)
		if err != nil {
			return nil, err
		}

		err = srv.Memfs.Mount(opts.Root)
		if err != nil {
			return nil, err
		}
	}

	return srv, nil
}

//
// RedirectTemp make the request to temporary redirect (307) to new URL.
//
func (srv *Server) RedirectTemp(res http.ResponseWriter, redirectURL string) {
	if len(redirectURL) == 0 {
		redirectURL = "/"
	}
	res.Header().Set(HeaderLocation, redirectURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

//
// RegisterEndpoint register the Endpoint based on Method.
// If Method field is not set, it will default to GET.
// The Endpoint.Call field MUST be set, or it will return an error.
//
// Endpoint with method HEAD and OPTIONS does not have any effect because it
// already handled automatically by server.
//
// Endpoint with method CONNECT and TRACE will return an error because its not
// supported yet.
//
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

//
// registerDelete register HTTP method DELETE with specific endpoint to handle
// it.
//
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

//
// RegisterEvaluator register HTTP middleware that will be called before
// Endpoint evalutor and callback is called.
//
func (srv *Server) RegisterEvaluator(eval Evaluator) {
	srv.evals = append(srv.evals, eval)
}

//
// registerGet register HTTP method GET with callback to handle it.
//
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

//
// registerPatch register HTTP method PATCH with callback to handle it.
//
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

//
// registerPost register HTTP method POST with callback to handle it.
//
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

//
// registerPut register HTTP method PUT with callback to handle it.
//
func (srv *Server) registerPut(ep *Endpoint) (err error) {
	ep.ResponseType = ResponseTypeNone

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

//
// ServeHTTP handle mapping of client request to registered endpoints.
//
func (srv *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if debug.Value >= 2 {
		log.Printf("> ServeHTTP: %s %+v\n", req.Method, req.URL)
	}

	switch req.Method {
	case http.MethodDelete:
		srv.handleDelete(res, req)

	case http.MethodGet:
		srv.handleGet(res, req)

	case http.MethodHead:
		srv.handleHead(res, req)

	case http.MethodOptions:
		srv.handleOptions(res, req)

	case http.MethodPatch:
		srv.handlePatch(res, req)

	case http.MethodPost:
		srv.handlePost(res, req)

	case http.MethodPut:
		srv.handlePut(res, req)

	default:
		res.WriteHeader(http.StatusNotImplemented)
		return
	}
}

//
// Start the HTTP server.
//
func (srv *Server) Start() (err error) {
	if srv.conn.TLSConfig == nil {
		err = srv.conn.ListenAndServe()
	} else {
		err = srv.conn.ListenAndServeTLS("", "")
	}
	return err
}

func (srv *Server) getFSNode(reqPath string) (node *memfs.Node) {
	if srv.Memfs == nil {
		return nil
	}

	var e error

	node, e = srv.Memfs.Get(reqPath)
	if e != nil {
		if e != os.ErrNotExist {
			log.Printf("http: getFSNode %q: %s", reqPath, e.Error())
			return nil
		}

		reqPath = path.Join(reqPath, "index.html")

		node, e = srv.Memfs.Get(reqPath)
		if e != nil {
			log.Printf("http: getFSNode %q: %s", reqPath, e.Error())
			return nil
		}
	}

	if node.IsDir() {
		indexHTML := path.Join(reqPath, "index.html")
		node, e = srv.Memfs.Get(indexHTML)
		if e != nil {
			return nil
		}
	}

	return node
}

//
// handleDelete handle the DELETE request by searching the registered route
// and calling the endpoint.
//
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

func (srv *Server) handleFS(
	res http.ResponseWriter, req *http.Request, method RequestMethod,
) {
	node := srv.getFSNode(req.URL.Path)
	if node == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set(ContentType, node.ContentType)

	if len(node.ContentEncoding) > 0 {
		res.Header().Set(ContentEncoding, node.ContentEncoding)
	}

	var (
		body []byte
		size int64
		err  error
	)

	if len(node.V) > 0 {
		body = node.V
		size = node.Size()
	} else {
		body, err = ioutil.ReadFile(node.SysPath)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		size = int64(len(body))
	}

	res.Header().Set(ContentLength, strconv.FormatInt(size, 10))

	if method == RequestMethodHead {
		res.WriteHeader(http.StatusOK)
		return
	}

	res.WriteHeader(http.StatusOK)
	_, err = res.Write(body)
	if err != nil {
		log.Println("handleFS: ", err.Error())
	}
}

//
// handleGet handle the GET request by searching the registered route and
// calling the endpoint.
//
func (srv *Server) handleGet(res http.ResponseWriter, req *http.Request) {
	for _, rute := range srv.routeGets {
		vals, ok := rute.parse(req.URL.Path)
		if ok {
			rute.endpoint.call(res, req, srv.evals, vals)
			return
		}
	}

	srv.handleFS(res, req, RequestMethodGet)
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
		srv.handleFS(res, req, RequestMethodHead)
		return
	}

	switch rute.endpoint.ResponseType {
	case ResponseTypeNone:
		res.WriteHeader(http.StatusNoContent)
		return
	case ResponseTypeBinary:
		res.Header().Set(ContentType, ContentTypeBinary)
	case ResponseTypeJSON:
		res.Header().Set(ContentType, ContentTypeJSON)
	case ResponseTypePlain:
		res.Header().Set(ContentType, ContentTypePlain)
	}

	res.WriteHeader(http.StatusOK)
}

//
// handleOptions return list of allowed methods on requested path in HTTP
// response header "Allow".
// If no path found, it will return 404.
//
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

	res.Header().Set("Allow", strings.Join(allows, ", "))
	res.WriteHeader(http.StatusOK)
}

//
// handlePatch handle the PATCH request by searching the registered route and
// calling the endpoint.
//
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

//
// handlePost handle the POST request by searching the registered route and
// calling the endpoint.
//
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

//
// handlePut handle the PUT request by searching the registered route and
// calling the endpoint.
//
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
