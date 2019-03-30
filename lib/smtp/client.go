// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
)

//
// Client for SMTP.
//
type Client struct {
	// ServerInfo contains the server information, from the response of
	// EHLO command.
	ServerInfo *ServerInfo

	data []byte
	buf  bytes.Buffer

	serverName string
	raddr      *net.TCPAddr
	conn       net.Conn
	insecure   bool
	isTLS      bool
	isStartTLS bool
}

//
// NewClient create and initialize connection to remote SMTP server.
//
// The localName define the client domain address, used when issuing EHLO
// command to server.  If its empty, it will set to current operating system's
// hostname.
//
// The remoteURL use the following format,
//
//	remoteURL = [ scheme "://" ](domain | IP-address [":" port])
//	scheme    = "smtp" / "smtps" / "smtp+starttls"
//
// If scheme is "smtp" and no port is given, client will connect to remote
// address at port 25.
// If scheme is "smtps" and no port is given, client will connect to remote
// address at port 465 (implicit TLS).
// If scheme is "smtp+starttls" and no port is given, client will connect to
// remote address at port 587.
//
// The "insecure" parameter, if set to true, will disable verifying
// remote certificate when connecting with TLS or STARTTLS.
//
// On success, it will return connected client, with implicit EHLO command
// issued to server immediately.  If scheme is "smtp+starttls", the connection
// also automatically upgraded to TLS after EHLO command success.
//
// On fail, it will return nil client with an error.
//
func NewClient(localName, remoteURL string, insecure bool) (cl *Client, err error) {
	var (
		rurl   *url.URL
		port   uint16
		scheme string
	)

	rurl, err = url.Parse(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("smtp: NewClient: " + err.Error())
	}

	cl = &Client{
		raddr: &net.TCPAddr{},
	}

	scheme = strings.ToLower(rurl.Scheme)
	switch scheme {
	case "smtp":
		port = 25
	case "smtps":
		port = 465
		cl.isTLS = true
	case "smtp+starttls":
		port = 587
		cl.isStartTLS = true
	default:
		return nil, fmt.Errorf("smtp: NewClient: invalid scheme '%s'", scheme)
	}

	cl.serverName, cl.raddr.IP, port = libnet.ParseIPPort(rurl.Host, port)
	if cl.raddr.IP == nil {
		cl.raddr.IP, err = lookup(cl.serverName)
		if err != nil {
			return nil, err
		}
		if cl.raddr.IP == nil {
			err = fmt.Errorf("smtp: NewClient: '%s' does not have MX record or IP address", cl.serverName)
			return nil, err
		}
	}
	if len(cl.serverName) == 0 {
		cl.insecure = true
	}

	cl.data = make([]byte, 4096)
	cl.raddr.Port = int(port)

	if debug.Value >= 3 {
		fmt.Printf("smtp: NewClient remote address '%v'\n", cl.raddr)
	}

	_, err = cl.connect(localName)
	if err != nil {
		return nil, err
	}

	if debug.Value >= 3 {
		fmt.Printf("smtp: ServerInfo: %+v\n", cl.ServerInfo)
	}

	return cl, nil
}

//
// Authenticate to server using one of SASL mechanism.
// Currently, the only mechanism available is PLAIN.
//
func (cl *Client) Authenticate(mech Mechanism, username, password string) (
	res *Response, err error,
) {
	var cmd []byte

	switch mech {
	case MechanismPLAIN:
		b := []byte("\x00" + username + "\x00" + password)
		initialResponse := base64.StdEncoding.EncodeToString(b)
		cmd = []byte("AUTH PLAIN " + initialResponse + "\r\n")
	default:
		return nil, fmt.Errorf("client.Authenticate: unknown mechanism")
	}

	return cl.SendCommand(cmd)
}

