// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"net"
	"sync"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
)

//
// UDPClient for DNS with UDP connection.
//
// Any implementation that need to query DNS message in multiple Go routines
// should create one client per routine.
//
type UDPClient struct {
	timeout time.Duration
	// addr contains address of remote connection.
	addr *net.UDPAddr
	conn *net.UDPConn
	sync.Mutex
}

//
// NewUDPClient will create new DNS client with UDP network connection.
//
// The nameserver contains the IP address, not host name, of parent DNS
// server.  Default port is 53, if not set.
//
func NewUDPClient(nameserver string) (cl *UDPClient, err error) {
	network := "udp"

	_, remoteIP, remotePort := libnet.ParseIPPort(nameserver, DefaultPort)
	if remoteIP == nil {
		return nil, fmt.Errorf("dns: invalid address '%s'", nameserver)
	}

	laddr := &net.UDPAddr{IP: nil, Port: 0}
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		return
	}

	cl = &UDPClient{
		timeout: clientTimeout,
		addr: &net.UDPAddr{
			IP:   remoteIP,
			Port: int(remotePort),
		},
		conn: conn,
	}

	return
}

//
// RemoteAddr return client remote nameserver address.
//
func (cl *UDPClient) RemoteAddr() string {
	return cl.addr.String()
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
func (cl *UDPClient) Lookup(
	allowRecursion bool, qtype, qclass uint16, qname string,
) (
	*Message, error,
) {
	if cl.addr == nil || cl.conn == nil {
		return nil, nil
	}

	msg := NewMessage()

	msg.Header.ID = getNextID()
	msg.Header.IsRD = allowRecursion
	msg.Header.QDCount = 1
	msg.Question.Type = qtype
	msg.Question.Class = qclass
	msg.Question.Name = qname

	_, _ = msg.Pack()

	res, err := cl.Query(msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//
// Query send DNS query to name server "ns" and return the unpacked response.
//
func (cl *UDPClient) Query(req *Message) (res *Message, err error) {
	logp := "Query"
	cl.Lock()
	defer cl.Unlock()

	_, err = cl.Write(req.packet)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	err = cl.conn.SetReadDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	packet := make([]byte, maxUdpPacketSize)

	n, _, err := cl.conn.ReadFromUDP(packet)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	res = &Message{
		packet: packet[:n],
	}

	if debug.Value >= 3 {
		libbytes.PrintHex(">>> UDPClient.recv:", res.packet, 8)
	}

	err = res.Unpack()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return res, nil
}

//
// Write raw DNS response message on active connection.
// This method is only used by server to write the response of query to
// client.
//
func (cl *UDPClient) Write(msg []byte) (n int, err error) {
	err = cl.conn.SetWriteDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return
	}

	return cl.conn.WriteToUDP(msg, cl.addr)
}

//
// SetRemoteAddr set the remote address for sending the packet.
//
func (cl *UDPClient) SetRemoteAddr(addr string) (err error) {
	cl.addr, err = net.ResolveUDPAddr("udp", addr)
	return
}

//
// SetTimeout for sending and receiving packet.
//
func (cl *UDPClient) SetTimeout(t time.Duration) {
	cl.timeout = t
}
