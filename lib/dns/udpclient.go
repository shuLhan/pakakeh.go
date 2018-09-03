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
// UDPClient for DNS with UDP connection and list of remote addresses.
//
type UDPClient struct {
	Timeout time.Duration

	// Address of remote nameserver.
	Addr *net.UDPAddr
	conn *net.UDPConn
}

//
// NewUDPClient will create new DNS client with UDP network connection.
//
func NewUDPClient(nameserver string) (cl *UDPClient, err error) {
	network := "udp"

	raddr, err := net.ResolveUDPAddr(network, nameserver)
	if err != nil {
		return
	}

	laddr := &net.UDPAddr{IP: nil, Port: 0}
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		return
	}

	cl = &UDPClient{
		Timeout: clientTimeout,
		Addr:    raddr,
		conn:    conn,
	}

	return
}

//
// RemoteAddr return client remote nameserver address.
//
func (cl *UDPClient) RemoteAddr() net.Addr {
	return cl.Addr
}

//
// Close client connection.
//
func (cl *UDPClient) Close() error {
	return cl.conn.Close()
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
func (cl *UDPClient) Lookup(qtype uint16, qclass uint16, qname []byte) (
	*Message, error,
) {
	if cl.Addr == nil || cl.conn == nil {
		return nil, nil
	}

	msg := msgPool.Get().(*Message)
	msg.Reset()

	msg.Header.ID = getNextID()
	msg.Header.QDCount = 1
	msg.Header.IsRD = true
	msg.Question.Type = qtype
	msg.Question.Class = qclass
	msg.Question.Name = append(msg.Question.Name, qname...)

	_, _ = msg.Pack()

	_, err := cl.Send(msg, cl.Addr)
	if err != nil {
		msgPool.Put(msg)
		return nil, err
	}

	resMsg := msgPool.Get().(*Message)
	resMsg.Reset()

	_, err = cl.Recv(resMsg)
	if err != nil {
		msgPool.Put(msg)
		msgPool.Put(resMsg)
		return nil, err
	}

	err = resMsg.Unpack()
	if err != nil {
		msgPool.Put(msg)
		msgPool.Put(resMsg)
		return nil, err
	}

	FreeMessage(msg)

	return resMsg, nil
}

//
// Send DNS message to name server using active connection in client.
//
// The message packet must already been filled, using Pack().
// The addr parameter must not be nil.
//
func (cl *UDPClient) Send(msg *Message, ns net.Addr) (n int, err error) {
	if ns == nil {
		return
	}

	raddr := ns.(*net.UDPAddr)

	err = cl.conn.SetWriteDeadline(time.Now().Add(cl.Timeout))
	if err != nil {
		return
	}

	n, err = cl.conn.WriteToUDP(msg.Packet, raddr)

	return
}

//
// Recv will read DNS message from active connection in client into `msg`.
//
func (cl *UDPClient) Recv(msg *Message) (n int, err error) {
	err = cl.conn.SetReadDeadline(time.Now().Add(cl.Timeout))
	if err != nil {
		return
	}

	n, _, err = cl.conn.ReadFromUDP(msg.Packet)
	if err != nil {
		return
	}

	msg.Packet = append(msg.Packet[:0], msg.Packet[:n]...)

	if debugLevel >= 2 {
		libbytes.PrintHex(">>> UDPClient: Recv:", msg.Packet, 8)
	}

	return
}
