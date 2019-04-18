// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
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

//
// Server define HTTP server.
//
type Server struct {
	mfs       *memfs.MemFS
	evals     []Evaluator
	conn      *http.Server
	regDelete map[string]*Endpoint
	regGet    map[string]*Endpoint
	regPatch  map[string]*Endpoint
	regPost   map[string]*Endpoint
	regPut    map[string]*Endpoint
}

//
// NewServer create and initialize new HTTP server that serve root directory
// with custom connection.
//
func NewServer(opts *ServerOptions) (srv *Server, e error) {
	srv = &Server{
		regDelete: make(map[string]*Endpoint),
		regGet:    make(map[string]*Endpoint),
		regPatch:  make(map[string]*Endpoint),
		regPost:   make(map[string]*Endpoint),
		regPut:    make(map[string]*Endpoint),
	}

	if len(opts.Address) == 0 {
		opts.Address = ":80"
	}

	if opts.Conn == nil {
		srv.conn = &http.Server{
			Addr:           opts.Address,
			Handler:        srv,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
	} else {
		opts.Conn.Handler = srv
		srv.conn = opts.Conn
	}

	memfs.Development = opts.Development

	srv.mfs, e = memfs.New(opts.Includes, opts.Excludes, true)
	if e != nil {
		return nil, e
	}

	e = srv.mfs.Mount(opts.Root)
	if e != nil {
		return nil, e
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
// RegisterDelete register HTTP method DELETE with callback to handle it.
//
func (srv *Server) RegisterDelete(ep *Endpoint) {
	if ep != nil {
		ep.Method = RequestMethodDelete
		ep.RequestType = RequestTypeQuery
		srv.register(ep)
	}
}

//
// RegisterEvaluator register HTTP middleware that will be called before
// Endpoint evalutor and callback is called.
//
func (srv *Server) RegisterEvaluator(eval Evaluator) {
	srv.evals = append(srv.evals, eval)
}

//
// RegisterGet register HTTP method GET with callback to handle it.
//
func (srv *Server) RegisterGet(ep *Endpoint) {
	if ep != nil {
		ep.Method = RequestMethodGet
		ep.RequestType = RequestTypeQuery
		srv.register(ep)
	}
}

//
// RegisterPatch register HTTP method PATCH with callback to handle it.
//
func (srv *Server) RegisterPatch(ep *Endpoint) {
	if ep != nil {
		ep.Method = RequestMethodPatch
		srv.register(ep)
	}
}

//
// RegisterPost register HTTP method POST with callback to handle it.
//
func (srv *Server) RegisterPost(ep *Endpoint) {
	if ep != nil {
		ep.Method = RequestMethodPost
		srv.register(ep)
	}
}

//
// RegisterPut register HTTP method PUT with callback to handle it.
//
func (srv *Server) RegisterPut(ep *Endpoint) {
	if ep != nil {
		ep.Method = RequestMethodPut
		ep.ResponseType = ResponseTypeNone
		srv.register(ep)
	}
}

//
// ServeHTTP handle mapping of client request to registered endpoints.
//
func (srv *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var (
		ep *Endpoint
		ok bool
	)

	if debug.Value > 0 {
		log.Printf("> ServeHTTP: %s %+v\n", req.Method, req.URL)
	}

	switch req.Method {
	case http.MethodDelete:
		ep, ok = srv.regDelete[req.URL.Path]

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
		ep, ok = srv.regPatch[req.URL.Path]

	case http.MethodPost:
		ep, ok = srv.regPost[req.URL.Path]

	case http.MethodPut:
		ep, ok = srv.regPut[req.URL.Path]

	default:
		res.WriteHeader(http.StatusNotImplemented)
		return
	}
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	ep.call(res, req, srv.evals)
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
	var e error

	node, e = srv.mfs.Get(reqPath)
	if e != nil {
		if e != os.ErrNotExist {
			log.Println("http: getFSNode: " + e.Error())
			return nil
		}

		reqPath = path.Join(reqPath, "index.html")

		node, e = srv.mfs.Get(reqPath)
		if e != nil {
			log.Println("http: getFSNode: " + e.Error())
			return nil
		}
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

	res.Header().Set(ContentType, node.ContentType)

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
	ep, ok := srv.regGet[req.URL.Path]
	if ok {
		ep.call(res, req, srv.evals)
		return
	}

	srv.handleFS(res, req, RequestMethodGet)
}

func (srv *Server) handleHead(res http.ResponseWriter, req *http.Request) {
	ep, ok := srv.regGet[req.URL.Path]
	if !ok {
		srv.handleFS(res, req, RequestMethodHead)
		return
	}

	switch ep.ResponseType {
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

	ep, ok := srv.regDelete[req.URL.Path]
	if ok && ep != nil {
		methods[http.MethodDelete] = true
	}
	_, ok = srv.regGet[req.URL.Path]
	if ok && ep != nil {
		methods[http.MethodGet] = true
	}
	_, ok = srv.regPatch[req.URL.Path]
	if ok && ep != nil {
		methods[http.MethodPatch] = true
	}
	_, ok = srv.regPost[req.URL.Path]
	if ok && ep != nil {
		methods[http.MethodPost] = true
	}
	ep, ok = srv.regPut[req.URL.Path]
	if ok && ep != nil {
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
// register new endpoint with specific method, path, request type, and
// response type.
//
func (srv *Server) register(ep *Endpoint) {
	if ep == nil || ep.Call == nil {
		return
	}
	if len(ep.Path) == 0 {
		ep.Path = "/"
	}

	switch ep.Method {
	case RequestMethodDelete:
		srv.regDelete[ep.Path] = ep
	case RequestMethodGet:
		srv.regGet[ep.Path] = ep
	case RequestMethodPatch:
		srv.regPatch[ep.Path] = ep
	case RequestMethodPost:
		srv.regPost[ep.Path] = ep
	case RequestMethodPut:
		srv.regPut[ep.Path] = ep
	}
}
