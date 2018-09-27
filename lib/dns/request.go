// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"net/http"
)

// List of known connection type.
const (
	ConnTypeUDP = 1
	ConnTypeTCP = 2
	ConnTypeDoH = 4
)

//
// Request contains UDP address and DNS query message from client.
//
// If Kind is UDP, Sender and UDPAddr must be non nil.
// If Kind is TCP, Sender must be non nil.
// If Kind is DoH, both Sender and UDPAddr must be nil and ResponseWriter and
// ChanResponded must be non nil and initialized.
//
type Request struct {
	// Kind define the connection type that this request is belong to,
	// e.g. UDP, TCP, or DoH.
	Kind int

	// Message define the DNS query.
	Message *Message

	// UDPAddr is address of client if connection is from UDP.
	UDPAddr *net.UDPAddr

	// Sender is server connection that receive the query and responsible
	// to answer back to client.
	Sender Sender

	// ResponseWriter is HTTP response writer, where answer for DoH
	// client query will be written.
	ResponseWriter http.ResponseWriter

	// ChanResponded is a channel that notify the DoH handler when answer
	// has been written to ResponseWriter.
	ChanResponded chan bool
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
