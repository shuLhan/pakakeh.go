// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/shuLhan/share/lib/debug"
	libhttp "github.com/shuLhan/share/lib/http"
)

const (
	_handshakeReqFormat = "GET %s HTTP/1.1\r\n" +
		"Host: %s\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Key: %s\r\n" +
		"Sec-Websocket-Version: 13\r\n"

	schemeWSS   = "wss"
	schemeHTTPS = "https"
	defTLSPort  = "443"
	defPort     = "80"
)

var (
	// ErrConnClosed define an error if client connection is not
	// connected.
	ErrConnClosed = fmt.Errorf("websocket: client is not connected")
)

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
// # Client Example
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
//		log.Fatal(err.Error())
//	}
//
//	err := cl.SendText([]byte("Hello from client"))
//	if err != nil {
//		log.Fatal(err.Error())
//	}
//
// At any time, server may send PING or CLOSE the connection.  For this
// messages, client already handled it by sending PONG message or by closing
// underlying connection automatically.
// Implementor can check closed connection from error returned from Send
// methods to match with ErrConnClosed.
type Client struct {
	conn net.Conn

	// Headers The headers field can be used to pass custom headers during
	// handshake with server.  Any primary header fields ("host",
	// "upgrade", "connection", "sec-websocket-key",
	// "sec-websocket-version") will be deleted before handshake.
	Headers http.Header

	remoteURL *url.URL

	//
	// TLSConfig define custom TLS configuration when connecting to secure
	// WebSocket server.
	// The scheme of Endpoint must be "https" or "wss", or it will be
	// resetting back to nil.
	//
	TLSConfig *tls.Config

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

	// gracefulClose is a channel to gracefully close connection by
	// client.
	gracefulClose chan bool

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

	remoteAddr string

	// The interval where PING control frame will be send to server.
	// The minimum and default value is 10 seconds.
	PingInterval time.Duration

	sync.Mutex

	allowRsv1 bool
	allowRsv2 bool
	allowRsv3 bool
}

// Close gracefully close the client connection.
func (cl *Client) Close() (err error) {
	cl.Lock()
	defer cl.Unlock()

	if cl.conn == nil {
		return
	}

	var (
		packet []byte = NewFrameClose(true, StatusNormal, nil)
		timer  *time.Timer
	)

	cl.gracefulClose = make(chan bool, 1)

	err = cl.send(packet)
	if err != nil {
		return fmt.Errorf("websocket: Close: %w", err)
	}

	// Wait for server to response with CLOSE.
	timer = time.NewTimer(defaultTimeout)
loop:
	for {
		select {
		case <-timer.C:
			// We did not receive server CLOSE frame in timely
			// manner.
			break loop
		case <-cl.gracefulClose:
			timer.Stop()
			break loop
		}
	}

	err = cl.conn.Close()
	if err != nil {
		err = fmt.Errorf("websocket: Close: %w", err)
	}
	cl.conn = nil

	return err
}

// Connect to endpoint.
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

	var rest []byte

	rest, err = cl.handshake()
	if err != nil {
		_ = cl.conn.Close()
		cl.conn = nil
		return fmt.Errorf("websocket: Connect: " + err.Error())
	}

	// At this point client successfully connected to server, but the
	// response from server may include WebSocket frame, not just HTTP
	// response.
	if len(rest) > 0 {
		var isClosing bool = cl.handleRaw(rest)
		if isClosing {
			return nil
		}
	}

	if cl.PingInterval < defaultPingInterval {
		cl.PingInterval = defaultPingInterval
	}

	go cl.pinger()
	go cl.serve()

	return nil
}

// dummyHandle define dummy handle for HandleText and HandleBin.
func dummyHandle(cl *Client, frame *Frame) error {
	return nil
}

// init parse the endpoint URI and (re) initialize the client remote address
// and headers.
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

