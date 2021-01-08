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

	// Address define listen address, using ip:port format.
	// This field is optional, default to ":80".
	Address string

	// Conn contains custom HTTP server connection.
	// This fields is optional.
	Conn *http.Server

	// CORSAllowOrigins contains global list of cross-site Origin that are
	// allowed during preflight requests by the OPTIONS method.
	// The list is case-sensitive.
	// To allow all Origin, one must add "*" string to the list.
	CORSAllowOrigins []string

	// CORSAllowHeaders contains global list of allowed headers during
	// preflight requests by the OPTIONS method.
	// The list is case-insensitive.
	// To allow all headers, one must add "*" string to the list.
	CORSAllowHeaders []string

	// CORSExposeHeaders contains list of allowed headers.
	// This list will be send when browser request OPTIONS without
	// request-method.
	CORSExposeHeaders []string
	exposeHeaders     string

	// CORSMaxAge gives the value in seconds for how long the response to
	// the preflight request can be cached for without sending another
	// preflight request.
	CORSMaxAge int
	corsMaxAge string

	// CORSAllowCredentials indicates whether or not the actual request
	// can be made using credentials.
	CORSAllowCredentials bool

	corsAllowHeadersAll bool // flag to indicate wildcards on list.
	corsAllowOriginsAll bool // flag to indicate wildcards on list.
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

	for x := 0; x < len(opts.CORSAllowOrigins); x++ {
		if opts.CORSAllowOrigins[x] == corsWildcard {
			opts.corsAllowOriginsAll = true
			break
		}
	}

	for x := 0; x < len(opts.CORSAllowHeaders); x++ {
		if opts.CORSAllowHeaders[x] == corsWildcard {
			opts.corsAllowHeadersAll = true
		} else {
			opts.CORSAllowHeaders[x] = strings.ToLower(opts.CORSAllowHeaders[x])
		}
	}

	if len(opts.CORSExposeHeaders) > 0 {
		opts.exposeHeaders = strings.Join(opts.CORSExposeHeaders, ",")
	}
	if opts.CORSMaxAge > 0 {
		opts.corsMaxAge = strconv.Itoa(opts.CORSMaxAge)
	}
}
