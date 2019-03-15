// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

// ContextKey define a type for context.
type ContextKey byte

// List of valid context key.
const (
	CtxKeyExternalJWT ContextKey = 1 << iota
	CtxKeyInternalJWT
	CtxKeyUID

	// Internal context keys used by client.
	ctxKeyWSAccept // ctxKeyWSAccept context key for WebSocket accept key.
)
