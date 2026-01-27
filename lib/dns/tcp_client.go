// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

// TCPClient for DNS with TCP connection and list of remote addresses.
type TCPClient struct {
	conn         net.Conn
	addr         *net.TCPAddr
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// NewTCPClient will create new DNS client with TCP network connection.
//
// The nameserver contains the IP address, not host name, of parent DNS
// server.  Default port is 53, if not set.
func NewTCPClient(nameserver string) (cl *TCPClient, err error) {
	var (
		raddr = &net.TCPAddr{}

		remoteIP   net.IP
		remotePort uint16
	)

	_, remoteIP, remotePort = libnet.ParseIPPort(nameserver, DefaultPort)
	if remoteIP == nil {
		return nil, fmt.Errorf("dns: invalid address '%s'", nameserver)
	}

	raddr.IP = remoteIP
	raddr.Port = int(remotePort)

	cl = &TCPClient{
		readTimeout:  clientTimeout,
		writeTimeout: clientTimeout,
		addr:         raddr,
	}

	err = cl.Connect(raddr)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

// Close client connection.
func (cl *TCPClient) Close() error {
	if cl.conn != nil {
		return cl.conn.Close()
	}
	return nil
}

// Connect to remote address.
func (cl *TCPClient) Connect(raddr *net.TCPAddr) (err error) {
	var (
		laddr = &net.TCPAddr{IP: nil, Port: 0}
	)

	cl.conn, err = net.DialTCP("tcp", laddr, raddr)

	return
}

// Lookup DNS records based on MessageQuestion Name and Type, in synchronous
// mode.
// The MessageQuestion Class default to IN.
//
// It will return an error if the client does not set the name server address,
// or no connection, or Name is empty.
func (cl *TCPClient) Lookup(q MessageQuestion, allowRecursion bool) (msg *Message, err error) {
	if cl.addr == nil || cl.conn == nil {
		return nil, errors.New(`Lookup: no name server or active connection`)
	}
	if len(q.Name) == 0 {
		return nil, errors.New(`Lookup: empty question name`)
	}
	if q.Type == 0 {
		q.Type = RecordTypeA
	}
	if q.Class == 0 {
		q.Class = RecordClassIN
	}

	msg = NewMessage()

	msg.Header.ID = getNextID()
	msg.Header.IsRD = allowRecursion
	msg.Header.QDCount = 1
	msg.Question = q

	_, err = msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("Lookup: %w", err)
	}

	msg, err = cl.Query(msg)
	if err != nil {
		return nil, fmt.Errorf("Lookup: %w", err)
	}

	return msg, nil
}

// Query send DNS query to name server.
// The addr parameter is unused.
func (cl *TCPClient) Query(msg *Message) (res *Message, err error) {
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
func (cl *TCPClient) RemoteAddr() string {
	return cl.addr.String()
}

// SetRemoteAddr set the remote address for sending the packet.
func (cl *TCPClient) SetRemoteAddr(addr string) (err error) {
	cl.addr, err = net.ResolveTCPAddr("udp", addr)
	return
}

// SetTimeout for sending and receiving packet.
func (cl *TCPClient) SetTimeout(t time.Duration) {
	cl.readTimeout = t
	cl.writeTimeout = t
}

// Write raw DNS response message on active connection.
// This method is only used by server to write the response of query to
// client.
func (cl *TCPClient) Write(msg []byte) (n int, err error) {
	var logp = `Write`

	if cl.writeTimeout > 0 {
		err = cl.conn.SetWriteDeadline(time.Now().Add(cl.writeTimeout))
		if err != nil {
			return 0, fmt.Errorf(`%s: %w`, logp, err)
		}
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

// recv receive DNS message.
func (cl *TCPClient) recv() (packet []byte, err error) {
	var logp = `recv`

	if cl.readTimeout > 0 {
		err = cl.conn.SetReadDeadline(time.Now().Add(cl.readTimeout))
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	var n int

	packet = make([]byte, maxTCPPacketSize)

	n, err = cl.conn.Read(packet)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	if n == 0 {
		return nil, io.EOF
	}
	if n < 2 {
		return nil, fmt.Errorf(`%s: invalid packet`, logp)
	}

	packet = packet[2:n]

	return packet, nil
}
