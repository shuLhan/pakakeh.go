// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
)

type ContextKey uint64

const (
	CtxKeyExternalJWT ContextKey = 1 << iota
	CtxKeyInternalJWT
	CtxKeyUID
)

type HandlerFn func(conn int, req *Frame)
type HandlerAuthFn func(req *Handshake) (ctx context.Context, err error)
type HandlerClientFn func(ctx context.Context, conn int)
