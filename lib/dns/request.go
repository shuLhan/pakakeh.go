// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"net/http"
)

//
// Request contains UDP address and DNS query message from client.
//
type Request struct {
	Message        *Message
	UDPAddr        *net.UDPAddr
	Sender         Sender
	ResponseWriter http.ResponseWriter
	ChanResponded  chan bool
}

//
// NewRequest create and initialize request.
//
func NewRequest() *Request {
	return &Request{
		Message: NewMessage(),
	}
}

//
// Reset message and UDP address in request.
//
func (req *Request) Reset() {
	req.Message.Reset()
	req.UDPAddr = nil
	req.Sender = nil
	req.ResponseWriter = nil
}
