// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// TCPClient for DNS with TCP connection and list of remote addresses.
//
type TCPClient struct {
	Timeout time.Duration
	addr    *net.TCPAddr
	conn    *net.TCPConn
}

//
// NewTCPClient will create new DNS client with TCP network connection.
//
func NewTCPClient(nameserver string) (*TCPClient, error) {
	network := "tcp"

	raddr, err := net.ResolveTCPAddr(network, nameserver)
	if err != nil {
		return nil, err
	}

	cl := &TCPClient{
		Timeout: clientTimeout,
		addr:    raddr,
	}

	err = cl.Connect(raddr)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

//
// RemoteAddr return client remote nameserver address.
//
func (cl *TCPClient) RemoteAddr() net.Addr {
	return cl.addr
}

//
// Connect to remote address.
//
func (cl *TCPClient) Connect(raddr *net.TCPAddr) (err error) {
	laddr := &net.TCPAddr{IP: nil, Port: 0}

	cl.conn, err = net.DialTCP("tcp", laddr, raddr)

	return
}

//
// Lookup will query one of the name server with specific type, class, and
// name in synchronous mode.
//
// Name could be a host name for standard query or IP address for inverse
// query.
//
// This function is safe to be used concurrently.
//
func (cl *TCPClient) Lookup(qtype uint16, qclass uint16, qname []byte) (
	*Message, error,
) {
	if cl.addr == nil || cl.conn == nil {
		return nil, nil
	}

	msg := AllocMessage()

	msg.Header.ID = getNextID()
	msg.Header.IsRD = true
	msg.Header.QDCount = 1
	msg.Question.Type = qtype
	msg.Question.Class = qclass
	msg.Question.Name = append(msg.Question.Name, qname...)

	_, _ = msg.MarshalBinary()

	_, err := cl.Send(msg, nil)
	if err != nil {
		FreeMessage(msg)
		return nil, err
	}

	resMsg := AllocMessage()

	_, err = cl.Recv(resMsg)
	if err != nil {
		FreeMessage(msg)
		FreeMessage(resMsg)
		return nil, err
	}

	err = resMsg.UnmarshalBinary(resMsg.Packet)
	if err != nil {
		FreeMessage(msg)
		FreeMessage(resMsg)
		return nil, err
	}

	FreeMessage(msg)

	return resMsg, nil
}

//
// Send DNS message to name server using active connection in client.
//
// The message packet must already been filled, using MarshalBinary().
// The addr parameter is unused.
//
func (cl *TCPClient) Send(msg *Message, addr net.Addr) (n int, err error) {
	err = cl.conn.SetWriteDeadline(time.Now().Add(cl.Timeout))
	if err != nil {
		return
	}

	packet := make([]byte, 0)

	libbytes.AppendUint16(&packet, uint16(len(msg.Packet)))
	packet = append(packet, msg.Packet...)

	n, err = cl.conn.Write(packet)

	return
}

//
// Recv will read DNS message from active connection in client into `msg`.
//
func (cl *TCPClient) Recv(msg *Message) (n int, err error) {
	err = cl.conn.SetReadDeadline(time.Now().Add(cl.Timeout))
	if err != nil {
		return
	}

	n, err = cl.conn.Read(msg.Packet)
	if err != nil {
		return
	}
	if n == 0 {
		return
	}

	msg.Packet = append(msg.Packet[:0], msg.Packet[:n]...)

	if debugLevel >= 2 {
		libbytes.PrintHex(">>> TCP msg.Packet:", msg.Packet, 8)
	}

	msg.Packet = append(msg.Packet[:0], msg.Packet[2:]...)

	if debugLevel >= 2 {
		libbytes.PrintHex(">>> DNS msg.Packet:", msg.Packet, 8)
	}

	return
}
