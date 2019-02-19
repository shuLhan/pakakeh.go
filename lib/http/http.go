// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package http implement custom HTTP server with memory file system and
// simplified routing handler.
//
// Problems
//
// There are two problems that this library try to handle.
// First, optimizing serving local file system; second, complexity of routing
// regarding to their method, request type, and response type.
//
// Assuming that we want to serve file system and API using ServeMux, the
// simplest registered handler are,
//
//	mux.HandleFunc("/", handleFileSystem)
//	mux.HandleFunc("/api", handleAPI)
//
// The first problem is regarding to "http.ServeFile".
// Everytime the request hit "handleFileSystem" the "http.ServeFile" try to
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
// Solutions
//
// The solution to the first problem is by mapping all content of files to be
// served into memory.
// This cause more memory to be consumed on server side but we minimize path
// lookup, and cache-miss on OS level.
//
// Serving file system is handled by memory file system using map of path to
// file node.
//
//	map[/index.html] = Node{Type: ..., Size: ..., ContentType: ...}
//
// There is a limit on size of file to be mapped on memory.
// See the package "lib/memfs" for more information.
//
// The solution to the second problem is by mapping the registered request per
// method and by path.
// User just need to focus on step 3, handling on how to process request, all
// of process on step 1, 2, and 4 will be handled by our library.
//
//	epAPILogin := &libhttp.Endpoint{
//		Path: "/api/login",
//		RequestType: libhttp.RequestTypeQuery,
//		ResponseType: libhttp.ResponseTypeJSON,
//		Call: handleLogin,
//	}
//	server.RequestPost(epAPILogin)
//
// Upon receiving request to "/api/login", the library will call
// "req.ParseForm()", read the content of body and pass them to
// "handleLogin",
//
//	func handleLogin(req *http.Request, reqBody []byte) (resBody []byte, err error) {
//		// Process login input from req.Form, req.PostForm, and/or
//		// reqBody.
//		// Return response body and error.
//	}
//
// Known Bugs
//
// * The server does not handle CONNECT method
//
// * Missing test for request with content-type multipart-form
//
package http

const (
	ContentLength     = "Content-Length"
	ContentType       = "Content-Type"
	ContentTypeBinary = "application/octet-stream"
	ContentTypeForm   = "application/x-www-form-urlencoded"
	ContentTypeJSON   = "application/json"
	ContentTypePlain  = "text/plain"
	HeaderLocation    = "Location"
)
