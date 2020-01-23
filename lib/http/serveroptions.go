// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

//
// ServerOptions define an options to initialize HTTP server.
//
type ServerOptions struct {
	// Root contains path to file system to be served.
	// This field is options, if its empty the server will not serve the
	// file system only registered handler.
	Root string

	// Address define listen address, using ip:port format.
	// This field is optional, default to ":80".
	Address string

	// Conn contains custom HTTP server connection.
	// This fields is optional.
	Conn *http.Server

	// Includes contains list of regex to include files to be served from
	// Root.
	// This field is optional.
	Includes []string

	// Excludes contains list of regex to exclude files to be served from
	// Root.
	// This field is optional.
	Excludes []string

	// Development if its true, the Root file system is served by reading
	// the content directly instead of using memory file system.
	Development bool
}
