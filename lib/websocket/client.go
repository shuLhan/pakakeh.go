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
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	_handshakeReqFormat = "GET %s HTTP/1.1\r\n" +
		"Host: %s\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Key: %s\r\n" +
		"Sec-Websocket-Version: 13\r\n"
)

//
// Client for websocket.
//
type Client struct {
	remoteURL       *url.URL
	remoteAddr      string
	handshakeHeader http.Header
	conn            net.Conn
	bb              bytes.Buffer
	pingQueue       chan *Frame
	isTLS           bool

	handlePing clientRawHandler
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
	cl = &Client{
		handlePing: handlePing,
	}

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
// servePing serve control frame PING from server, by replying with PONG.
//
// If client error when sending PONG, the connection will be force closed.
//
func (cl *Client) servePing() {
	for f := range cl.pingQueue {
		err := cl.SendPong(f.payload)
		if err != nil {
			log.Println("websocket: client.servePing: " + err.Error())
			cl.Quit()
			return
		}
	}
}

//
// parseURI parse websocket connection URI from "endpoint" and get the remote
// URL (for checking up scheme) and remote address.
// By default, if no port is given, it will set to 80 for URL with any scheme
// or 443 for "wss" scheme.
//
// On success it will set the remote address that can be used on open().
// On fail it will return an error.
//
func (cl *Client) parseURI(endpoint string) (err error) {
	cl.remoteURL, err = url.ParseRequestURI(endpoint)
	if err != nil {
		cl = nil
		return err
	}

	serverPort := cl.remoteURL.Port()

	if len(serverPort) != 0 {
		cl.remoteAddr = cl.remoteURL.Host
		return nil
	}

	switch cl.remoteURL.Scheme {
	case "wss":
		serverPort = "443"
		cl.isTLS = true
	default:
		serverPort = "80"
	}

	cl.remoteAddr = cl.remoteURL.Hostname() + ":" + serverPort

	return nil
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

		cl.conn, err = tls.DialWithDialer(dialer, "tcp", cl.remoteAddr, cfg)
	} else {
		cl.conn, err = dialer.Dial("tcp", cl.remoteAddr)
	}
	if err != nil {
		return err
	}

	return nil
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

	return cl.send(ctx, cl.bb.Bytes(), cl.handleHandshake)
}

func (cl *Client) handleHandshake(ctx context.Context, resp []byte) (err error) {
	httpBuf := bufio.NewReader(bytes.NewBuffer(resp))

	httpRes, err := http.ReadResponse(httpBuf, nil)
	httpRes.Body.Close()
	if err != nil {
		return err
	}

	if httpRes.StatusCode != http.StatusSwitchingProtocols {
		err = fmt.Errorf(httpRes.Status)
		return err
	}

	expAccept := ctx.Value(ctxKeyWSAccept)
	gotAccept := httpRes.Header.Get(_hdrKeyWSAccept)
	if expAccept != gotAccept {
		err = fmt.Errorf("websocket: client.handleHandshake: invalid server accept key")
		return err
	}

	return nil
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
		return err
	}

	err = cl.handshake()
	if err != nil {
		_ = cl.conn.Close()
		return err
	}

	cl.pingQueue = make(chan *Frame, 24)
	go cl.servePing()

	return nil
}

//
// SendBin send data frame as binary to server.
// If handler is nil, no response will be read from server.
//
func (cl *Client) SendBin(ctx context.Context, bin []byte, handler ClientRecvHandler) error {
	return cl.sendData(ctx, bin, opcodeBin, handler)
}

//
// SendClose send the control CLOSE frame to server.
// If waitResponse is true, client will wait for CLOSE response from server
// before closing the connection.
//
func (cl *Client) SendClose(waitResponse bool) (err error) {
	packet := NewFrameClose(true, 0, nil)
	if waitResponse {
		err = cl.send(context.Background(), packet, cl.handleClose)
	} else {
		err = cl.send(context.Background(), packet, nil)
	}

	errClose := cl.conn.Close()
	if errClose != nil {
		log.Println("websocket: Client.SendClose: " + err.Error())
	}
	cl.conn = nil
	close(cl.pingQueue)

	return err
}

//
// SendPing send control PING frame to server, expecting PONG as response.
//
func (cl *Client) SendPing(ctx context.Context, payload []byte) error {
	packet := NewFramePing(true, payload)
	return cl.send(ctx, packet, cl.handlePing)
}

//
// SendPong send the control frame PONG to server, by using payload from PING
// frame.
//
func (cl *Client) SendPong(payload []byte) error {
	packet := NewFramePong(true, payload)
	return cl.send(context.Background(), packet, nil)
}

