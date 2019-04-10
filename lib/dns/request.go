// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"net/http"
)

//
// request contains UDP address and DNS query message from client.
//
// If Kind is UDP, Sender and UDPAddr must be non nil.
// If Kind is TCP, Sender must be non nil.
// If Kind is DoH, both Sender and UDPAddr must be nil and ResponseWriter and
// ChanResponded must be non nil and initialized.
//
type request struct {
	// Kind define the connection type that this request is belong to,
	// e.g. UDP, TCP, or DoH.
	kind ConnType

	// Message define the DNS query.
	message *Message

	// UDPAddr is address of client if connection is from UDP.
	udpAddr *net.UDPAddr

	// Sender is server connection that receive the query and responsible
	// to answer back to client.
	sender Sender

	// ResponseWriter is HTTP response writer, where answer for DoH
	// client query will be written.
	responseWriter http.ResponseWriter

	// ChanResponded is a channel that notify the DoH handler when answer
	// has been written to ResponseWriter.
	chanResponded chan bool
}

//
// newRequest create and initialize request.
//
func newRequest() *request {
	return &request{
		message: NewMessage(),
	}
}
