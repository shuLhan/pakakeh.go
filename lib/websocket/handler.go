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
// ClientRecvHandler define a custom callback type for handling response from
// request in the form of frames.
//
type ClientRecvHandler func(ctx context.Context, frames *Frames) (err error)

// HandlerFn callback type to handle handshake request.
type HandlerFn func(conn int, req *Frame)

// HandlerAuthFn callback type to handle authentication request.
type HandlerAuthFn func(req *Handshake) (ctx context.Context, err error)

// HandlerClientFn callback type to handle client request.
type HandlerClientFn func(ctx context.Context, conn int)

// HandlerPayload define a callback type to handle data frame.
type HandlerPayload func(conn int, payload []byte)