//
// connect open a connection to server and issue EHLO command immediately.
//
// If remoteURL scheme is "smtp+starttls", the client will issue STARTTLS
// command immediately after connect.
//
func (cl *Client) connect(localName string) (res *Response, err error) {
	cl.conn, err = net.DialTCP("tcp", nil, cl.raddr)
	if err != nil {
		return nil, err
	}

	if cl.isTLS {
		tlsConfig := &tls.Config{
			ServerName:         cl.serverName,
			InsecureSkipVerify: cl.insecure, // nolint: gosec
		}

		cl.conn = tls.Client(cl.conn, tlsConfig)
	}

	res, err = cl.recv()
	if err != nil {
		return res, err
	}
	if res.Code != StatusReady {
		return res, fmt.Errorf("server return %d, want 220", res.Code)
	}

	res, err = cl.ehlo(localName)
	if err != nil {
		return res, err
	}

	if cl.isStartTLS {
		return cl.StartTLS()
	}

	return res, nil
}

//
// ehlo initialize the SMTP session by sending the EHLO command to server.
// If server does not support EHLO it would return an error, there is no
// fallback to HELO.
//
// Client MUST use localName that resolved to DNS A RR (address) (RFC 5321,
// section 2.3.5), or SHOULD use IP address if not possible (RFC 5321, section
// 4.1.4).
//
func (cl *Client) ehlo(localName string) (res *Response, err error) {
	if len(localName) == 0 {
		localName, err = os.Hostname()
		if err != nil {
			localName = getUnicastAddress()
			if len(localName) == 0 {
				err = errors.New("smtp: unable to get hostname or unicast address")
				return nil, err
			}

			localName = "[" + localName + "]"
		}
	}

	req := []byte("EHLO " + localName + "\r\n")
	res, err = cl.SendCommand(req)
	if err != nil {
		return nil, err
	}

	if res.Code == StatusOK {
		cl.ServerInfo = NewServerInfo(res)
		return res, nil
	}

	err = fmt.Errorf("smtp: EHLO response code %d, want 250", res.Code)

	return res, err
}

//
// Expand get members of mailing-list.
//
func (cl *Client) Expand(mlist string) (res *Response, err error) {
	if len(mlist) == 0 {
		return nil, nil
	}
	cmd := []byte("EXPN " + mlist + "\r\n")
	return cl.SendCommand(cmd)
}

//
// Help get information on specific command from server.
//
func (cl *Client) Help(cmdName string) (res *Response, err error) {
	cmd := []byte("HELP " + cmdName + "\r\n")
	return cl.SendCommand(cmd)
}

//
// Quit signal the server that the client will close the connection.
//
func (cl *Client) Quit() (res *Response, err error) {
	_, err = cl.conn.Write([]byte("QUIT\r\n"))
	if err == nil {
		res, err = cl.recv()
	}

	_ = cl.conn.Close()

	return res, err
}

//
// MailTx send the mail to server.
// This function is implementation of mail transaction (MAIL, RCPT, and DATA
// commands as described in RFC 5321, section 3.3).
// The MailTx.Data must be internet message format which contains headers and
// content as defined by RFC 5322.
//
// On success, it will return the last response, which is the success status
// of data transaction (250).
//
// On fail, it will return response from the failed command with error is
// string combination of command, response code and message.
//
func (cl *Client) MailTx(mail *MailTx) (res *Response, err error) {
	if mail == nil {
		// No operation.
		return nil, nil
	}
	if len(mail.From) == 0 {
		return nil, errors.New("SendMailTx: empty mail 'From' parameter")
	}
	if len(mail.Recipients) == 0 {
		return nil, errors.New("SendMailTx: empty mail 'Recipients' parameter")
	}

	cl.buf.Reset()
	fmt.Fprintf(&cl.buf, "MAIL FROM:<%s>\r\n", mail.From)

	res, err = cl.SendCommand(cl.buf.Bytes())
	if err != nil {
		return nil, err
	}
	if res.Code != StatusOK {
		err = fmt.Errorf("client.MailTx: MAIL FROM: %d - %s", res.Code, res.Message)
		return nil, err
	}

	for _, to := range mail.Recipients {
		cl.buf.Reset()
		fmt.Fprintf(&cl.buf, "RCPT TO:<%s>\r\n", to)

		res, err = cl.SendCommand(cl.buf.Bytes())
		if err != nil || res.Code != StatusOK {
			err = fmt.Errorf("client.MailTx: RCPT TO: %d - %s", res.Code, res.Message)
			return res, err
		}
	}

	cl.buf.Reset()
	cl.buf.WriteString("DATA\r\n")

	res, err = cl.SendCommand(cl.buf.Bytes())
	if err != nil || res.Code != StatusDataReady {
		err = fmt.Errorf("client.MailTx: DATA: %d - %s", res.Code, res.Message)
		return res, err
	}

	cl.buf.Reset()
	cl.buf.Write(mail.Data)
	cl.buf.WriteString("\r\n.\r\n")

	_, err = cl.conn.Write(cl.buf.Bytes())
	if err != nil {
		return nil, err
	}

	res, err = cl.recv()
	if err != nil || res.Code != StatusOK {
		err = fmt.Errorf("client.MailTx: Message: %d - %s", res.Code, res.Message)
	}

	return res, err
}

