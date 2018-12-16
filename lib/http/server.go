// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/memfs"
)

const (
	contentLength     = "Content-Length"
	contentType       = "Content-Type"
	contentTypeBinary = "application/octet-stream"
	contentTypeForm   = "application/x-www-form-urlencoded"
	contentTypeJSON   = "application/json"
	contentTypePlain  = "text/plain"
)

//
// Server define HTTP server.
//
type Server struct {
	mfs       *memfs.MemFS
	conn      *http.Server
	regDelete map[string]*handler
	regGet    map[string]*handler
	regPatch  map[string]*handler
	regPost   map[string]*handler
	regPut    map[string]*handler
}

//
// NewServer create and initialize new HTTP server that serve root directory
// with custom connection.
//
func NewServer(root string, conn *http.Server) (srv *Server, e error) {
	srv = &Server{
		regDelete: make(map[string]*handler),
		regGet:    make(map[string]*handler),
		regPatch:  make(map[string]*handler),
		regPost:   make(map[string]*handler),
		regPut:    make(map[string]*handler),
	}

	if conn == nil {
		srv.conn = &http.Server{
			Addr:           ":80",
			Handler:        srv,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
	} else {
		conn.Handler = srv
		srv.conn = conn
	}

	srv.mfs, e = memfs.New(nil, nil)
	if e != nil {
		return nil, e
	}

	e = srv.mfs.Mount(root)
	if e != nil {
		return nil, e
	}

	return srv, nil
}

//
// RegisterDelete register HTTP method DELETE with callback to handle it.
//
func (srv *Server) RegisterDelete(
	reqPath string, resType ResponseType, cb Callback,
) {
	srv.register(reqPath, RequestMethodDelete, RequestTypeQuery, resType, cb)
}

//
// RegisterGet register HTTP method GET with callback to handle it.
//
func (srv *Server) RegisterGet(
	reqPath string, resType ResponseType, cb Callback,
) {
	srv.register(reqPath, RequestMethodGet, RequestTypeQuery, resType, cb)
}

//
// RegisterPatch register HTTP method PATCH with callback to handle it.
//
func (srv *Server) RegisterPatch(
	reqPath string, reqType RequestType, resType ResponseType, cb Callback,
) {
	srv.register(reqPath, RequestMethodPatch, reqType, resType, cb)
}

//
// RegisterPost register HTTP method POST with callback to handle it.
//
func (srv *Server) RegisterPost(
	reqPath string, reqType RequestType, resType ResponseType, cb Callback,
) {
	srv.register(reqPath, RequestMethodPost, reqType, resType, cb)
}

//
// RegisterPut register HTTP method PUT with callback to handle it.
//
func (srv *Server) RegisterPut(
	reqPath string, reqType RequestType, cb Callback,
) {
	srv.register(reqPath, RequestMethodPut, reqType, ResponseTypeNone, cb)
}

//
// ServeHTTP handle mapping of client request to registered handler.
//
func (srv *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var (
		h  *handler
		ok bool
	)

	if debug.Value > 0 {
		log.Printf("> ServeHTTP: %s %+v\n", req.Method, req.URL)
	}

	switch req.Method {
	case http.MethodDelete:
		h, ok = srv.regDelete[req.URL.Path]

	case http.MethodGet:
		srv.handleGet(res, req)
		return

	case http.MethodHead:
		srv.handleHead(res, req)
		return

	case http.MethodOptions:
		srv.handleOptions(res, req)
		return

	case http.MethodPatch:
		h, ok = srv.regPatch[req.URL.Path]

	case http.MethodPost:
		h, ok = srv.regPost[req.URL.Path]

	case http.MethodPut:
		h, ok = srv.regPut[req.URL.Path]

	default:
		res.WriteHeader(http.StatusNotImplemented)
		return
	}

	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if h == nil {
		res.WriteHeader(http.StatusNotImplemented)
		return
	}

	h.call(res, req)
}

//
// Start the HTTP server.
//
func (srv *Server) Start() error {
	return srv.conn.ListenAndServe()
}

func (srv *Server) getFSNode(reqPath string) (node *memfs.Node) {
	var e error

	node, e = srv.mfs.Get(reqPath)
	if e != nil {
		return nil
	}

	if node.Mode.IsDir() {
		indexHTML := path.Join(reqPath, "index.html")
		node, e = srv.mfs.Get(indexHTML)
		if e != nil {
			return nil
		}
	}

	return node
}

func (srv *Server) handleFS(
	res http.ResponseWriter, req *http.Request, method RequestMethod,
) {
	node := srv.getFSNode(req.URL.Path)
	if node == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set(contentType, node.ContentType)

	if method == RequestMethodHead {
		res.Header().Set("Content-Length", strconv.FormatInt(node.Size, 10))
		res.WriteHeader(http.StatusOK)
		return
	}

	var (
		v []byte
		e error
	)

	if len(node.V) > 0 {
		v = node.V
	} else {
		v, e = ioutil.ReadFile(node.SysPath)
		if e != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	res.WriteHeader(http.StatusOK)
	_, e = res.Write(v)
	if e != nil {
		log.Println("handleFS: ", e.Error())
	}
}

func (srv *Server) handleGet(res http.ResponseWriter, req *http.Request) {
	h, ok := srv.regGet[req.URL.Path]
	if ok {
		h.call(res, req)
		return
	}

	srv.handleFS(res, req, RequestMethodGet)
}

func (srv *Server) handleHead(res http.ResponseWriter, req *http.Request) {
	h, ok := srv.regGet[req.URL.Path]
	if !ok {
		srv.handleFS(res, req, RequestMethodHead)
		return
	}

	switch h.resType {
	case ResponseTypeNone:
		res.WriteHeader(http.StatusNoContent)
		return
	case ResponseTypeBinary:
		res.Header().Set(contentType, contentTypeBinary)
	case ResponseTypeJSON:
		res.Header().Set(contentType, contentTypeJSON)
	case ResponseTypePlain:
		res.Header().Set(contentType, contentTypePlain)
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

	h, ok := srv.regDelete[req.URL.Path]
	if ok && h != nil {
		methods[http.MethodDelete] = true
	}
	_, ok = srv.regGet[req.URL.Path]
	if ok && h != nil {
		methods[http.MethodGet] = true
	}
	_, ok = srv.regPatch[req.URL.Path]
	if ok && h != nil {
		methods[http.MethodPatch] = true
	}
	_, ok = srv.regPost[req.URL.Path]
	if ok && h != nil {
		methods[http.MethodPost] = true
	}
	h, ok = srv.regPut[req.URL.Path]
	if ok && h != nil {
		methods[http.MethodPut] = true
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
// register new handler with specific method, path, request type, and response
// type.
//
func (srv *Server) register(reqPath string, reqMethod RequestMethod,
	reqType RequestType, resType ResponseType, cb Callback,
) {
	if cb == nil {
		return
	}
	if len(reqPath) == 0 {
		reqPath = "/"
	}

	handler := &handler{
		reqType: reqType,
		resType: resType,
		cb:      cb,
	}

	switch reqMethod {
	case RequestMethodDelete:
		srv.regDelete[reqPath] = handler
	case RequestMethodGet:
		srv.regGet[reqPath] = handler
	case RequestMethodPatch:
		srv.regPatch[reqPath] = handler
	case RequestMethodPost:
		srv.regPost[reqPath] = handler
	case RequestMethodPut:
		srv.regPut[reqPath] = handler
	}
}
