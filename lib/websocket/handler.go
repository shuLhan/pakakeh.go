// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
)

//
// clientRawHandler define a callback type for handling raw packet from
// send().
//
type clientRawHandler func(ctx context.Context, resp []byte) (err error)

//
// ClientHandler define a callback type for client to handle packet from
// server (either broadcast or from response of request) in the form of frame.
//
type ClientHandler func(frame *Frame) (err error)

// HandlerAuthFn define server callback type to handle authentication request.
type HandlerAuthFn func(req *Handshake) (ctx context.Context, err error)

// HandlerClientFn define server callback type to handle new client connection
// or removed client connection.
type HandlerClientFn func(ctx context.Context, conn int)

// HandlerPayloadFn define server callback type to handle data frame from
// client.
type HandlerPayloadFn func(conn int, payload []byte)

//
// HandlerFrameFn define a server callback type to handle client request with
// single frame.
//
type HandlerFrameFn func(conn int, frame *Frame)
