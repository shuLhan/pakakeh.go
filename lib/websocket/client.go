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
	"os"
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
	State           ConnState
	URL             *url.URL
	serverAddr      string
	handshakePath   string
	handshakeOrigin string
	handshakeExt    string
	handshakeProto  string
	handshakeHeader http.Header
	conn            net.Conn
	bb              bytes.Buffer
	IsTLS           bool
}

//
// NewClient create a new client connection to websocket server with a
// handshake.
//
// The endpoint use the following format,
//
//
//	3.  WebSocket URIs
//
//	   This specification defines two URI schemes, using the ABNF syntax
//	   defined in RFC 5234 [RFC5234], and terminology and ABNF productions
//	   defined by the URI specification RFC 3986 [RFC3986].
//
//	     ws-URI = "ws:" "//" host [ ":" port ] path [ "?" query ]
//	     wss-URI = "wss:" "//" host [ ":" port ] path [ "?" query ]
//
//	     ...
//
//	   The port component is OPTIONAL; the default for "ws" is port 80,
//	   while the default for "wss" is port 443.
//
//
func NewClient(endpoint string, headers http.Header) (cl *Client, err error) {
	cl = &Client{}

	cl.serverAddr, err = cl.ParseURI(endpoint)
	if err != nil {
		cl = nil
		return
	}

	cl.handshakeHeader = headers

	err = cl.Reconnect()
	if err != nil {
		cl = nil
	}

	return
}

//
// ParseURI of websocket connection scheme from "endpoint" and set client URL
// and TLS status to true if scheme is "wss://".
//
// On success it will set and return server address that can be used on
// Open().
//
// On fail it will return empty server address and error.
//
func (cl *Client) ParseURI(endpoint string) (serverAddr string, err error) {
	cl.URL, err = url.ParseRequestURI(endpoint)
	if err != nil {
		cl = nil
		return
	}

	if cl.URL.Scheme == _schemeWSS {
		cl.IsTLS = true
	}

	cl.serverAddr = GetConnectAddr(cl.URL)
	serverAddr = cl.serverAddr

	return
}

//
// GetConnectAddr return "host:port" from value in URL. By default, if no port
// is given, it will set to 80.
//
func GetConnectAddr(u *url.URL) (addr string) {
	serverPort := u.Port()

	if len(serverPort) == 0 {
		switch u.Scheme {
		case _schemeWS:
			serverPort = _defPort
		case _schemeWSS:
			serverPort = _defPortSecure
		default:
			serverPort = _defPort
		}

		addr = u.Hostname() + ":" + serverPort
	} else {
		addr = u.Host
	}

	return
}

//
// Open TCP connection to websocket server address in "host:port" format.
// If client "IsTLS" field is true, the connection is opened with TLS protocol
// and the remote name MUST have a valid certificate.
//
func (cl *Client) Open(addr string) (err error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	if cl.IsTLS {
		cfg := &tls.Config{
			InsecureSkipVerify: cl.IsTLS, //nolint:gas
		}

		cl.conn, err = tls.DialWithDialer(dialer, _netNameTCP, addr, cfg)
	} else {
		cl.conn, err = dialer.Dial(_netNameTCP, addr)
	}
	if err != nil {
		return
	}

	cl.State = ConnStateOpen

	return
}

