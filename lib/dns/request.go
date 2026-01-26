// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package dns

import (
	"fmt"
	"io"
	"log"
	"time"
)

// List of connection type for [request.kind].
const (
	connTypeDoH = `DoH`
	connTypeDoT = `DoT`
	connTypeTCP = `TCP`
	connTypeUDP = `UDP`
)

// request contains UDP address and DNS query message from client.
//
// If Kind is UDP, Sender and UDPAddr must be non nil.
// If Kind is TCP, Sender must be non nil.
// If Kind is DoH, both Sender and UDPAddr must be nil and ResponseWriter and
// ChanResponded must be non nil and initialized.
type request struct {
	// writer represent client connection on server that receive the query
	// and responsible to write the answer back.
	// On UDP connection, writer is an instance of UDPClient with
	// connection reference to UDP server and with peer address.
	// On TCP connection, writer is a TCP connection from accept.
	// On Doh connection, writer is http ResponseWriter.
	writer io.Writer

	// Message define the DNS query.
	message *Message

	// startAt set the start time the request received by server.
	startAt time.Time

	// kind define the connection type that this request is belong to:
	// DOH, DOT, TCP, or UDP.
	kind string
}

// newRequest create and initialize request.
func newRequest() (req *request) {
	req = &request{
		message: NewMessage(),
		startAt: time.Now(),
	}
	return req
}

func (req *request) String() string {
	return fmt.Sprintf(`{%d %s %s}`, req.message.Header.ID,
		req.message.Question.Name,
		RecordTypeNames[req.message.Question.Type])
}

// error set the request message as an error.
func (req *request) error(rcode ResponseCode) {
	var err error

	req.message.SetQuery(false)
	req.message.SetResponseCode(rcode)

	_, err = req.writer.Write(req.message.packet)
	if err != nil {
		log.Println("dns: request.error:", err.Error())
	}
}
