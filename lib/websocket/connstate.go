// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

// ConnState represent socket connection status.
type ConnState int

// List of socket connection status.
const (
	ConnStateClosed ConnState = 0
	ConnStateOpen             = 1 << iota
	ConnStateConnected
	ConnStateError
)
