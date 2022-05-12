// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
)

// DoTClient client for DNS over TLS.
type DoTClient struct {
	conn    *tls.Conn
	timeout time.Duration
}

// NewDoTClient will create new DNS client over TLS connection.
//
// The nameserver contains the IP address, not host name, of parent DNS
// server.  Default port is 853, if not set.
func NewDoTClient(nameserver string, allowInsecure bool) (cl *DoTClient, err error) {
	var (
		tlsConfig tls.Config
		remoteIP  net.IP
		port      uint16
	)

	_, remoteIP, port = libnet.ParseIPPort(nameserver, DefaultTLSPort)
	if remoteIP == nil {
		return nil, fmt.Errorf("dns: invalid address '%s'", nameserver)
	}

	cl = &DoTClient{
		timeout: clientTimeout,
	}

	nameserver = fmt.Sprintf("%s:%d", remoteIP, port)

	tlsConfig.InsecureSkipVerify = allowInsecure

	cl.conn, err = tls.Dial("tcp", nameserver, &tlsConfig)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

// Close the client connection.
func (cl *DoTClient) Close() error {
	if cl.conn != nil {
		return cl.conn.Close()
	}
	return nil
}

// Lookup DNS records based on MessageQuestion Name and Type, in synchronous
// mode.
// The MessageQuestion Class default to IN.
//
// It will return an error if the Name is empty.
func (cl *DoTClient) Lookup(q MessageQuestion, allowRecursion bool) (res *Message, err error) {
	if len(q.Name) == 0 {
		return nil, fmt.Errorf("Lookup: empty question name")
	}
	if q.Type == 0 {
		q.Type = RecordTypeA
	}
	if q.Class == 0 {
		q.Class = RecordClassIN
	}

	var (
		msg *Message = NewMessage()
	)

	msg.Header.ID = getNextID()
	msg.Header.IsRD = allowRecursion
	msg.Header.QDCount = 1
	msg.Question = q

	_, err = msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("Lookup: %w", err)
	}

	res, err = cl.Query(msg)
	if err != nil {
		return nil, fmt.Errorf("Lookup: %w", err)
	}

	return res, nil
}

// Query send DNS Message to name server.
func (cl *DoTClient) Query(msg *Message) (res *Message, err error) {
	_, err = cl.Write(msg.packet)
	if err != nil {
		return nil, err
	}

	res = NewMessage()

	_, err = cl.recv(res)
	if err != nil {
		return nil, err
	}

	err = res.Unpack()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// RemoteAddr return client remote nameserver address.
func (cl *DoTClient) RemoteAddr() string {
	return cl.conn.RemoteAddr().String()
}

// recv will read DNS message from active connection in client into `msg`.
func (cl *DoTClient) recv(msg *Message) (n int, err error) {
	err = cl.conn.SetReadDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return
	}

	var packet = make([]byte, maxTcpPacketSize)

	n, err = cl.conn.Read(packet)
	if err != nil {
		return
	}
	if n == 0 {
		return
	}

	msg.packet = packet[2:n]

	if debug.Value >= 3 {
		libbytes.PrintHex(">>> DoTClient: recv: ", msg.packet, 8)
	}

	return
}

// Write raw DNS message on active connection.
func (cl *DoTClient) Write(msg []byte) (n int, err error) {
	err = cl.conn.SetWriteDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return
	}

	var (
		lenmsg int    = len(msg)
		packet []byte = make([]byte, 0, 2+lenmsg)
	)

	packet = libbytes.AppendUint16(packet, uint16(lenmsg))
	packet = append(packet, msg...)

	n, err = cl.conn.Write(packet)

	return
}

// SetRemoteAddr no-op.
func (cl *DoTClient) SetRemoteAddr(addr string) (err error) {
	return
}

// SetTimeout for sending and receiving packet.
func (cl *DoTClient) SetTimeout(t time.Duration) {
	cl.timeout = t
}
