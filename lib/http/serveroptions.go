// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
)

// ServerOptions define an options to initialize HTTP server.
type ServerOptions struct {
	// Memfs contains the content of file systems to be served in memory.
	// The MemFS instance to be served should be already embedded in Go
	// file, generated using memfs.MemFS.GoEmbed().
	// Otherwise, it will try to read from file system directly.
	//
	// See https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/memfs#hdr-Go_embed
	Memfs *memfs.MemFS

	// HandleFS inspect each GET request to Memfs.
	// Some usage of this handler is to check for authorization on
	// specific path, handling redirect, and so on.
	// If nil it means all request are allowed.
	// See FSHandler for more information.
	HandleFS FSHandler

	// Address define listen address, using ip:port format.
	// This field is optional, default to ":80".
	Address string

	// Conn contains custom HTTP server connection.
	// This fields is optional.
	Conn *http.Server

	// ErrorWriter define the writer where output from panic in handler
	// will be written.  Basically this will create new log.Logger and set
	// the default Server.ErrorLog.
	// This field is optional, but if its set it will be used only if Conn
	// is not set by caller.
	ErrorWriter io.Writer

	// The options for Cross-Origin Resource Sharing.
	CORS CORSOptions

	// If true, server generate index.html automatically if its not
	// exist in the directory.
	// The index.html contains the list of files inside the requested
	// path.
	EnableIndexHtml bool //revive:disable-line
}

func (opts *ServerOptions) init() {
	if len(opts.Address) == 0 {
		opts.Address = ":80"
	}

	if opts.Conn == nil {
		opts.Conn = &http.Server{
			ReadTimeout:    defRWTimeout,
			WriteTimeout:   defRWTimeout,
			MaxHeaderBytes: 1 << 20,
		}
		if opts.ErrorWriter != nil {
			opts.Conn.ErrorLog = log.New(opts.ErrorWriter, "", 0)
		}
	}

	for x := 0; x < len(opts.CORS.AllowOrigins); x++ {
		if opts.CORS.AllowOrigins[x] == corsWildcard {
			opts.CORS.allowOriginsAll = true
			break
		}
	}

	for x := 0; x < len(opts.CORS.AllowHeaders); x++ {
		if opts.CORS.AllowHeaders[x] == corsWildcard {
			opts.CORS.allowHeadersAll = true
		} else {
			opts.CORS.AllowHeaders[x] = strings.ToLower(opts.CORS.AllowHeaders[x])
		}
	}

	if len(opts.CORS.ExposeHeaders) > 0 {
		opts.CORS.exposeHeaders = strings.Join(opts.CORS.ExposeHeaders, ",")
	}
	if opts.CORS.MaxAge > 0 {
		opts.CORS.maxAge = strconv.Itoa(opts.CORS.MaxAge)
	}
}
