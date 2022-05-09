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
	"io"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
)

// Client for SMTP.
type Client struct {
	opts ClientOptions

	// ServerInfo contains the server information, from the response of
	// EHLO command.
	ServerInfo *ServerInfo

	conn       net.Conn
	raddr      net.TCPAddr
	serverName string

	data []byte
	buf  bytes.Buffer

	isTLS      bool
	isStartTLS bool
}

// NewClient create and initialize connection to remote SMTP server.
//
// When connected, the client send implicit EHLO command issued to server
// immediately.
// If scheme is "smtp+starttls", the connection automatically upgraded to
// TLS after EHLO command success.
//
// If both AuthUser and AuthPass in the ClientOptions is not empty, the client
// will try to authenticate to remote server.
//
// On fail, it will return nil client with an error.
func NewClient(opts ClientOptions) (cl *Client, err error) {
	var (
		logp = "NewClient"

		res  *Response
		rurl *url.URL
		port uint16
	)

	rurl, err = url.Parse(opts.ServerUrl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	cl = &Client{
		opts: opts,
	}

	rurl.Scheme = strings.ToLower(rurl.Scheme)
	switch rurl.Scheme {
	case "smtp":
		port = 25
	case "smtps":
		port = 465
		cl.isTLS = true
	case "smtp+starttls":
		port = 587
		cl.isStartTLS = true
	default:
		return nil, fmt.Errorf("%s: invalid server URL scheme", logp)
	}

	cl.serverName, cl.raddr.IP, port = libnet.ParseIPPort(rurl.Host, port)
	if cl.raddr.IP == nil {
		cl.raddr.IP, err = lookup(cl.serverName)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}
		if cl.raddr.IP == nil {
			return nil, fmt.Errorf("%s: %q does not have MX record or IP address", logp, cl.serverName)
		}
	}

	cl.data = make([]byte, 4096)
	cl.raddr.Port = int(port)

	if debug.Value >= 3 {
		fmt.Printf("%s: remote address is %v\n", logp, cl.raddr)
	}

	_, err = cl.connect(opts.LocalName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if debug.Value >= 3 {
		fmt.Printf("%s: ServerInfo: %+v\n", logp, cl.ServerInfo)
	}

	if len(opts.AuthUser) == 0 || len(opts.AuthPass) == 0 {
		// Do not authenticate this connection, yet.
		return cl, nil
	}

	res, err = cl.Authenticate(cl.opts.AuthMechanism, cl.opts.AuthUser, cl.opts.AuthPass)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.Code != StatusAuthenticated {
		return nil, fmt.Errorf("%s: %d %s", logp, res.Code, res.Message)
	}

	return cl, nil
}

// Authenticate to server using one of SASL mechanism.
// Currently, the only mechanism available is PLAIN.
func (cl *Client) Authenticate(mech SaslMechanism, username, password string) (
	res *Response, err error,
) {
	var cmd []byte

	switch mech {
	case SaslMechanismPlain:
		b := []byte("\x00" + username + "\x00" + password)
		initialResponse := base64.StdEncoding.EncodeToString(b)
		cmd = []byte("AUTH PLAIN " + initialResponse + "\r\n")
	default:
		return nil, fmt.Errorf("client.Authenticate: unknown mechanism")
	}

	return cl.SendCommand(cmd)
}

// connect open a connection to server and issue EHLO command immediately.
//
// If remoteURL scheme is "smtp+starttls", the client will issue STARTTLS
// command immediately after connect.
func (cl *Client) connect(localName string) (res *Response, err error) {
	logp := "connect"

	cl.conn, err = net.DialTCP("tcp", nil, &cl.raddr)
	if err != nil {
		return nil, err
	}

	if cl.isTLS {
		tlsConfig := &tls.Config{
			ServerName:         cl.serverName,
			InsecureSkipVerify: cl.opts.Insecure,
		}

		cl.conn = tls.Client(cl.conn, tlsConfig)
	}

	if debug.Value >= 3 {
		fmt.Println(">>> Connected ...")
	}

	res, err = cl.recv()
	if err != nil {
		return res, fmt.Errorf("%s: %w", logp, err)
	}
	if res.Code != StatusReady {
		return res, fmt.Errorf("%s: server return %d, want 220", logp, res.Code)
	}

	res, err = cl.ehlo(localName)
	if err != nil {
		return res, fmt.Errorf("%s: %w", logp, err)
	}

	if cl.isStartTLS {
		return cl.StartTLS()
	}

	return res, nil
}

// ehlo initialize the SMTP session by sending the EHLO command to server.
// If server does not support EHLO it would return an error, there is no
// fallback to HELO.
//
// Client MUST use localName that resolved to DNS A RR (address) (RFC 5321,
// section 2.3.5), or SHOULD use IP address if not possible (RFC 5321, section
// 4.1.4).
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

// Expand get members of mailing-list.
func (cl *Client) Expand(mlist string) (res *Response, err error) {
	if len(mlist) == 0 {
		return nil, nil
	}
	cmd := []byte("EXPN " + mlist + "\r\n")
	return cl.SendCommand(cmd)
}

// Help get information on specific command from server.
func (cl *Client) Help(cmdName string) (res *Response, err error) {
	cmd := []byte("HELP " + cmdName + "\r\n")
	return cl.SendCommand(cmd)
}

// Quit signal the server that the client will close the connection.
func (cl *Client) Quit() (res *Response, err error) {
	_, err = cl.conn.Write([]byte("QUIT\r\n"))
	if err == nil {
		res, err = cl.recv()
	}

	errClose := cl.conn.Close()
	if errClose != nil {
		if err == nil {
			err = errClose
		}
	}

	return res, err
}

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

// Noop send the NOOP command to server with optional message.
//
// On success, it will return response with Code 250, StatusOK.
func (cl *Client) Noop(msg string) (res *Response, err error) {
	var cmd string
	if len(msg) > 0 {
		cmd = fmt.Sprintf("NOOP %s\r\n", msg)
	} else {
		cmd = "NOOP\r\n"
	}
	return cl.SendCommand([]byte(cmd))
}

// Reset send the RSET command to server.
// This command clear the current buffer on MAIL, RCPT, and DATA, but not the
// EHLO/HELO buffer.
//
// On success, it will return response with Code 250, StatusOK.
func (cl *Client) Reset() (res *Response, err error) {
	cmd := []byte("RSET\r\n")
	return cl.SendCommand(cmd)
}

// SendCommand send any custom command to server.
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

// SendEmail is the wrapper that simplify MailTx.
func (cl *Client) SendEmail(from string, to []string, subject, bodyText, bodyHtml []byte) (err error) {
	return nil
}

// Verify send the VRFY command to server to check if mailbox is exist.
func (cl *Client) Verify(mailbox string) (res *Response, err error) {
	if len(mailbox) == 0 {
		return nil, nil
	}
	cmd := []byte("VRFY " + mailbox + "\r\n")
	return cl.SendCommand(cmd)
}

// The remote address can be a hostname or IP address with port.
// If its a host name, the client will try to lookup the MX record first, if
// its fail it will resolve the IP address and use it.
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

// getUnicastAddress return the local unicast address other than localhost.
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

// recv read and parse the response from server.
func (cl *Client) recv() (res *Response, err error) {
	cl.buf.Reset()

	for {
		n, err := cl.conn.Read(cl.data)
		if n > 0 {
			_, err = cl.buf.Write(cl.data[:n])
			if err != nil {
				break
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				if cl.buf.Len() == 0 {
					return nil, err
				}
				break
			}
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
	if res == nil {
		return nil, io.EOF
	}

	return res, nil
}

// StartTLS upgrade the underlying connection to TLS.  This method only works
// if client connected to remote URL using scheme "smtp+starttls" or on port
// 587, and on server that support STARTTLS extension.
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
		InsecureSkipVerify: cl.opts.Insecure,
	}

	cl.conn = tls.Client(cl.conn, tlsConfig)

	return res, nil
}
