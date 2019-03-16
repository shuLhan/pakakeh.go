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
	"sync"
	"time"
	"unicode/utf8"

	"github.com/shuLhan/share/lib/debug"
)

const (
	_handshakeReqFormat = "GET %s HTTP/1.1\r\n" +
		"Host: %s\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Key: %s\r\n" +
		"Sec-Websocket-Version: 13\r\n"
)

var (
	ErrConnClosed = fmt.Errorf("websocket: client is not connected")
)

//
// Client for WebSocket protocol.
//
// Unlike HTTP client or other most commmon TCP oriented client, the WebSocket
// client is actually asynchronous or passive-active instead of synchronous.
// At any time client connection is open to server, client can receive a
// message broadcast from server.
//
// Case examples: if client send "A" to server, and expect that server
// response with "A+", server may send message "B" before sending "A+".
// Another case is when client connection is open, server may send "B" and "C"
// in any order without any request send by client previously.
//
// Due to this model, the way to handle response from server is centralized
// using handlers instead of using single send request-response.
//
// Client Example
//
// The following snippet show how to create a client and handling response
// from request or broadcast from server,
//
//	cl := &Client{
//		Endpoint: "ws://127.0.0.1:9001",
//		HandleText: func(cl *Client, frame *Frame) error {
//			// Process response from request or broadcast from
//			// server.
//			return nil
//		}
//	}
//
//	err := cl.Connect()
//	if err != nil {
//		log.Fatal("Connect: "+ err.Error())
//	}
//
//	err := cl.SendText([]byte("Hello from client"))
//	if err != nil {
//		log.Fatal("Connect: "+ err.Error())
//	}
//
// At any time, server may send PING or CLOSE the connection.  For this
// messages, client already handled it by sending PONG message or by closing
// underlying connection automatically.
// Implementor can check closed connection from error returned from Send
// methods to match with ErrConnClosed.
//
type Client struct {
	sync.Mutex
	conn net.Conn

	//
	// Endpoint contains URI of remote server.  The endpoint use the
	// following format,
	//
	//	ws-URI = "ws:" "//" host [ ":" port ] path [ "?" query ]
	//	wss-URI = "wss:" "//" host [ ":" port ] path [ "?" query ]
	//
	// The port component is OPTIONAL, default is 80 for "ws" scheme, and
	// 443 for "wss" scheme.
	//
	Endpoint string

	frame  *Frame
	frames *Frames

	// HandleBin callback that will be called after receiving data
	// frame binary from server.
	HandleBin ClientHandler

	// handleClose function that will be called when client receive
	// control CLOSE frame from server.  Default handle is to response
	// with control CLOSE frame with the same payload.
	// This field is not exported, and only defined to allow testing.
	handleClose ClientHandler

	// handlePing function that will be called when client receive control
	// PING frame from server.  Default handler is to response with PONG.
	// This field is not exported, and only defined to allow testing.
	handlePing ClientHandler

	// handlePong a function that will be called when client receive
	// control PONG frame from server.  Default is nil.
	handlePong ClientHandler

	// HandleQuit function that will be called when client connection is
	// closed.
	// Default is nil.
	HandleQuit func()

	// HandleRsvControl function that will be called when client received
	// reserved control frame (opcode 0xB-F) from server.
	// Default handler is nil.
	HandleRsvControl ClientHandler

	// HandleText callback that will be called after receiving data
	// frame text from server.
	HandleText ClientHandler

	// Headers The headers field can be used to pass custom headers during
	// handshake with server.  Any primary header fields ("host",
	// "upgrade", "connection", "sec-websocket-key",
	// "sec-websocket-version") will be deleted before handshake.
	Headers http.Header

	remoteURL  *url.URL
	remoteAddr string

	allowRsv1 bool
	allowRsv2 bool
	allowRsv3 bool
	isTLS     bool
}

//
// Connect to endpoint.
//
func (cl *Client) Connect() (err error) {
	err = cl.init()
	if err != nil {
		return fmt.Errorf("websocket: Connect: " + err.Error())
	}

	if cl.conn != nil {
		cl.Quit()
	}

	err = cl.open()
	if err != nil {
		return fmt.Errorf("websocket: Connect: " + err.Error())
	}

	err = cl.handshake()
	if err != nil {
		_ = cl.conn.Close()
		cl.conn = nil
		return fmt.Errorf("websocket: Connect: " + err.Error())
	}

	go cl.serve()

	return nil
}

