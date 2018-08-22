// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"sync"
)

var _requestPool = sync.Pool{
	New: func() interface{} {
		req := &Request{
			Message: AllocMessage(),
		}
		return req
	},
}

//
// Request contains UDP address and DNS query message from client.
//
type Request struct {
	Message *Message
	UDPAddr *net.UDPAddr
}

//
// Reset message and UDP address in request.
//
func (req *Request) Reset() {
	req.UDPAddr = nil
	req.Message.Reset()
}
