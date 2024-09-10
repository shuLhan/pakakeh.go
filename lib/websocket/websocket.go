// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

import (
	"bytes"
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

var (
	defaultTimeout      = 10 * time.Second
	defaultPingInterval = 10 * time.Second

	_bbPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)