//
// SendCommand send any custom command to server.
//
func (cl *Client) SendCommand(cmd []byte) (res *Response, err error) {
	if debug.Value >= 3 {
		fmt.Printf(">>> smtp: Client.SendCommand: %s", cmd)
	}

	_, err = cl.conn.Write(cmd)
	if err != nil {
		return nil, err
	}

	return cl.recv()
}

//
// Verify send the VRFY command to server to check if mailbox is exist.
//
func (cl *Client) Verify(mailbox string) (res *Response, err error) {
	if len(mailbox) == 0 {
		return nil, nil
	}
	cmd := []byte("VRFY " + mailbox + "\r\n")
	return cl.SendCommand(cmd)
}

//
// The remote address can be a hostname or IP address with port.
// If its a host name, the client will try to lookup the MX record first, if
// its fail it will resolve the IP address and use it.
//
func lookup(address string) (ip net.IP, err error) {
	mxs, err := net.LookupMX(address)
	if err == nil && len(mxs) > 0 {
		// Select the lowest MX preferences.
		pref := uint16(65535)
		for _, mx := range mxs {
			if mx.Pref < pref {
				address = mx.Host
			}
		}
	}

	ips, err := net.LookupIP(address)
	if err != nil {
		return nil, err
	}

	if len(ips) > 0 {
		return ips[0], nil
	}

	return nil, nil
}

//
// getUnicastAddress return the local unicast address other than localhost.
//
func getUnicastAddress() (saddr string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		saddr = addr.String()
		if strings.HasSuffix(saddr, "127") {
			continue
		}
		return saddr
	}
	return ""
}

//
// recv read and parse the response from server.
//
func (cl *Client) recv() (res *Response, err error) {
	cl.buf.Reset()

	for {
		n, err := cl.conn.Read(cl.data)
		if n > 0 {
			_, _ = cl.buf.Write(cl.data[:n])
		}
		if err != nil {
			return nil, err
		}
		if n == cap(cl.data) {
			continue
		}
		break
	}

	if debug.Value >= 3 {
		fmt.Printf("<<< smtp: Client.recv: %s", cl.buf.Bytes())
	}

	res, err = NewResponse(cl.buf.Bytes())
	if err != nil {
		return nil, err
	}

	return res, nil
}

//
// StartTLS upgrade the underlying connection to TLS.  This method only works
// if client connected to remote URL using scheme "smtp+starttls" or on port
// 587, and on server that support STARTTLS extension.
//
func (cl *Client) StartTLS() (res *Response, err error) {
	req := []byte("STARTTLS\r\n")
	res, err = cl.SendCommand(req)
	if err != nil {
		return nil, err
	}

	if res.Code != StatusReady {
		return nil, fmt.Errorf("smtp: STARTTLS response %d, want 220",
			res.Code)
	}

	tlsConfig := &tls.Config{
		ServerName:         cl.serverName,
		InsecureSkipVerify: cl.insecure, // nolint: gosec
	}

	cl.conn = tls.Client(cl.conn, tlsConfig)

	return res, nil
}
