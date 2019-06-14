// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bytes" //nolint:gosec
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

//nolint:gochecknoglobals
var (
	defaultTimeout = 10 * time.Second

	_bbPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

//nolint:gochecknoinits
func init() {
	rand.Seed(time.Now().UnixNano())
}