// parseURI parse WebSocket connection URI from Endpoint and get the remote
// URL (for checking up scheme) and remote address.
// By default, if no port is given, it will be set to 80 or 443 for "wss" or
// "https" scheme.
//
// On success it will set the remote address that can be used on open().
// On fail it will return an error.
func (cl *Client) parseURI() (err error) {
	cl.remoteURL, err = url.ParseRequestURI(cl.Endpoint)
	if err != nil {
		cl = nil
		return err
	}

	var (
		serverAddress string = cl.remoteURL.Hostname()
		serverPort    string = cl.remoteURL.Port()
	)

	switch cl.remoteURL.Scheme {
	case schemeWSS, schemeHTTPS:
		if len(serverPort) == 0 {
			serverPort = defTLSPort
		}
		if cl.TLSConfig == nil {
			cl.TLSConfig = &tls.Config{}
		}
	default:
		if len(serverPort) == 0 {
			serverPort = defPort
		}
		// Remove TLSConfig if scheme is not https or wss to prevent
		// error "Connect: tls: first record does not look like a TLS
		// handshake".
		cl.TLSConfig = nil
	}

	cl.remoteAddr = serverAddress + ":" + serverPort

	return nil
}

// open TCP connection to WebSocket remote address.
// If client "TLSConfig" field is not nil, the connection is opened with TLS
// protocol and the remote name MUST have a valid certificate.
func (cl *Client) open() (err error) {
	var (
		dialer = &net.Dialer{
			Timeout: 30 * time.Second,
		}
	)

	if debug.Value >= 3 {
		fmt.Printf("websocket: Client.open: remoteAddr: %s\n",
			cl.remoteAddr)
	}

	if cl.TLSConfig != nil {
		cl.conn, err = tls.DialWithDialer(dialer, "tcp",
			cl.remoteAddr, cl.TLSConfig)
	} else {
		cl.conn, err = dialer.Dial("tcp", cl.remoteAddr)
	}
	if err != nil {
		return err
	}

	return nil
}

// handshake send the WebSocket opening handshake.
func (cl *Client) handshake() (rest []byte, err error) {
	var (
		path      string = cl.remoteURL.EscapedPath()
		key       []byte = generateHandshakeKey()
		keyAccept string = generateHandshakeAccept(key)

		bb  bytes.Buffer
		req []byte
	)
	if len(path) == 0 {
		path = "/"
	}

	if len(cl.remoteURL.RawQuery) > 0 {
		path += "?" + cl.remoteURL.RawQuery
	}

	_, err = fmt.Fprintf(&bb, _handshakeReqFormat, path, cl.remoteURL.Host, key)
	if err != nil {
		return nil, err
	}

	if len(cl.Headers) > 0 {
		err = cl.Headers.Write(&bb)
		if err != nil {
			return nil, err
		}
	}

	bb.WriteString("\r\n")
	req = bb.Bytes()

	if debug.Value >= 3 {
		fmt.Printf("websocket: Client.handshake:\n%s\n--\n", req)
	}

	rest, err = cl.doHandshake(keyAccept, req)
	if err != nil {
		return nil, err
	}

	return rest, nil
}

