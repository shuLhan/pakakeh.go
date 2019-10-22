// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"io"
	"log"
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
	kind connType

	// Message define the DNS query.
	message *Message

	// writer represent client connection on server that receive the query
	// and responsible to write the answer back.
	// On UDP connection, writer is an instance of UDPClient with
	// connection reference to UDP server and with peer address.
	// On TCP connection, writer is a TCP connection from accept.
	// On Doh connection, writer is http ResponseWriter.
	writer io.Writer
}

//
// newRequest create and initialize request.
//
func newRequest() *request {
	return &request{
		message: NewMessage(),
	}
}

//
// error set the request message as an error.
//
func (req *request) error(rcode ResponseCode) {
	req.message.SetQuery(false)
	req.message.SetResponseCode(rcode)

	_, err := req.writer.Write(req.message.Packet)
	if err != nil {
		log.Println("dns: request.error:", err.Error())
	}
}