//
// Handshake send the websocket opening handshake.
//
//	RFC6455 P17-P19
//	1.   The handshake MUST be a valid HTTP request as specified by
//	     [RFC2616].
//
//	2.   The method of the request MUST be GET, and the HTTP version MUST
//	     be at least 1.1.
//
//	     For example, if the WebSocket URI is "ws://example.com/chat",
//	     the first line sent should be "GET /chat HTTP/1.1".
//
//	3.   The "Request-URI" part of the request MUST match the /resource
//	     name/ defined in Section 3 (a relative URI) or be an absolute
//	     http/https URI that, when parsed, has a /resource name/, /host/,
//	     and /port/ that match the corresponding ws/wss URI.
//
//	4.   The request MUST contain a |Host| header field whose value
//	     contains /host/ plus optionally ":" followed by /port/ (when not
//	     using the default port).
//
//	5.   The request MUST contain an |Upgrade| header field whose value
//	     MUST include the "websocket" keyword.
//
//	6.   The request MUST contain a |Connection| header field whose value
//	     MUST include the "Upgrade" token.
//
//	7.   The request MUST include a header field with the name
//	     |Sec-WebSocket-Key|.  The value of this header field MUST be a
//	     nonce consisting of a randomly selected 16-byte value that has
//	     been base64-encoded (see Section 4 of [RFC4648]).  The nonce
//	     MUST be selected randomly for each connection.
//
//	     NOTE: As an example, if the randomly selected value was the
//	     sequence of bytes 0x01 0x02 0x03 0x04 0x05 0x06 0x07 0x08 0x09
//	     0x0a 0x0b 0x0c 0x0d 0x0e 0x0f 0x10, the value of the header
//	     field would be "AQIDBAUGBwgJCgsMDQ4PEC=="
//
//	8.   The request MUST include a header field with the name |Origin|
//	     [RFC6454] if the request is coming from a browser client.  If
//	     the connection is from a non-browser client, the request MAY
//	     include this header field if the semantics of that client match
//	     the use-case described here for browser clients.  The value of
//	     this header field is the ASCII serialization of origin of the
//	     context in which the code establishing the connection is
//	     running.  See [RFC6454] for the details of how this header field
//	     value is constructed.
//
//	     As an example, if code downloaded from www.example.com attempts
//	     to establish a connection to ww2.example.com, the value of the
//	     header field would be "http://www.example.com".
//
//	9.   The request MUST include a header field with the name
//	     |Sec-WebSocket-Version|.  The value of this header field MUST be
//	     13.
//
//	     NOTE: Although draft versions of this document (-09, -10, -11,
//	     and -12) were posted (they were mostly comprised of editorial
//	     changes and clarifications and not changes to the wire
//	     protocol), values 9, 10, 11, and 12 were not used as valid
//	     values for Sec-WebSocket-Version.  These values were reserved in
//	     the IANA registry but were not and will not be used.
//
//	10.  The request MAY include a header field with the name
//	     |Sec-WebSocket-Protocol|.  If present, this value indicates one
//	     or more comma-separated subprotocol the client wishes to speak,
//	     ordered by preference.  The elements that comprise this value
//	     MUST be non-empty strings with characters in the range U+0021 to
//	     U+007E not including separator characters as defined in
//	     [RFC2616] and MUST all be unique strings.  The ABNF for the
//	     value of this header field is 1#token, where the definitions of
//	     constructs and rules are as given in [RFC2616].
//
//	11.  The request MAY include a header field with the name
//	     |Sec-WebSocket-Extensions|.  If present, this value indicates
//	     the protocol-level extension(s) the client wishes to speak.  The
//	     interpretation and format of this header field is described in
//	     Section 9.1.
//
//	12.  The request MAY include any other header fields, for example,
//	     cookies [RFC6265] and/or authentication-related header fields
//	     such as the |Authorization| header field [RFC2616], which are
//	     processed according to documents that define them.
//
func (cl *Client) Handshake(path, origin, proto, ext string, headers http.Header) (err error) {
	if len(path) == 0 {
		path = cl.URL.EscapedPath() + "?" + cl.URL.RawQuery
	}

	cl.handshakePath = path

	cl.bb.Reset()
	key := GenerateHandshakeKey()
	keyAccept := GenerateHandshakeAccept(key)

	_, err = fmt.Fprintf(&cl.bb, _handshakeReqFormat, path, cl.URL.Host, key)
	if err != nil {
		return err
	}

	// (8)
	if len(origin) > 0 {
		cl.handshakeOrigin = origin
		_, _ = cl.bb.WriteString(_hdrKeyOrigin + ": " + origin + "\r\n")
		if err != nil {
			return err
		}
	}
	// (10)
	if len(proto) > 0 {
		cl.handshakeProto = proto
		_, err = cl.bb.WriteString(_hdrKeyWSProtocol + ": " + proto + "\r\n")
		if err != nil {
			return err
		}
	}
	// (11)
	if len(ext) > 0 {
		cl.handshakeExt = ext
		_, err = cl.bb.WriteString(_hdrKeyWSExtensions + ": " + ext + "\r\n")
		if err != nil {
			return err
		}
	}
	// (12)
	if len(headers) > 0 {
		cl.handshakeHeader = headers
		err = headers.Write(&cl.bb)
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
		fmt.Fprintln(os.Stderr, "handleHandshake:", err)
		cl.State = ConnStateError
		return
	}

	if httpRes.StatusCode != http.StatusSwitchingProtocols {
		err = fmt.Errorf("handleHandshake:" + httpRes.Status)
		cl.State = ConnStateError
		httpRes.Body.Close()
		return
	}

	expAccept := ctx.Value(ctxKeyWSAccept)
	gotAccept := httpRes.Header.Get(_hdrKeyWSAccept)
	if expAccept != gotAccept {
		err = fmt.Errorf("handleHandshake: invalid server accept key")
		cl.State = ConnStateError
		httpRes.Body.Close()
		return
	}

	cl.State = ConnStateConnected
	httpRes.Body.Close()

	return
}

//
// Reconnect to server using previous address and handshake parameters.
//
func (cl *Client) Reconnect() (err error) {
	if cl.conn != nil {
		_ = cl.conn.Close()
	}

	err = cl.Open(cl.serverAddr)
	if err != nil {
		return
	}

	err = cl.Handshake(cl.handshakePath, cl.handshakeOrigin,
		cl.handshakeProto, cl.handshakeExt, cl.handshakeHeader)

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