//
// dummyHandle define dummy handle for HandleText and HandleBin.
//
func dummyHandle(cl *Client, frame *Frame) error {
	return nil
}

//
// init parse the endpoint URI and (re) initialize the client remote address
// and headers.
//
func (cl *Client) init() (err error) {
	if cl.HandleBin == nil {
		cl.HandleBin = dummyHandle
	}
	if cl.handleClose == nil {
		cl.handleClose = clientOnClose
	}
	if cl.handlePing == nil {
		cl.handlePing = clientOnPing
	}
	if cl.HandleText == nil {
		cl.HandleText = dummyHandle
	}

	err = cl.parseURI()
	if err != nil {
		return err
	}

	if len(cl.Headers) > 0 {
		cl.Headers.Del(_hdrKeyHost)
		cl.Headers.Del(_hdrKeyUpgrade)
		cl.Headers.Del(_hdrKeyOrigin)
		cl.Headers.Del(_hdrKeyWSKey)
		cl.Headers.Del(_hdrKeyWSVersion)
	}

	return nil
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
func (cl *Client) parseURI() (err error) {
	cl.remoteURL, err = url.ParseRequestURI(cl.Endpoint)
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
	var bb bytes.Buffer

	path := cl.remoteURL.EscapedPath()
	if len(path) == 0 {
		path = "/"
	}

	if len(cl.remoteURL.RawQuery) > 0 {
		path += "?" + cl.remoteURL.RawQuery
	}

	key := generateHandshakeKey()
	keyAccept := generateHandshakeAccept(key)

	_, err = fmt.Fprintf(&bb, _handshakeReqFormat, path, cl.remoteURL.Host, key)
	if err != nil {
		return err
	}

	if len(cl.Headers) > 0 {
		err = cl.Headers.Write(&bb)
		if err != nil {
			return err
		}
	}

	bb.WriteString("\r\n")

	ctx := context.WithValue(context.Background(), ctxKeyWSAccept, keyAccept)

	return cl.sendWithHandler(ctx, bb.Bytes(), cl.handleHandshake)
}

//
// handleBadRequest by sending Close frame with status.
//
func (cl *Client) handleBadRequest() {
	frameClose := NewFrameClose(true, StatusBadRequest, nil)

	err := cl.send(frameClose)
	if err != nil {
		log.Println("websocket: server.handleBadRequest: " + err.Error())
	}
}

//
// clientOnClose request from server.
//
func clientOnClose(cl *Client, frame *Frame) error {
	switch {
	case frame.closeCode == 0:
		frame.closeCode = StatusBadRequest
	case frame.closeCode < StatusNormal:
		frame.closeCode = StatusBadRequest
	case frame.closeCode == 1004:
		// Reserved.  The specific meaning might be defined in the future.
		frame.closeCode = StatusBadRequest
	case frame.closeCode == 1005:
		// 1005 is a reserved value and MUST NOT be set as a status
		// code in a Close control frame by an endpoint.  It is
		// designated for use in applications expecting a status code
		// to indicate that no status code was actually present.
		frame.closeCode = StatusBadRequest
	case frame.closeCode == 1006:
		// 1006 is a reserved value and MUST NOT be set as a status
		// code in a Close control frame by an endpoint.  It is
		// designated for use in applications expecting a status code
		// to indicate that the connection was closed abnormally,
		// e.g., without sending or receiving a Close control frame.
		frame.closeCode = StatusBadRequest
	case frame.closeCode >= 1015 && frame.closeCode <= 2999:
		frame.closeCode = StatusBadRequest
	case frame.closeCode >= 3000 && frame.closeCode <= 3999:
		// Status codes in the range 3000-3999 are reserved for use by
		// libraries, frameworks, and applications.  These status
		// codes are registered directly with IANA.  The
		// interpretation of these codes is undefined by this
		// protocol.
	case frame.closeCode >= 4000 && frame.closeCode <= 4999:
		// Status codes in the range 4000-4999 are reserved for
		// private use and thus can't be registered.  Such codes can
		// be used by prior agreements between WebSocket applications.
		// The interpretation of these codes is undefined by this
		// protocol.
	}
	if len(frame.payload) >= 2 {
		frame.payload = frame.payload[2:]
		if !utf8.Valid(frame.payload) {
			frame.closeCode = StatusBadRequest
		}
	}

	packet := NewFrameClose(true, frame.closeCode, frame.payload)

	if debug.Value >= 3 {
		log.Printf("websocket: Client.onClose: %+v\n", frame)
	}

	err := cl.send(packet)
	if err != nil {
		log.Println("websocket: Client.onClose: Send: " + err.Error())
	}

	cl.Quit()

	return nil
}

//
// handleFragment will handle continuation frame (fragmentation).
//
func (cl *Client) handleFragment(frame *Frame) (isInvalid bool) {
	if debug.Value >= 3 {
		log.Printf("websocket: Client.handleFragment: frame: {fin:%d opcode:%d masked:%d len:%d, payload.len:%d}\n",
			frame.fin, frame.opcode, frame.masked, frame.len,
			len(frame.payload))
	}

	if cl.frames == nil {
		if frame.opcode == OpcodeCont {
			// If a connection does not have continuous frame,
			// then current frame opcode must not be 0.
			cl.handleBadRequest()
			return true
		}
	} else if frame.opcode != OpcodeCont {
		// If a connection have continuous frame, the next frame
		// opcode must be 0.
		cl.handleBadRequest()
		return true
	}

	if frame.fin == 0 {
		if uint64(len(frame.payload)) < frame.len {
			// Continuous frame with unfinished payload.
			cl.frame = frame
		} else {
			if cl.frames == nil {
				cl.frames = new(Frames)
			}
			cl.frames.Append(frame)
			cl.frame = nil
		}
		return false
	}

	if cl.frame == nil && uint64(len(frame.payload)) < frame.len {
		// Final frame with unfinished payload.
		cl.frame = frame
		return false
	}

	if cl.frames != nil {
		frame = cl.frames.fin(frame)
	}

	cl.frame = nil
	cl.frames = nil

	var err error
	if frame.opcode == OpcodeText {
		if !utf8.Valid(frame.payload) {
			cl.handleInvalidData()
			return true
		}
		err = cl.HandleText(cl, frame)
	} else {
		err = cl.HandleBin(cl, frame)
	}
	if err != nil {
		cl.handleBadRequest()
		return true
	}

	return false
}

//
// handleFrame handle a single frame from client.
//
func (cl *Client) handleFrame(frame *Frame) (isClosing bool) {
	if !frame.isValid(false, cl.allowRsv1, cl.allowRsv2, cl.allowRsv3) {
		cl.handleBadRequest()
		return true
	}

	if debug.Value >= 3 {
		log.Printf("websocket: Client.handleFrame: %+v\n", frame)
	}

	switch frame.opcode {
	case OpcodeCont, OpcodeText, OpcodeBin:
		isInvalid := cl.handleFragment(frame)
		if isInvalid {
			isClosing = true
		}
	case OpcodeDataRsv3, OpcodeDataRsv4, OpcodeDataRsv5, OpcodeDataRsv6, OpcodeDataRsv7:
		cl.handleBadRequest()
		return true
	case OpcodeClose:
		cl.handleClose(cl, frame)
		return true
	case OpcodePing:
		_ = cl.handlePing(cl, frame)
	case OpcodePong:
		if cl.handlePong != nil {
			_ = cl.handlePong(cl, frame)
		}
	case OpcodeControlRsvB, OpcodeControlRsvC, OpcodeControlRsvD, OpcodeControlRsvE, OpcodeControlRsvF:
		if cl.HandleRsvControl != nil {
			_ = cl.HandleRsvControl(cl, frame)
		} else {
			cl.handleClose(cl, frame)
			isClosing = true
		}
	}

	return isClosing
}

func (cl *Client) handleHandshake(ctx context.Context, resp []byte) (err error) {
	if debug.Value >= 3 {
		max := 512
		if len(resp) < 512 {
			max = len(resp)
		}
		fmt.Printf("websocket: Client.handleHandshake:\n%s\n--\n", resp[:max])
	}
	httpBuf := bufio.NewReader(bytes.NewBuffer(resp))

	httpRes, err := http.ReadResponse(httpBuf, nil)
	if err != nil {
		fmt.Printf("websocket: Client.handleHandshake: http.ReadResponse")
		return err
	}

	httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusSwitchingProtocols {
		fmt.Printf("websocket: Client.handleHandshake: status code: %d\n", httpRes.StatusCode)
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
// handleInvalidData by sending Close frame with status 1007.
//
func (cl *Client) handleInvalidData() {
	frameClose := NewFrameClose(true, StatusInvalidData, nil)

	err := cl.send(frameClose)
	if err != nil {
		log.Println("websocket: Client.handleInvalidData: " + err.Error())
	}
}

//
// SendBin send data frame as binary to server.
// If handler is nil, no response will be read from server.
//
func (cl *Client) SendBin(payload []byte) error {
	packet := NewFrameBin(true, payload)
	return cl.send(packet)
}

//
// SendClose send the control CLOSE frame to server.
// If waitResponse is true, client will wait for CLOSE response from server
// before closing the connection.
//
func (cl *Client) SendClose(status CloseCode, payload []byte) (err error) {
	packet := NewFrameClose(true, status, payload)
	return cl.send(packet)
}

//
// SendPing send control PING frame to server, expecting PONG as response.
//
func (cl *Client) SendPing(payload []byte) error {
	packet := NewFramePing(true, payload)
	return cl.send(packet)
}

//
// SendPong send the control frame PONG to server, by using payload from PING
// frame.
//
func (cl *Client) SendPong(payload []byte) error {
	packet := NewFramePong(true, payload)
	return cl.send(packet)
}

//
// SendText send data frame as text to server.
// If handler is nil, no response will be read from server.
//
func (cl *Client) SendText(payload []byte) (err error) {
	packet := NewFrameText(true, payload)
	return cl.send(packet)
}

//
// serve read one data frame at a time from server and propagated to handler.
//
func (cl *Client) serve() {
	if cl.conn == nil {
		log.Println("websocket: Client.serve: client is not connected")
		return
	}

	for {
		packet, err := cl.recv()
		if err != nil {
			log.Println("websocket: Client.serve: " + err.Error())
			break
		}
		if len(packet) == 0 {
			// Empty packet may indicated that server has closed
			// the connection abnormally.
			log.Println("websocket: Client.serve: empty packet received, closing")
			break
		}

		if debug.Value >= 3 {
			log.Printf("websocket: Client.serve: packet: % x\n", packet)
		}

		if cl.frame != nil {
			packet = cl.frame.unpack(packet)
			if cl.frame.isComplete {
				frame := cl.frame
				cl.frame = nil
				isClosing := cl.handleFrame(frame)
				if isClosing {
					return
				}
			}
			if len(packet) == 0 {
				continue
			}
		}

		frames := Unpack(packet)
		if frames == nil {
			log.Println("websocket: client.serve: uncomplete frames received")
			continue
		}

		for _, f := range frames.v {
			if !f.isComplete {
				cl.frame = f
				continue
			}
			isClosing := cl.handleFrame(f)
			if isClosing {
				return
			}
		}
	}
	cl.Quit()
}

//
// Quit force close the client connection without sending CLOSE control frame.
// This function MUST be used only when error receiving packet from server
// (e.g. lost connection) to release the resource.
//
func (cl *Client) Quit() {
	cl.Lock()
	if cl.conn == nil {
		cl.Unlock()
		return
	}
	err := cl.conn.Close()
	if err != nil {
		log.Println("websocket: client.Close: " + err.Error())
	}
	cl.conn = nil
	cl.Unlock()
	if cl.HandleQuit != nil {
		cl.HandleQuit()
	}
}

//
// clientOnPing default handler when client receive control PING frame from
// server.
//
func clientOnPing(cl *Client, frame *Frame) error {
	if frame == nil {
		return nil
	}
	return cl.SendPong(frame.payload)
}

//
// recv read raw stream from server.
//
func (cl *Client) recv() (packet []byte, err error) {
	if cl.conn == nil {
		return nil, ErrConnClosed
	}

	buf := make([]byte, 512)

	for {
		err = cl.conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			break
		}

		n, err := cl.conn.Read(buf)
		if err != nil || n == 0 {
			break
		}
		packet = append(packet, buf[:n]...)
		if n < len(buf) {
			break
		}
	}

	return packet, err
}

func (cl *Client) send(packet []byte) (err error) {
	if cl.conn == nil {
		return ErrConnClosed
	}

	err = cl.conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		return err
	}

	if debug.Value >= 3 {
		log.Printf("websocket: Client.send: % x\n", packet)
	}

	_, err = cl.conn.Write(packet)
	if err != nil {
		return err
	}

	return nil
}

//
// sendWithHandler send message to server, read the response, and pass it to
// handler.
//
func (cl *Client) sendWithHandler(ctx context.Context, req []byte, handleRaw clientRawHandler) (err error) {
	if cl.conn == nil {
		return ErrConnClosed
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
		return handleRaw(ctx, resp)
	}

	return nil
}
