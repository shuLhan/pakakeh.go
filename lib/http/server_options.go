// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package http

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"kilabit.info/pakakeh.go/lib/memfs"
)

// ServerOptions define an options to initialize HTTP server.
type ServerOptions struct {
	// The network listener to be used for serving HTTP connection.
	// This field is optional, if set it will be used instead of Address.
	Listener net.Listener

	// Memory file systems to be served when server running.
	// The MemFS instance to be served should be already embedded in Go
	// file, generated using [memfs.MemFS.GoEmbed].
	// Otherwise, it will try to read from file system directly.
	//
	// See https://pkg.go.dev/kilabit.info/pakakeh.go/lib/memfs#hdr-Go_embed
	// for more information on how the creates Go embed file using memfs.
	//
	// This field is optional.
	Memfs *memfs.MemFS

	// HandleFS inspect each GET request to Memfs.
	// If nil the request will be handled using content from
	// [memfs.MemFS.Get].
	//
	// Some usage of this handler is to check for authorization on
	// specific path, handling redirect, and so on.
	// See [FSHandler] for more information.
	HandleFS FSHandler

	// Address define listen address, using ip:port format.
	// This field is optional, default to ":80".
	Address string

	// BasePath define the base path or prefix to serve the HTTP request
	// and response.
	// Any trailing slash in the BasePath will be removed.
	//
	// Each request that server received will remove the BasePath first
	// from the [http.Request.URL.Path] before passing it to the
	// [Endpoint.Call].
	//
	// Each redirect that server sent will add the BasePath as the prefix
	// to redirect URL.
	//
	// This field is optional.
	BasePath string

	// Conn contains custom HTTP server connection.
	// This fields is optional, it will replace the [Server.Server] field.
	Conn *http.Server

	// The writer where output from panic in handler will be written.
	// It will create new [log.Logger] and set the default
	// [http.Server.ErrorLog].
	// This field is optional, and will be used only if Conn field is not
	// set.
	ErrorWriter io.Writer

	// Options for Cross-Origin Resource Sharing (CORS).
	// This field is optional.
	CORS CORSOptions

	// The duration where the server will automatically stop accepting new
	// connection and then shutting down the server.
	// This field is optional.
	ShutdownIdleDuration time.Duration

	// If true, server generates index.html automatically if the request
	// path is directory without index.html in Memfs.
	// The index.html contains the list of files inside the directory.
	// This field is optional.
	EnableIndexHTML bool
}

func (opts *ServerOptions) init() {
	if len(opts.Address) == 0 {
		opts.Address = ":80"
	}

	opts.BasePath = strings.TrimRight(opts.BasePath, `/`)

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

	opts.CORS.init()
}