//
// SendText send data frame as text to server.
// If handler is nil, no response will be read from server.
//
func (cl *Client) SendText(ctx context.Context, text []byte, handler ClientRecvHandler) (err error) {
	return cl.sendData(ctx, text, opcodeText, handler)
}

//
// Recv read message as frames from server.
// One should not use this method manually, instead of the handler in Send()
// method.
//
func (cl *Client) Recv() (frames *Frames, err error) {
	if cl.conn == nil {
		return nil, fmt.Errorf("websocket: client.Send: client is not connected")
	}

	cl.bb.Reset()
	frames = &Frames{}
	bs := _bsPool.Get().(*[]byte)

	// Read all packet until we received frame with fin or operation code
	// CLOSE.
	for {
		err = cl.conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			goto out
		}

		n, err := cl.conn.Read(*bs)
		if err != nil {
			goto out
		}
		_, err = cl.bb.Write((*bs)[:n])
		if err != nil {
			goto out
		}
		if n == _maxBuffer {
			continue
		}

		f, _ := frameUnpack(cl.bb.Bytes())
		if f == nil {
			goto out
		}
		switch f.opcode {
		case opcodePing:
			cl.pingQueue <- f
		case opcodePong:
			// Ignore control PONG frame.
		case opcodeClose:
			frames.Append(f)
			goto out
		default:
			frames.Append(f)
			if f.fin == frameIsFinished {
				goto out
			}
		}
	}

out:
	_bsPool.Put(bs)

	return frames, err
}

//
// Quit force close the client connection without sending CLOSE control frame.
// This function MUST be used only when error receiving packet from server
// (e.g. lost connection) to release the resource.
//
func (cl *Client) Quit() {
	err := cl.conn.Close()
	if err != nil {
		log.Println("websocket: client.Close: " + err.Error())
	}
	cl.conn = nil
	close(cl.pingQueue)
}

//
// handleClose define a callback for SendClose() that expect server to
// response with CLOSE frame.
//
func (cl *Client) handleClose(ctx context.Context, packet []byte) error {
	f, _ := frameUnpack(packet)
	if f == nil {
		return fmt.Errorf("websocket: Client.handleClose: empty response")
	}
	if f.opcode != opcodeClose {
		return fmt.Errorf("websocket: Client.handleClose: expecting CLOSE frame, got %d",
			f.opcode)
	}
	return nil
}

//
// handlePing define a callback for SendPing() that expect server to response
// with PONG frame.
//
func handlePing(ctx context.Context, packet []byte) error {
	f, _ := frameUnpack(packet)
	if f == nil {
		return fmt.Errorf("websocket: Client.handlePing: empty response")
	}
	if f.opcode != opcodePong {
		return fmt.Errorf("websocket: Client.handleClose: expecting PONG frame, got %d",
			f.opcode)
	}
	return nil
}

//
// recv read raw stream from server.
//
func (cl *Client) recv() (packet []byte, err error) {
	if cl.conn == nil {
		return nil, fmt.Errorf("websocket: client.SendBin: client is not connected")
	}

	cl.bb.Reset()
	bs := _bsPool.Get().(*[]byte)

	for {
		err = cl.conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			break
		}

		n, err := cl.conn.Read(*bs)
		if err != nil {
			break
		}

		_, err = cl.bb.Write((*bs)[:n])
		if err != nil {
			break
		}

		if n != _maxBuffer {
			break
		}
	}

	packet = cl.bb.Bytes()

	_bsPool.Put(bs)

	return packet, err
}

//
// send raw stream to server, read the response, and pass it to handler.
//
func (cl *Client) send(ctx context.Context, req []byte, handleRaw clientRawHandler) (err error) {
	if cl.conn == nil {
		return fmt.Errorf("websocket: client.SendBin: client is not connected")
	}

	err = cl.conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		return err
	}

	_, err = cl.conn.Write(req)
	if err != nil {
		return err
	}

	if handleRaw != nil {
		var resp []byte
		resp, err = cl.recv()
		if err != nil {
			return err
		}
		err = handleRaw(ctx, resp)
	}

	return err
}

func (cl *Client) sendData(ctx context.Context, req []byte, opcode opcode, handler ClientRecvHandler) (err error) {
	if cl.conn == nil {
		return fmt.Errorf("websocket: client.SendBin: client is not connected")
	}

	var packet []byte

	if opcode == opcodeText {
		packet = NewFrameText(true, req)
	} else {
		packet = NewFrameBin(true, req)
	}

	err = cl.conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		return err
	}

	_, err = cl.conn.Write(packet)
	if err != nil {
		return err
	}

	if handler != nil {
		var frames *Frames
		frames, err = cl.Recv()
		if err != nil {
			return err
		}

		err = handler(ctx, frames)
	}

	return err
}
