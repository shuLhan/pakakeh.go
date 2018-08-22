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
	addr *net.UDPAddr
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
		addr: raddr,
		conn: conn,
	}

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
func (cl *UDPClient) Lookup(qtype uint16, qclass uint16, qname []byte) (
	*Message, error,
) {
	if cl.addr == nil || cl.conn == nil {
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

	_, _ = msg.MarshalBinary()

	err := cl.Send(msg, cl.addr)
	if err != nil {
		msgPool.Put(msg)
		return nil, err
	}

	resMsg := msgPool.Get().(*Message)
	resMsg.Reset()

	err = cl.Recv(resMsg)
	if err != nil {
		msgPool.Put(msg)
		msgPool.Put(resMsg)
		return nil, err
	}

	err = resMsg.UnmarshalBinary(resMsg.Packet)
	if err != nil {
		msgPool.Put(msg)
		msgPool.Put(resMsg)
		return nil, err
	}

	return resMsg, nil
}

//
// Send DNS message to name server using active connection in client.
//
// The message packet must already been filled, using MarshalBinary().
// The ns parameter must not be nil.
//
func (cl *UDPClient) Send(msg *Message, ns *net.UDPAddr) (err error) {
	if ns == nil {
		return
	}

	err = cl.conn.SetDeadline(time.Now().Add(clientTimeout))
	if err != nil {
		return
	}

	_, err = cl.conn.WriteToUDP(msg.Packet, ns)

	return
}

//
// Recv will read DNS message from active connection in client into `msg`.
//
func (cl *UDPClient) Recv(msg *Message) (err error) {
	n, _, err := cl.conn.ReadFromUDP(msg.Packet)
	if err != nil {
		return
	}

	msg.Packet = append(msg.Packet[:0], msg.Packet[:n]...)

	if debugLevel >= 2 {
		libbytes.PrintHex(">>> msg.Packet:", msg.Packet, 8)
	}

	return
}
