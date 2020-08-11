// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"log"
	"path"
)

const (
	defServerAddress     = ":80"
	defServerConnectPath = "/"
	defServerStatusPath  = "/status"
)

type ServerOptions struct {
	// Address to listen for WebSocket connection.
	// Default to ":80".
	Address string

	// ConnectPath define the HTTP path where WebSocket connection
	// handshake will be processed. Default to "/".
	ConnectPath string

	// StatusPath define a HTTP path to check for server status.
	// Default to ConnectPath +"/status" if its empty.
	// The StatusPath is handled by HandleStatus callback in the server.
	StatusPath string
}

func (opts *ServerOptions) init() {
	if len(opts.Address) == 0 {
		opts.Address = defServerAddress
	}
	if len(opts.ConnectPath) == 0 {
		opts.ConnectPath = defServerConnectPath
	}
	if len(opts.StatusPath) == 0 {
		opts.StatusPath = path.Join(opts.ConnectPath, defServerStatusPath)
	}

	log.Printf("opts: %+v\n", opts)
}
