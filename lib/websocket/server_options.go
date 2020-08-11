// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
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

	// HandleAuth callback that will be called when receiving
	// client handshake.
	HandleAuth HandlerAuthFn

	// HandleClientAdd callback that will called after client handshake
	// and, if HandleAuth is defined, after client is authenticated.
	HandleClientAdd HandlerClientFn

	// HandleClientRemove callback that will be called before client
	// connection being removed and closed by server.
	HandleClientRemove HandlerClientFn

	// HandleRsvControl callback that will be called when server received
	// reserved control frame (opcode 0xB-F) from client.
	// Default handle is nil.
	HandleRsvControl HandlerFrameFn

	// HandleText callback that will be called after receiving data
	// frame(s) text from client.
	// Default handle parse the payload into Request and pass it to
	// registered routes.
	HandleText HandlerPayloadFn

	// HandleBin callback that will be called after receiving data
	// frame(s) binary from client.
	HandleBin HandlerPayloadFn

	// HandleStatus function that will be called when server receive
	// request for status as defined in ServerOptions.StatusPath.
	HandleStatus HandlerStatusFn
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
}
