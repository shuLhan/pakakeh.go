// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	_defPort            = "80"
	_defPortSecure      = "443"
	_netNameTCP         = "tcp"
	_schemeWS           = "ws"
	_schemeWSS          = "wss"
	_handshakeReqFormat = "GET %s HTTP/1.1\r\n" +
		"Host: %s\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Key: %s\r\n" +
		"Sec-Websocket-Version: 13\r\n"
)

var (
	_defRWTO = 10 * time.Second //nolint: gochecknoglobals
)

type ctxKey int

const (
	ctxKeyWSAccept ctxKey = 1
)

//
// ClientRecvHandler define a custom callback type for handling response from
// request.
//
type ClientRecvHandler func(ctx context.Context, resp []byte) (err error)

//
// Client for websocket.
//
type Client struct {
	state           ConnState
	remoteURL       *url.URL
	remoteAddr      string
	handshakeHeader http.Header
	conn            net.Conn
	bb              bytes.Buffer
	isTLS           bool
}

//
// NewClient create a new client connection to websocket server with a
// handshake.
//
// The endpoint use the following format,
//
//	ws-URI = "ws:" "//" host [ ":" port ] path [ "?" query ]
//	wss-URI = "wss:" "//" host [ ":" port ] path [ "?" query ]
//
// The port component is OPTIONAL; the default for "ws" is port 80, while the
// default for "wss" is port 443.
//
// The headers parameter can be used to pass custom headers, except primary
// header fields ("host", "upgrade", "connection", "sec-websocket-key",
// "sec-websocket-version") will be deleted.
//
// On success it will return client thats connected to endpoint.
// On fail it will return nil and error.
//
func NewClient(endpoint string, headers http.Header) (cl *Client, err error) {
	cl = &Client{}

	err = cl.parseURI(endpoint)
	if err != nil {
		return nil, fmt.Errorf("websocket: NewClient: " + err.Error())
	}

	if len(headers) > 0 {
		headers.Del(_hdrKeyHost)
		headers.Del(_hdrKeyUpgrade)
		headers.Del(_hdrKeyOrigin)
		headers.Del(_hdrKeyWSKey)
		headers.Del(_hdrKeyWSVersion)
	}

	cl.handshakeHeader = headers

	err = cl.connect()
	if err != nil {
		return nil, fmt.Errorf("websocket: NewClient: " + err.Error())
	}

	return cl, nil
}

//
// parseURI parse websocket connection URI from "endpoint" and set the
// remoteURL.
//
// On success it will set the remote address that can be used on open().
// On fail it will return an error.
//
func (cl *Client) parseURI(endpoint string) (err error) {
	cl.remoteURL, err = url.ParseRequestURI(endpoint)
	if err != nil {
		cl = nil
		return
	}

	if cl.remoteURL.Scheme == _schemeWSS {
		cl.isTLS = true
	}

	cl.parseRemoteAddr()

	return
}

//
// parseRemoteAddr parse "host:port" from value in remote URL. By default, if
// no port is given, it will set to 80 or 433, depends on URL scheme.
//
func (cl *Client) parseRemoteAddr() {
	serverPort := cl.remoteURL.Port()

	if len(serverPort) != 0 {
		cl.remoteAddr = cl.remoteURL.Host
		return
	}

	switch cl.remoteURL.Scheme {
	case _schemeWS:
		serverPort = _defPort
	case _schemeWSS:
		serverPort = _defPortSecure
	default:
		serverPort = _defPort
	}

	cl.remoteAddr = cl.remoteURL.Hostname() + ":" + serverPort
}

//
// open TCP connection to websocket remote address.
// If client "isTLS" field is true, the connection is opened with TLS protocol
// and the remote name MUST have a valid certificate.
//
func (cl *Client) open() (err error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	if cl.isTLS {
		cfg := &tls.Config{
			InsecureSkipVerify: cl.isTLS, //nolint:gas
		}

		cl.conn, err = tls.DialWithDialer(dialer, _netNameTCP,
			cl.remoteAddr, cfg)
	} else {
		cl.conn, err = dialer.Dial(_netNameTCP, cl.remoteAddr)
	}
	if err != nil {
		return fmt.Errorf("websocket: open: " + err.Error())
	}

	cl.state = ConnStateOpen

	return
}

