// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

// connState represent socket connection status.
type connState int

// List of socket connection status.
const (
	connStateClosed connState = 0
	connStateOpen             = iota
	connStateConnected
	connStateError
)
