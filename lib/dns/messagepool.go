// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync"
)

var msgPool = sync.Pool{
	New: func() interface{} {
		return NewMessage()
	},
}

//
// AllocMessage from pool.
//
func AllocMessage() (msg *Message) {
	msg = msgPool.Get().(*Message)
	msg.Reset()

	return
}

//
// FreeMessage put the message back to the pool.
//
func FreeMessage(msg *Message) {
	msgPool.Put(msg)
}
