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

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
)

// ServerOptions define an options to initialize HTTP server.
type ServerOptions struct {
	// Listener define the network listener to be used for serving HTTP
	// connection.
	// The Listener can be activated using systemd socket.
	Listener net.Listener

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

	// BasePath define the base path or prefix to serve the HTTP request
	// and response.
	// Each request that server received will remove the BasePath first
	// from the [http.Request.URL.Path] before passing to the handler.
	// Each redirect that server sent will add the BasePath as the prefix
	// to redirect URL.
	//
	// Any trailing slash in the BasePath will be removed.
	BasePath string

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

	// ShutdownIdleDuration define the duration where the server will
	// automatically stop accepting new connection and then shutting down
	// the server.
	ShutdownIdleDuration time.Duration

	// If true, server generate index.html automatically if its not
	// exist in the directory.
	// The index.html contains the list of files inside the requested
	// path.
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
