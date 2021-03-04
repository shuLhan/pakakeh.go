// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/memfs"
)

//
// ServerOptions define an options to initialize HTTP server.
//
type ServerOptions struct {
	// The server options embed memfs.Options to allow serving directory
	// from the memory.
	//
	// Root contains path to file system to be served.
	// This field is optional, if its empty the server will not serve the
	// local file system, only registered handler.
	//
	// Includes contains list of regex to include files to be served from
	// Root.
	// This field is optional.
	//
	// Excludes contains list of regex to exclude files to be served from
	// Root.
	// This field is optional.
	//
	// Development if its true, the Root file system is served by reading
	// the content directly instead of using memory file system.
	memfs.Options

	// Memfs contains the content of file systems to be served in memory.
	// It will be initialized only if Root is not empty and if its nil.
	Memfs *memfs.MemFS

	// Address define listen address, using ip:port format.
	// This field is optional, default to ":80".
	Address string

	// Conn contains custom HTTP server connection.
	// This fields is optional.
	Conn *http.Server

	// The options for Cross-Origin Resource Sharing.
	CORS CORSOptions
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
