// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package websocket provide the websocket library for server and
// client.
//
// The websocket server is implemented with epoll, which means it's only
// run on Linux.
//
// References
//
// - https://tools.ietf.org/html/rfc6455
//
// - https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers
//
// - http://man7.org/linux/man-pages/man7/epoll.7.html
//
package websocket

import (
	"bytes" // nolint: gosec
	"math/rand"
	"sync"
	"time"
)

// List of frame length.
const (
	frameSmallPayload  = 125
	frameMediumPayload = 126
	frameLargePayload  = 127
)

// List of frame FIN and MASK values.
const (
	frameIsFinished = 0x80
	frameIsMasked   = 0x80
)

const (
	_qKeyTicket = "ticket"
)

var defaultTimeout = 10 * time.Second //nolint: gochecknoglobals

var _rng *rand.Rand //nolint: gochecknoglobals

var _bbPool = sync.Pool{ //nolint: gochecknoglobals
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
