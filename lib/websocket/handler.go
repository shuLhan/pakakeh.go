// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
)

//
// ClientHandler define a callback type for client to handle packet from
// server (either broadcast or from response of request) in the form of frame.
//
// Returning a non-nil error will cause the underlying connection to be
// closed.
//
type ClientHandler func(cl *Client, frame *Frame) (err error)

// HandlerAuthFn define server callback type to handle authentication request.
type HandlerAuthFn func(req *Handshake) (ctx context.Context, err error)

// HandlerClientFn define server callback type to handle new client connection
// or removed client connection.
type HandlerClientFn func(ctx context.Context, conn int)

// HandlerPayloadFn define server callback type to handle data frame from
// client.
type HandlerPayloadFn func(conn int, payload []byte)

//
// HandlerStatusFn define server callback type to handle status request.
// It must return the content type of data, for example "text/plain", and the
// status data to be send to client.
//
type HandlerStatusFn func() (contentType string, data []byte)

//
// HandlerFrameFn define a server callback type to handle client request with
// single frame.
//
type HandlerFrameFn func(conn int, frame *Frame)
