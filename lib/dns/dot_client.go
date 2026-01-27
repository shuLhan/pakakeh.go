// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
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
		return nil, errors.New(`Lookup: empty question name`)
	}
	if q.Type == 0 {
		q.Type = RecordTypeA
	}
	if q.Class == 0 {
		q.Class = RecordClassIN
	}

	var msg = NewMessage()

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
	var logp = `Query`

	_, err = cl.Write(msg.packet)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var packet []byte

	packet, err = cl.recv()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	res, err = UnpackMessage(packet)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return res, nil
}

// RemoteAddr return client remote nameserver address.
func (cl *DoTClient) RemoteAddr() string {
	return cl.conn.RemoteAddr().String()
}

// recv will read DNS message from active connection.
func (cl *DoTClient) recv() (packet []byte, err error) {
	var logp = `recv`

	err = cl.conn.SetReadDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var n int

	packet = make([]byte, maxTCPPacketSize)

	n, err = cl.conn.Read(packet)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	if n == 0 {
		return packet, nil
	}

	packet = packet[2:n]

	return packet, nil
}

// Write raw DNS message on active connection.
func (cl *DoTClient) Write(msg []byte) (n int, err error) {
	var logp = `Write`

	err = cl.conn.SetWriteDeadline(time.Now().Add(cl.timeout))
	if err != nil {
		return 0, fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		lenmsg = uint16(len(msg))
		packet = make([]byte, 2+lenmsg)
	)

	packet[0] = byte(lenmsg >> 8)
	packet[1] = byte(lenmsg)
	copy(packet[2:], msg)

	n, err = cl.conn.Write(packet)
	if err != nil {
		return 0, fmt.Errorf(`%s: %w`, logp, err)
	}

	return n, nil
}

// SetRemoteAddr no-op.
func (cl *DoTClient) SetRemoteAddr(_ string) (err error) {
	return
}

// SetTimeout for sending and receiving packet.
func (cl *DoTClient) SetTimeout(t time.Duration) {
	cl.timeout = t
}
