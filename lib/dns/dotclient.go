// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
	"fmt"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
)

//
// DoTClient client for DNS over TLS.
//
type DoTClient struct {
	timeout time.Duration
	conn    *tls.Conn
}

//
// NewDoTClient will create new DNS client over TLS connection.
//
// The nameserver contains the IP address, not host name, of parent DNS
// server.  Default port is 853, if not set.
//
func NewDoTClient(nameserver string, allowInsecure bool) (cl *DoTClient, err error) {
	_, remoteIP, port := libnet.ParseIPPort(nameserver, DefaultTLSPort)
	if remoteIP == nil {
		return nil, fmt.Errorf("dns: invalid address '%s'", nameserver)
	}

	cl = &DoTClient{
		timeout: clientTimeout,
	}

	nameserver = fmt.Sprintf("%s:%d", remoteIP, port)

	tlsConfig := tls.Config{
		InsecureSkipVerify: allowInsecure,
	}

	cl.conn, err = tls.Dial("tcp", nameserver, &tlsConfig)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

//
// Close the client connection.
//
func (cl *DoTClient) Close() error {
	if cl.conn != nil {
		return cl.conn.Close()
	}
	return nil
}

//
// Lookup specific type, class, and name in synchronous mode.
//
func (cl *DoTClient) Lookup(allowRecursion bool, qtype, qclass uint16, qname []byte) (
	res *Message, err error,
) {
	if len(qname) == 0 {
		return nil, nil
	}
	if qtype == 0 {
		qtype = QueryTypeA
	}
	if qclass == 0 {
		qclass = QueryClassIN
	}

	msg := NewMessage()

	msg.Header.ID = getNextID()
	msg.Header.IsRD = allowRecursion
	msg.Header.QDCount = 1
	msg.Question.Type = qtype
	msg.Question.Class = qclass
	msg.Question.Name = append(msg.Question.Name, qname...)

	_, _ = msg.Pack()

	res, err = cl.Query(msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//
// Query send DNS Message to name server.
//
func (cl *DoTClient) Query(msg *Message) (*Message, error) {
	_, err := cl.Write(msg.Packet)
	if err != nil {
		return nil, err
	}

	res := NewMessage()

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

//
// RemoteAddr return client remote nameserver address.
//
func (cl *DoTClient) RemoteAddr() string {
	return cl.conn.RemoteAddr().String()
}

//
// recv will read DNS message from active connection in client into `msg`.
//
func (cl *DoTClient) recv(msg *Message) (n int, err error) {
	err = cl.conn.SetReadDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return
	}

	packet := make([]byte, maxUDPPacketSize)

	n, err = cl.conn.Read(packet)
	if err != nil {
		return
	}
	if n == 0 {
		return
	}

	msg.Packet = libbytes.Copy(packet[2:n])

	if debug.Value >= 3 {
		libbytes.PrintHex(">>> DoTClient: recv: ", msg.Packet, 8)
	}

	return
}

//
// Write raw DNS message on active connection.
//
func (cl *DoTClient) Write(msg []byte) (n int, err error) {
	err = cl.conn.SetWriteDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return
	}

	lenmsg := len(msg)
	packet := make([]byte, 0, 2+lenmsg)

	libbytes.AppendUint16(&packet, uint16(lenmsg))
	packet = append(packet, msg...)

	n, err = cl.conn.Write(packet)

	return
}

//
// SetRemoteAddr no-op.
//
func (cl *DoTClient) SetRemoteAddr(addr string) (err error) {
	return
}

//
// SetTimeout for sending and receiving packet.
//
func (cl *DoTClient) SetTimeout(t time.Duration) {
	cl.timeout = t
}
