// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

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

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
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

// ErrConnClosed define an error if client is not connected and try to
// send a message.
var ErrConnClosed = errors.New(`client is not connected`)

// Client for WebSocket protocol.
//
// Unlike HTTP client or other most commmon TCP oriented client, the WebSocket
// client is asynchronous or passive-active instead of synchronous.
// At any time client connection is open to server, client can receive a
// message broadcasted from server.
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
//		Endpoint: `ws://127.0.0.1:9001`,
//		HandleText: func(cl *Client, frame *Frame) error {
//			// Process response from request or broadcast from
//			// server.
//			return nil
//		}
//	}
//
//	err := cl.Connect()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = cl.SendText([]byte("Hello from client"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
// At any time, server may send PING or CLOSE the connection.
// For this messages, client already handled it by sending PONG message or by
// closing underlying connection automatically.
// Implementor can check a closed connection from error returned from Send
// methods to match with ErrConnClosed.
type Client struct {
	conn net.Conn

	// Headers The headers field can be used to pass custom headers during
	// handshake with server.
	// Any primary header fields ("host", "upgrade", "connection",
	// "sec-websocket-key", "sec-websocket-version") will be deleted
	// before handshake.
	Headers http.Header

	remoteURL *url.URL

	// TLSConfig define custom TLS configuration when connecting to secure
	// WebSocket server.
	// The scheme of Endpoint must be "https" or "wss", or it will be
	// resetting back to nil.
	TLSConfig *tls.Config

	frame  *Frame
	frames *Frames

	// HandleBin callback that will be called after receiving data
	// frame binary from server.
	HandleBin ClientHandler

	// handleClose function that will be called when client receive
	// control CLOSE frame from server.
	// Default handle is to response with control CLOSE frame with the
	// same payload.
	// This field is not exported, and only defined to allow testing.
	handleClose ClientHandler

	// handlePing function that will be called when client receive control
	// PING frame from server.
	// Default handler is to response with PONG.
	// This field is not exported, and only defined to allow testing.
	handlePing ClientHandler

	// handlePong a function that will be called when client receive
	// control PONG frame from server.
	// Default is nil.
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

	// Endpoint the address of remote server.
	// The endpoint use the following format,
	//
	//	ws-URI = "ws://" host [ ":" port ] path [ "?" query ]
	//	wss-URI = "wss://" host [ ":" port ] path [ "?" query ]
	//
	// The port component is OPTIONAL, default is 80 for "ws" scheme, and
	// 443 for "wss" scheme.
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

// Close gracefully close the client connection by sending control CLOSE frame
// with status normal to server and wait for response for as long as 10
// seconds.
func (cl *Client) Close() (err error) {
	var logp = `Close`

	cl.gracefulClose = make(chan bool, 1)
	defer func() {
		close(cl.gracefulClose)
		cl.gracefulClose = nil
	}()

	err = cl.sendClose(StatusNormal, nil)
	if errors.Is(err, ErrConnClosed) {
		return nil
	}

	cl.Lock()
	defer cl.Unlock()

	var (
		timer = time.NewTimer(defaultTimeout)
		wait  = true
	)

	// Wait for server to response with CLOSE or until 10 seconds without
	// reponse.
	for wait {
		select {
		case <-timer.C:
			// We did not receive server CLOSE frame in timely
			// manner.
			wait = false
		case <-cl.gracefulClose:
			timer.Stop()
			wait = false
		}
	}

	if cl.conn == nil {
		return nil
	}

	err = cl.conn.Close()
	if err != nil {
		err = fmt.Errorf(`%s: %w`, logp, err)
	}
	cl.conn = nil

	return err
}

// Connect to endpoint.
func (cl *Client) Connect() (err error) {
	var logp = `Connect`

	cl.Lock()

	if cl.conn != nil {
		_ = cl.conn.Close()
		cl.conn = nil
	}

	err = cl.init()
	if err != nil {
		cl.Unlock()
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = cl.open()
	if err != nil {
		cl.Unlock()
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var rest []byte

	rest, err = cl.handshake()
	if err != nil {
		_ = cl.conn.Close()
		cl.conn = nil
		cl.Unlock()
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	cl.Unlock()

	// At this point client successfully connected to server, but the
	// response from server may include WebSocket frame, not just HTTP
	// response.
	if len(rest) > 0 {
		var isClosing = cl.handleRaw(rest)
		if isClosing {
			cl.Quit()
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
func dummyHandle(_ *Client, _ *Frame) error {
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
		return fmt.Errorf(`init: %w`, err)
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
	var logp = `parseURI`

	cl.remoteURL, err = url.ParseRequestURI(cl.Endpoint)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		serverAddress = cl.remoteURL.Hostname()
		serverPort    = cl.remoteURL.Port()
	)

	switch cl.remoteURL.Scheme {
	case schemeWSS, schemeHTTPS:
		if len(serverPort) == 0 {
			serverPort = defTLSPort
		}
		if cl.TLSConfig == nil {
			cl.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
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
		logp   = `open`
		dialer = &net.Dialer{
			Timeout: 30 * time.Second,
		}
	)

	if cl.TLSConfig != nil {
		cl.conn, err = tls.DialWithDialer(dialer, "tcp", cl.remoteAddr, cl.TLSConfig)
	} else {
		cl.conn, err = dialer.Dial("tcp", cl.remoteAddr)
	}
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// handshake send the WebSocket opening handshake.
func (cl *Client) handshake() (rest []byte, err error) {
	var (
		logp      = `handshake`
		path      = cl.remoteURL.EscapedPath()
		key       = generateHandshakeKey()
		keyAccept = generateHandshakeAccept(key)

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
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	if len(cl.Headers) > 0 {
		err = cl.Headers.Write(&bb)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	bb.WriteString("\r\n")
	req = bb.Bytes()

	rest, err = cl.doHandshake(keyAccept, req)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
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

// clientOnClose request from server.
func clientOnClose(cl *Client, frame *Frame) (err error) {
	switch {
	case frame.closeCode == 0:
		frame.closeCode = StatusBadRequest

	case frame.closeCode < StatusNormal:
		frame.closeCode = StatusBadRequest

	case frame.closeCode == 1004:
		// Reserved.
		// The specific meaning might be defined in the future.
		frame.closeCode = StatusBadRequest

	case frame.closeCode == 1005:
		// 1005 is a reserved value and MUST NOT be set as a status
		// code in a Close control frame by an endpoint.
		// It is designated for use in applications expecting a status
		// code to indicate that no status code was actually present.
		frame.closeCode = StatusBadRequest

	case frame.closeCode == 1006:
		// 1006 is a reserved value and MUST NOT be set as a status
		// code in a Close control frame by an endpoint.
		// It is designated for use in applications expecting a status
		// code to indicate that the connection was closed abnormally,
		// e.g., without sending or receiving a Close control frame.
		frame.closeCode = StatusBadRequest

	case frame.closeCode >= 1015 && frame.closeCode <= 2999:
		frame.closeCode = StatusBadRequest

	case frame.closeCode >= 3000 && frame.closeCode <= 3999:
		// Status codes in the range 3000-3999 are reserved for use by
		// libraries, frameworks, and applications.
		// These status codes are registered directly with IANA.
		// The interpretation of these codes is undefined by this
		// protocol.
	case frame.closeCode >= 4000 && frame.closeCode <= 4999:
		// Status codes in the range 4000-4999 are reserved for
		// private use and thus can't be registered.
		// Such codes can be used by prior agreements between
		// WebSocket applications.
		// The interpretation of these codes is undefined by this
		// protocol.
	}
	if len(frame.payload) >= 2 {
		frame.payload = frame.payload[2:]
		if !utf8.Valid(frame.payload) {
			frame.closeCode = StatusBadRequest
		}
	}

	var packet = NewFrameClose(true, frame.closeCode, frame.payload)

	cl.Lock()
	err = cl.send(packet)
	cl.Unlock()

	cl.Quit()

	return err
}

// handleFragment will handle continuation frame (fragmentation).
func (cl *Client) handleFragment(frame *Frame) (isInvalid bool) {
	if cl.frames == nil {
		if frame.opcode == OpcodeCont {
			// If a connection does not have continuous frame,
			// then current frame opcode must not be 0.
			_ = cl.sendClose(StatusBadRequest, nil)
			return true
		}
	} else if frame.opcode != OpcodeCont {
		// If a connection have continuous frame, the next frame
		// opcode must be 0.
		_ = cl.sendClose(StatusBadRequest, nil)
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
			_ = cl.sendClose(StatusInvalidData, nil)
			return true
		}
		err = cl.HandleText(cl, frame)
	} else {
		err = cl.HandleBin(cl, frame)
	}
	if err != nil {
		_ = cl.sendClose(StatusBadRequest, nil)
		return true
	}

	return false
}

// handleFrame handle a single frame from client.
func (cl *Client) handleFrame(frame *Frame) (isClosing bool) {
	if !frame.isValid(false, cl.allowRsv1, cl.allowRsv2, cl.allowRsv3) {
		_ = cl.sendClose(StatusBadRequest, nil)
		return true
	}

	switch frame.opcode {
	case OpcodeCont, OpcodeText, OpcodeBin:
		var isInvalid = cl.handleFragment(frame)
		if isInvalid {
			isClosing = true
		}
	case OpcodeDataRsv3, OpcodeDataRsv4, OpcodeDataRsv5, OpcodeDataRsv6,
		OpcodeDataRsv7:
		_ = cl.sendClose(StatusBadRequest, nil)
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

func (cl *Client) handleHandshake(keyAccept string, resp []byte) (rest []byte, err error) {
	var httpRes *http.Response

	httpRes, rest, err = libhttp.ParseResponseHeader(resp) //nolint: bodyclose
	if err != nil {
		return nil, err
	}

	if httpRes.StatusCode != http.StatusSwitchingProtocols {
		return nil, errors.New(httpRes.Status)
	}

	var gotAccept = httpRes.Header.Get(_hdrKeyWSAccept)

	if keyAccept != gotAccept {
		return nil, errors.New(`invalid server accept key`)
	}

	return rest, nil
}

// handleRaw packet from server.
func (cl *Client) handleRaw(packet []byte) (isClosing bool) {
	var (
		logp   = `handleRaw`
		frames = Unpack(packet)

		f *Frame
	)

	if frames == nil {
		log.Printf(`%s: incomplete frames received`, logp)
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
func (cl *Client) SendBin(payload []byte) (err error) {
	var (
		logp   = `SendBin`
		packet = NewFrameBin(true, payload)
	)
	cl.Lock()
	err = cl.send(packet)
	cl.Unlock()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// sendClose send the control CLOSE frame to server with optional payload.
func (cl *Client) sendClose(status CloseCode, payload []byte) (err error) {
	var (
		logp   = `sendClose`
		packet = NewFrameClose(true, status, payload)
	)
	cl.Lock()
	err = cl.send(packet)
	cl.Unlock()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// SendPing send control PING frame to server, expecting PONG as response.
func (cl *Client) SendPing(payload []byte) (err error) {
	var (
		logp   = `SendPing`
		packet = NewFramePing(true, payload)
	)
	cl.Lock()
	err = cl.send(packet)
	cl.Unlock()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// SendPong send the control frame PONG to server, by using payload from PING
// frame.
func (cl *Client) SendPong(payload []byte) (err error) {
	var (
		logp   = `SendPong`
		packet = NewFramePong(true, payload)
	)
	cl.Lock()
	err = cl.send(packet)
	cl.Unlock()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// SendText send data frame as text to server.
// If handler is nil, no response will be read from server.
func (cl *Client) SendText(payload []byte) (err error) {
	var (
		logp   = `SendText`
		packet = NewFrameText(true, payload)
	)
	cl.Lock()
	err = cl.send(packet)
	cl.Unlock()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// serve read one data frame at a time from server and propagated to handler.
func (cl *Client) serve() {
	var logp = `serve`

	if cl.conn == nil {
		log.Printf(`%s: client is not connected`, logp)
		return
	}

	var (
		frame     *Frame
		packet    []byte
		err       error
		isClosing bool
	)

	for !isClosing {
		packet, err = cl.recv()
		if err != nil {
			log.Printf(`%s: %s`, logp, err)
			isClosing = true
			continue
		}
		if len(packet) == 0 {
			// Empty packet may indicated that server has closed
			// the connection abnormally.
			log.Printf(`%s: empty packet received, closing`, logp)
			isClosing = true
			continue
		}
		if cl.frame != nil {
			packet = cl.frame.unpack(packet)
			if cl.frame.isComplete {
				frame = cl.frame
				cl.frame = nil
				isClosing = cl.handleFrame(frame)
				if isClosing {
					continue
				}
			}
			if len(packet) == 0 {
				continue
			}
		}
		isClosing = cl.handleRaw(packet)
	}
	cl.Quit()
}

// Quit force close the client connection without sending control CLOSE frame.
// This function MUST be used only when error receiving packet from server
// (e.g. lost connection) to release the resource.
func (cl *Client) Quit() {
	var (
		logp = `Quit`
		err  error
	)

	cl.Lock()

	if cl.conn == nil {
		goto out
	}

	err = cl.conn.Close()
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
	}

	cl.conn = nil
	if cl.HandleQuit != nil {
		cl.HandleQuit()
	}

out:
	cl.Unlock()
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
	var logp = `recv`

	if cl.conn == nil {
		return nil, fmt.Errorf(`%s: %w`, logp, ErrConnClosed)
	}

	var (
		buf = make([]byte, 512)
		n   int
	)

	for {
		err = cl.conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		n, err = cl.conn.Read(buf)
		if err != nil {
			var neterr net.Error
			if errors.As(err, &neterr) && neterr.Timeout() {
				continue
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		if n == 0 {
			break
		}

		packet = append(packet, buf[:n]...)
		if n < len(buf) {
			break
		}
	}

	return packet, nil
}

func (cl *Client) send(packet []byte) (err error) {
	var logp = `send`

	if cl.conn == nil {
		return fmt.Errorf(`%s: %w`, logp, ErrConnClosed)
	}

	err = cl.conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	_, err = cl.conn.Write(packet)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// pinger send the PING control frame every 10 seconds.
func (cl *Client) pinger() {
	var (
		logp = `pinger`
		t    = time.NewTicker(cl.PingInterval)

		err error
	)

	for range t.C {
		err = cl.SendPing(nil)
		if err != nil {
			if errors.Is(err, ErrConnClosed) {
				return
			}
			log.Printf(`%s: %s`, logp, err)
		}
	}
}
