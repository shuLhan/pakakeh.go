// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"sync"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// Client for DNS with UDP connection and list of remote addresses.
//
type Client struct {
	sync.Mutex
	conn  *net.UDPConn
	ns    []*net.UDPAddr
	nsIdx int
}

//
// NewClient will create new DNS client with list of parent name servers to
// query.
// Name-server should be in IP:port address format, not with a host name.
// Name-server port is optional, e.g. "8.8.8.8:", if it's not defined, then
// default to 53.
//
func NewClient(nameServers []string) (*Client, error) {
	cl := new(Client)

	for x := 0; x < len(nameServers); x++ {
		err := cl.AddRemoteAddress(nameServers[x])
		if err != nil {
			return nil, err
		}
	}

	return cl, nil
}

//
// AddRemoteAddress to list of remote name servers.
//
// This function is safe to be used concurrently.
//
func (cl *Client) AddRemoteAddress(address string) error {
	if len(address) == 0 {
		return nil
	}

	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	if udpAddr.Port == 0 {
		udpAddr.Port = Port
	}

	cl.AddRemoteUDPAddr(udpAddr)

	return nil
}

//
// AddRemoteUDPAddr to list of parent name servers.
//
func (cl *Client) AddRemoteUDPAddr(addr *net.UDPAddr) {
	cl.Lock()
	cl.ns = append(cl.ns, addr)
	cl.Unlock()
}

//
// getRotatedNameServer will return the next name server from the list.  Every
// call to Lookup will rotate the name server.
//
// This function is safe to be used concurrently.
//
func (cl *Client) getRotatedNameServer() *net.UDPAddr {
	cl.Lock()
	cl.nsIdx = cl.nsIdx % len(cl.ns)
	ns := cl.ns[cl.nsIdx]
	cl.nsIdx++
	cl.Unlock()

	return ns
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
func (cl *Client) Lookup(qtype QueryType, qclass uint16, qname []byte) (
	msg *Message, err error,
) {
	connPool := udpConnPool.Get()
	if connPool == nil {
		return nil, ErrNewConnection
	}

	cl.conn = connPool.(*net.UDPConn)

	msg = msgPool.Get().(*Message)
	msg.Reset()
	msg.Header.ID = getID()
	msg.Header.QDCount = 1
	msg.Question.Type = qtype
	msg.Question.Class = qclass
	msg.Question.Name = append(msg.Question.Name, qname...)

	_, _ = msg.MarshalBinary()

	err = cl.Send(msg, nil)
	if err != nil {
		goto out
	}

	msg.Reset()

	err = cl.Recv(msg)
	if err != nil {
		goto out
	}

	err = msg.UnmarshalBinary(msg.Packet)
out:
	udpConnPool.Put(cl.conn)
	cl.conn = nil
	if err != nil {
		msgPool.Put(msg)
		msg = nil
	}

	return
}

//
// Send DNS message to name server using active connection in client.
// If ns is nil it will use one of the name-server in clients.
// The message packet must already been filled, using MarshalBinary.
//
func (cl *Client) Send(msg *Message, ns *net.UDPAddr) error {
	if ns == nil {
		ns = cl.getRotatedNameServer()
	}
	if cl.conn == nil {
		connPool := udpConnPool.Get()
		if connPool == nil {
			return ErrNewConnection
		}
		cl.conn = connPool.(*net.UDPConn)
	}

	_, err := cl.conn.WriteToUDP(msg.Packet, ns)

	return err
}

//
// Recv will read DNS message from active connection in client into `msg`.
//
func (cl *Client) Recv(msg *Message) error {
	n, _, err := cl.conn.ReadFromUDP(msg.Packet)
	if err != nil {
		return err
	}

	msg.Packet = msg.Packet[:n]

	if debugLevel >= 2 {
		libbytes.PrintHex(">>> msg.Packet:", msg.Packet, 8)
	}

	return nil
}