//
// handshake send the websocket opening handshake.
//
func (cl *Client) handshake() (err error) {
	cl.bb.Reset()
	path := cl.remoteURL.EscapedPath() + "?" + cl.remoteURL.RawQuery
	key := generateHandshakeKey()
	keyAccept := generateHandshakeAccept(key)

	_, err = fmt.Fprintf(&cl.bb, _handshakeReqFormat, path, cl.remoteURL.Host, key)
	if err != nil {
		return err
	}

	if len(cl.handshakeHeader) > 0 {
		err = cl.handshakeHeader.Write(&cl.bb)
		if err != nil {
			return err
		}
	}

	cl.bb.Write([]byte{'\r', '\n'})

	ctx := context.WithValue(context.Background(), ctxKeyWSAccept, keyAccept)

	return cl.Send(ctx, cl.bb.Bytes(), cl.handleHandshake)
}

func (cl *Client) handleHandshake(ctx context.Context, resp []byte) (err error) {
	httpBuf := bufio.NewReader(bytes.NewBuffer(resp))

	httpRes, err := http.ReadResponse(httpBuf, nil)
	if err != nil {
		err = fmt.Errorf("websocket: client.handleHandshake: " + err.Error())
		cl.state = ConnStateError
		return err
	}

	if httpRes.StatusCode != http.StatusSwitchingProtocols {
		err = fmt.Errorf("websocket: client.handleHandshake: " + httpRes.Status)
		cl.state = ConnStateError
		httpRes.Body.Close()
		return err
	}

	expAccept := ctx.Value(ctxKeyWSAccept)
	gotAccept := httpRes.Header.Get(_hdrKeyWSAccept)
	if expAccept != gotAccept {
		err = fmt.Errorf("websocket: client.handleHandshake: invalid server accept key")
		cl.state = ConnStateError
		httpRes.Body.Close()
		return err
	}

	cl.state = ConnStateConnected
	httpRes.Body.Close()

	return
}

//
// connect to server remote address and handshake parameters.
//
func (cl *Client) connect() (err error) {
	if cl.conn != nil {
		_ = cl.conn.Close()
	}

	err = cl.open()
	if err != nil {
		return
	}

	err = cl.handshake()

	return
}

//
// Send message to server.
//
func (cl *Client) Send(ctx context.Context, req []byte, handler ClientRecvHandler) (err error) {
	if len(req) == 0 {
		return
	}

	err = cl.conn.SetWriteDeadline(time.Now().Add(_defRWTO))
	if err != nil {
		return
	}

	_, err = cl.conn.Write(req)
	if err != nil {
		return
	}

	if handler == nil {
		return
	}

	resp, err := cl.Recv()
	if err != nil {
		return
	}
	if len(resp) == 0 {
		return
	}

	err = handler(ctx, resp)

	return
}

//
// Recv message from server.
//
func (cl *Client) Recv() (packet []byte, err error) {
	err = cl.conn.SetReadDeadline(time.Now().Add(_defRWTO))
	if err != nil {
		return nil, err
	}

	bs := _bsPool.Get().(*[]byte)

	n, err := cl.conn.Read(*bs)
	if err != nil {
		_bsPool.Put(bs)
		return nil, err
	}
	if n == 0 {
		_bsPool.Put(bs)
		return nil, nil
	}

	bb := _bbPool.Get().(*bytes.Buffer)
	bb.Reset()

	for n == _maxBuffer {
		_, err = bb.Write((*bs)[:n])
		if err != nil {
			goto out
		}

		err = cl.conn.SetReadDeadline(time.Now().Add(_defRWTO))
		if err != nil {
			return nil, err
		}

		n, err = cl.conn.Read(*bs)
		if err != nil {
			goto out
		}

	}
	if n > 0 {
		_, err = bb.Write((*bs)[:n])
		if err != nil {
			goto out
		}
	}

out:
	if err == nil {
		packet = make([]byte, bb.Len())
		copy(packet, bb.Bytes())
	}

	_bsPool.Put(bs)
	_bbPool.Put(bb)

	return packet, err
}