func (cl *Client) doHandshake(keyAccept string, req []byte) (res []byte, err error) {
	err = cl.send(req)
	if err != nil {
		return nil, err
	}

	res, err = cl.recv()
	if err != nil {
		return nil, err
	}

	res, err = cl.handleHandshake(keyAccept, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// handleBadRequest by sending Close frame with status.
func (cl *Client) handleBadRequest() {
	var (
		frameClose []byte = NewFrameClose(true, StatusBadRequest, nil)
		err        error
	)

	err = cl.send(frameClose)
	if err != nil {
		log.Println("websocket: Client.handleBadRequest: " + err.Error())
	}
}

// clientOnClose request from server.
func clientOnClose(cl *Client, frame *Frame) (err error) {
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

	var packet []byte = NewFrameClose(true, frame.closeCode, frame.payload)

	if debug.Value >= 3 {
		fmt.Printf("websocket: clientOnClose: payload: %s\n", frame.payload)
	}

	err = cl.send(packet)
	if err != nil {
		log.Println("websocket: clientOnClose: send: " + err.Error())
	}

	cl.Quit()

	return nil
}

// handleFragment will handle continuation frame (fragmentation).
func (cl *Client) handleFragment(frame *Frame) (isInvalid bool) {
	if debug.Value >= 3 {
		fmt.Printf("websocket: Client.handleFragment: frame:{fin:%d opcode:%d masked:%d len:%d, payload.len:%d}\n",
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

// handleFrame handle a single frame from client.
func (cl *Client) handleFrame(frame *Frame) (isClosing bool) {
	if !frame.isValid(false, cl.allowRsv1, cl.allowRsv2, cl.allowRsv3) {
		cl.handleBadRequest()
		return true
	}

	switch frame.opcode {
	case OpcodeCont, OpcodeText, OpcodeBin:
		var isInvalid bool = cl.handleFragment(frame)
		if isInvalid {
			isClosing = true
		}
	case OpcodeDataRsv3, OpcodeDataRsv4, OpcodeDataRsv5, OpcodeDataRsv6,
		OpcodeDataRsv7:
		cl.handleBadRequest()
		return true
	case OpcodeClose:
		// Check if we are requesting the close.
		if cl.gracefulClose != nil {
			cl.gracefulClose <- true
		} else {
			_ = cl.handleClose(cl, frame)
		}
		return true
	case OpcodePing:
		_ = cl.handlePing(cl, frame)
	case OpcodePong:
		if cl.handlePong != nil {
			_ = cl.handlePong(cl, frame)
		}
	case OpcodeControlRsvB, OpcodeControlRsvC, OpcodeControlRsvD,
		OpcodeControlRsvE, OpcodeControlRsvF:
		if cl.HandleRsvControl != nil {
			_ = cl.HandleRsvControl(cl, frame)
		} else {
			_ = cl.handleClose(cl, frame)
			isClosing = true
		}
	}

	return isClosing
}

func (cl *Client) handleHandshake(keyAccept string, resp []byte) (
	rest []byte, err error,
) {
	if debug.Value >= 3 {
		var max int = 512
		if len(resp) < 512 {
			max = len(resp)
		}
		fmt.Printf("websocket: Client.handleHandshake:\n%s\n--\n",
			resp[:max])
	}

	var httpRes *http.Response

	httpRes, rest, err = libhttp.ParseResponseHeader(resp)
	if err != nil {
		return nil, err
	}

	if httpRes.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf(httpRes.Status)
	}

	var gotAccept string = httpRes.Header.Get(_hdrKeyWSAccept)
	if keyAccept != gotAccept {
		return nil, fmt.Errorf("invalid server accept key")
	}

	return rest, nil
}

// handleInvalidData by sending Close frame with status 1007.
func (cl *Client) handleInvalidData() {
	var (
		frameClose []byte = NewFrameClose(true, StatusInvalidData, nil)
		err        error
	)

	err = cl.send(frameClose)
	if err != nil {
		log.Println("websocket: Client.handleInvalidData: " + err.Error())
	}
}

// handleRaw packet from server.
func (cl *Client) handleRaw(packet []byte) (isClosing bool) {
	var (
		frames *Frames = Unpack(packet)
		f      *Frame
	)

	if frames == nil {
		log.Println("websocket: Client.handleRaw: incomplete frames received")
		return false
	}

	for _, f = range frames.v {
		if !f.isComplete {
			cl.frame = f
			continue
		}
		isClosing = cl.handleFrame(f)
		if isClosing {
			return true
		}
	}

	return false
}

// SendBin send data frame as binary to server.
// If handler is nil, no response will be read from server.
func (cl *Client) SendBin(payload []byte) error {
	var packet []byte = NewFrameBin(true, payload)
	return cl.send(packet)
}

// sendClose send the control CLOSE frame to server.
func (cl *Client) sendClose(status CloseCode, payload []byte) (err error) {
	var packet []byte = NewFrameClose(true, status, payload)
	return cl.send(packet)
}

// SendPing send control PING frame to server, expecting PONG as response.
func (cl *Client) SendPing(payload []byte) error {
	var packet []byte = NewFramePing(true, payload)
	return cl.send(packet)
}

// SendPong send the control frame PONG to server, by using payload from PING
// frame.
func (cl *Client) SendPong(payload []byte) error {
	var packet []byte = NewFramePong(true, payload)
	return cl.send(packet)
}

// SendText send data frame as text to server.
// If handler is nil, no response will be read from server.
func (cl *Client) SendText(payload []byte) (err error) {
	var packet []byte = NewFrameText(true, payload)
	return cl.send(packet)
}

// serve read one data frame at a time from server and propagated to handler.
func (cl *Client) serve() {
	if cl.conn == nil {
		log.Println("websocket: Client.serve: client is not connected")
		return
	}

	var (
		frame     *Frame
		packet    []byte
		err       error
		isClosing bool
	)

	for {
		packet, err = cl.recv()
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

		if cl.frame != nil {
			packet = cl.frame.unpack(packet)
			if cl.frame.isComplete {
				frame = cl.frame
				cl.frame = nil
				isClosing = cl.handleFrame(frame)
				if isClosing {
					return
				}
			}
			if len(packet) == 0 {
				continue
			}
		}

		isClosing = cl.handleRaw(packet)
		if isClosing {
			return
		}
	}
	cl.Quit()
}

// Quit force close the client connection without sending CLOSE control frame.
// This function MUST be used only when error receiving packet from server
// (e.g. lost connection) to release the resource.
func (cl *Client) Quit() {
	cl.Lock()
	defer cl.Unlock()

	if cl.conn == nil {
		return
	}

	var err error = cl.conn.Close()
	if err != nil {
		log.Println("websocket: client.Close: " + err.Error())
	}

	cl.conn = nil
	if cl.HandleQuit != nil {
		cl.HandleQuit()
	}
}

// clientOnPing default handler when client receive control PING frame from
// server.
func clientOnPing(cl *Client, frame *Frame) error {
	if frame == nil {
		return nil
	}
	return cl.SendPong(frame.payload)
}

// recv read raw stream from server.
func (cl *Client) recv() (packet []byte, err error) {
	if cl.conn == nil {
		return nil, ErrConnClosed
	}

	var (
		buf    []byte = make([]byte, 512)
		neterr net.Error
		max    int
		n      int
		ok     bool
	)

	for {
		err = cl.conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			break
		}

		n, err = cl.conn.Read(buf)
		if err != nil {
			neterr, ok = err.(net.Error)
			if ok && neterr.Timeout() {
				continue
			}
			break
		}
		if n == 0 {
			break
		}

		packet = append(packet, buf[:n]...)
		if n < len(buf) {
			break
		}
	}

	if debug.Value >= 3 {
		max = len(packet)
		if max > 16 {
			max = 16
		}
		fmt.Printf("websocket: Client.recv: packet: len:%d % x\n", len(packet), packet[:max])
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
		var max int = len(packet)
		if max > 16 {
			max = 16
		}
		fmt.Printf("websocket: Client.send: % x\n", packet[:max])
	}

	_, err = cl.conn.Write(packet)
	if err != nil {
		return err
	}

	return nil
}

// pinger send the PING control frame every 10 seconds.
func (cl *Client) pinger() {
	var (
		t   *time.Ticker = time.NewTicker(cl.PingInterval)
		err error
	)

	for range t.C {
		err = cl.SendPing(nil)
		if err != nil {
			if errors.Is(err, ErrConnClosed) {
				return
			}
			log.Println("websocket: pinger: " + err.Error())
		}
	}
}
