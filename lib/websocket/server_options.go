// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"path"
	"time"
)

const (
	defServerAddress     = ":80"
	defServerConnectPath = "/"
	defServerStatusPath  = "/status"

	defServerReadWriteTimeout           = 30 * time.Second
	defServerMaxGoroutinePinger         = _maxQueue / 4
	defServerMaxGoroutineReader         = 1024
	defServerMaxGoroutineUpgrader int32 = 128
)

// ServerOptions contain options to configure the WebSocket server.
type ServerOptions struct {
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

	// Address to listen for WebSocket connection.
	// Default to ":80".
	Address string

	// ConnectPath define the HTTP path where WebSocket connection
	// handshake will be processed.
	// Default to "/".
	ConnectPath string

	// StatusPath define a HTTP path to check for server status.
	// Default to ConnectPath +"/status" if its empty.
	// The StatusPath is handled by HandleStatus callback in the server.
	StatusPath string

	// ReadWriteTimeout define the maximum duration the server wait for
	// receiving/sending packet from/to client before considering the
	// connection as broken.
	// Default to 30 seconds.
	ReadWriteTimeout time.Duration

	// maxGoroutinePinger define maximum number of goroutines to ping each
	// connected clients at the same time.
	maxGoroutinePinger int32

	maxGoroutineReader int32

	// maxGoroutineUpgrader define maximum goroutines running at the same
	// time to handle client upgrade.
	// The new goroutine only dispatched when others are full, so it will
	// run incrementally not all at once.
	// Default to defServerMaxGoroutineUpgrader if its not set.
	maxGoroutineUpgrader int32
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
	if opts.ReadWriteTimeout <= 0 {
		opts.ReadWriteTimeout = defServerReadWriteTimeout
	}
	if opts.maxGoroutinePinger <= 0 {
		opts.maxGoroutinePinger = defServerMaxGoroutinePinger
	}
	if opts.maxGoroutineReader <= 0 {
		opts.maxGoroutineReader = defServerMaxGoroutineReader
	}
	if opts.maxGoroutineUpgrader <= 0 {
		opts.maxGoroutineUpgrader = defServerMaxGoroutineUpgrader
	}
}
