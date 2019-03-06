// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/sys/unix"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

const (
	_maxQueue = 128

	_resUpgradeOK = "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Accept: %s\r\n\r\n"
)

//
// Server for websocket.
//
type Server struct {
	Clients *ClientManager

	sock      int
	chUpgrade chan int
	epollRead int
	routes    *rootRoute

	// HandleAuth callback that will be called when receiving
	// client handshake.
	HandleAuth HandlerAuthFn

	// HandleClientAdd callback that will called after client handshake
	// and, if HandleAuth is defined, after client is authenticated.
	HandleClientAdd HandlerClientFn

	// HandleClientRemove callback that will be called before client
	// connection being removed and closed by server.
	HandleClientRemove HandlerClientFn

	// HandleText callback that will be called after receiving data
	// frame(s) text from client.
	// Default handle parse the payload into Request and pass it to
	// registered routes.
	HandleText HandlerPayloadFn

	// HandleBin callback that will be called after receiving data
	// frame(s) binary from client.
	HandleBin HandlerPayloadFn
}

//
// NewServer will create new web-socket server that listen on port number.
//
func NewServer(port int) (serv *Server, err error) {
	serv = &Server{
		chUpgrade: make(chan int, _maxQueue),
		Clients:   newClientManager(),
		routes:    newRootRoute(),
	}

	err = serv.createEpoolRead()
	if err != nil {
		return
	}

	err = serv.createSockServer(port)
	if err != nil {
		return
	}

	serv.HandleText = serv.handleText
	serv.HandleBin = serv.handleBin
	serv.HandleClientAdd = nil
	serv.HandleClientRemove = nil

	return
}

func (serv *Server) createEpoolRead() (err error) {
	serv.epollRead, err = unix.EpollCreate1(0)
	if err != nil {
		return
	}

	return
}

func (serv *Server) createSockServer(port int) (err error) {
	serv.sock, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return
	}

	err = unix.SetsockoptInt(serv.sock, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		return
	}

	addr := unix.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())

	err = unix.Bind(serv.sock, &addr)
	if err != nil {
		return
	}

	err = unix.Listen(serv.sock, _maxQueue)

	return
}

//
// RegisterTextHandler register specific function to be called by server when
// request opcode is text, and method and target matched with Request.
//
func (serv *Server) RegisterTextHandler(method, target string, handler RouteHandler) (err error) {
	if len(method) == 0 || len(target) == 0 || handler == nil {
		return
	}

	err = serv.routes.add(method, target, handler)

	return
}

func (serv *Server) handleError(conn int, code int, msg string) {
	rspBody := "HTTP/1.1 " + strconv.Itoa(code) + " " + msg + "\r\n\r\n"

	_, err := unix.Write(conn, []byte(rspBody))
	if err != nil {
		log.Println("websocket: server.handleError: " + err.Error())
	}

	unix.Close(conn)
}

//
// handleUpgrade parse and validate websocket HTTP handshake from client.
// If HandleAuth is not nil, the HTTP handshake will be passed to that
// function to allow custom authentication.
//
// On success it will return the context from authentication and the WebSocket
// key.
//
func (serv *Server) handleUpgrade(httpRequest []byte) (
	ctx context.Context, key []byte, err error,
) {
	handshake := _handshakePool.Get().(*Handshake)

	err = handshake.parse(httpRequest)
	if err == nil {
		key = libbytes.Copy(handshake.Key)
		if serv.HandleAuth != nil {
			ctx, err = serv.HandleAuth(handshake)
		}
	}

	handshake.reset(nil)
	_handshakePool.Put(handshake)

	return ctx, key, err
}

//
// clientAdd add the new client connection to epoll and to list of clients.
//
func (serv *Server) clientAdd(ctx context.Context, conn int) (err error) {
	event := unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLONESHOT,
		Fd:     int32(conn),
	}

	err = unix.SetNonblock(conn, true)
	if err != nil {
		return
	}

	err = unix.EpollCtl(serv.epollRead, unix.EPOLL_CTL_ADD, conn, &event)
	if err != nil {
		return
	}

	if ctx != nil {
		serv.Clients.add(ctx, conn)

		if serv.HandleClientAdd != nil {
			go serv.HandleClientAdd(ctx, conn)
		}
	}

	return nil
}

//
// ClientRemove remove client connection from server.
//
func (serv *Server) ClientRemove(conn int) {
	ctx := serv.Clients.Context(conn)

	if ctx != nil && serv.HandleClientRemove != nil {
		serv.HandleClientRemove(ctx, conn)
	}

	serv.Clients.remove(conn)

	err := unix.EpollCtl(serv.epollRead, unix.EPOLL_CTL_DEL, conn, nil)
	if err != nil {
		log.Println("websocket: server.ClientRemove: " + err.Error())
	}

	err = unix.Close(conn)
	if err != nil {
		log.Println("websocket: server.ClientRemove: " + err.Error())
	}
}

func (serv *Server) upgrader() {
	for conn := range serv.chUpgrade {
		packet, err := Recv(conn)
		if err != nil {
			log.Println("websocket: server.upgrader: " + err.Error())
			unix.Close(conn)
			continue
		}
		if len(packet) == 0 {
			unix.Close(conn)
			continue
		}

		ctx, key, err := serv.handleUpgrade(packet)
		if err != nil {
			serv.handleError(conn, http.StatusBadRequest, err.Error())
			continue
		}

		wsAccept := generateHandshakeAccept(key)

		bb := _bbPool.Get().(*bytes.Buffer)
		bb.Reset()

		fmt.Fprintf(bb, _resUpgradeOK, wsAccept)

		_, err = unix.Write(conn, bb.Bytes())
		_bbPool.Put(bb)

		if err != nil {
			log.Println("websocket: server.upgrader: " + err.Error())
			unix.Close(conn)
			continue
		}

		err = serv.clientAdd(ctx, conn)
		if err != nil {
			log.Println("websocket: server.upgrader: " + err.Error())
			unix.Close(conn)
		}
	}
}

//
// handleFragment will handle continuation frame (fragmentation).
//
// (RFC 6455 Section 5.4 Page 34)
// A fragmented message consists of a single frame with the FIN bit
// clear and an opcode other than 0, followed by zero or more frames
// with the FIN bit clear and the opcode set to 0, and terminated by
// a single frame with the FIN bit set and an opcode of 0.  A
// fragmented message is conceptually equivalent to a single larger
// message whose payload is equal to the concatenation of the
// payloads of the fragments in order; however, in the presence of
// extensions, this may not hold true as the extension defines the
// interpretation of the "Extension data" present.  For instance,
// "Extension data" may only be present at the beginning of the first
// fragment and apply to subsequent fragments, or there may be
// "Extension data" present in each of the fragments that applies
// only to that particular fragment.  In the absence of "Extension
// data", the following example demonstrates how fragmentation works.
//
// EXAMPLE: For a text message sent as three fragments, the first
// fragment would have an opcode of 0x1 and a FIN bit clear, the
// second fragment would have an opcode of 0x0 and a FIN bit clear,
// and the third fragment would have an opcode of 0x0 and a FIN bit
// that is set.
// (RFC 6455 Section 5.4 Page 34)
//
func (serv *Server) handleFragment(conn int, req *Frame) {
	frames := serv.Clients.Frames(conn)

	if req.fin == 0 {
		if frames == nil {
			frames = &Frames{}
		}
		if req.opcode == opcodeBin || req.opcode == opcodeText {
			// Non-zero opcode indicate first fragment, so we
			// clear any previous fragmentations.
			frames.v = frames.v[:0]
		}

		frames.Append(req)

		serv.Clients.SetFrames(conn, frames)
	} else {
		var (
			payload []byte
			oc      opcode
		)

		if frames == nil {
			payload = req.payload
			oc = req.opcode
		} else {
			frames.Append(req)

			payload = frames.Payload()
			oc = frames.v[0].opcode
		}

		serv.Clients.SetFrames(conn, nil)

		if oc == opcodeText {
			go serv.HandleText(conn, payload)
		} else {
			go serv.HandleBin(conn, payload)
		}
	}
}

//
// handleText message from client.
//
func (serv *Server) handleText(conn int, payload []byte) {
	var (
		handler RouteHandler
		err     error
		ctx     context.Context
		req     *Request
	)

	res := _resPool.Get().(*Response)
	res.Reset()

	ctx, ok := serv.Clients.ctx[conn]
	if !ok {
		err = errors.New("client context not found")
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		goto out
	}

	req = _reqPool.Get().(*Request)
	req.Reset()

	err = json.Unmarshal(payload, req)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		goto out
	}

	res.ID = req.ID

	handler, err = req.unpack(serv.routes)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = req.Target
		goto out
	}
	if handler == nil {
		res.Code = http.StatusNotFound
		res.Message = req.Target
		goto out
	}

	handler(ctx, req, res)

out:
	err = serv.SendResponse(conn, res)
	if err != nil {
		serv.ClientRemove(conn)
	}

	if req != nil {
		_reqPool.Put(req)
	}
	_resPool.Put(res)
}

//
// handleBin message from client.  This is the dummy handler, that can be
// overwritten by implementer.
//
func (serv *Server) handleBin(conn int, payload []byte) {
}

//
// handleClose request from client.
//
func (serv *Server) handleClose(conn int, req *Frame) {
	req.opcode = opcodeClose
	req.masked = 0

	res := req.pack(false)

	_, err := unix.Write(conn, res)
	if err != nil {
		log.Println("websocket: server.handleClose: " + err.Error())
	}

	serv.ClientRemove(conn)
}

//
// handleBadRequest by sending Close frame with status.
//
func (serv *Server) handleBadRequest(conn int) {
	frameClose := NewFrameClose(false, StatusBadRequest, nil)

	err := Send(conn, frameClose)
	if err != nil {
		log.Println("websocket: server.handleBadRequest: " + err.Error())
	}

	_, err = Recv(conn)
	if err != nil {
		log.Println("websocket: server.handleBadRequest: " + err.Error())
	}

	serv.ClientRemove(conn)
}

//
// handlePing from client by sending pong response.
//
//```RFC6455
// (5.5.3.P3)
// A Pong frame sent in response to a Ping frame must have identical
// "Application data" as found in the message body of the Ping frame
// being replied to.
//```
//
func (serv *Server) handlePing(conn int, req *Frame) {
	req.opcode = opcodePong
	req.masked = 0

	res := req.pack(false)

	_, err := unix.Write(conn, res)
	if err != nil {
		log.Println("websocket: server.handlePing: " + err.Error())
		serv.ClientRemove(conn)
		return
	}
}

//
// reader read request from client.
//
// To avoid confusing network intermediaries (such as intercepting proxies)
// and for security reasons that are further discussed in Section 10.3, a
// client MUST mask all frames that it sends to the server (see Section 5.3
// for further details).  (Note that masking is done whether or not the
// WebSocket Protocol is running over TLS.)  The server MUST close the
// connection upon receiving a frame that is not masked.  In this case, a
// server MAY send a Close frame with a status code of 1002 (protocol error)
// as defined in Section 7.4.1. (RFC 6455, section 5.1, P27).
//
func (serv *Server) reader() {
	var (
		events    [128]unix.EpollEvent
		isClosing bool
	)

	for {
		nevents, err := unix.EpollWait(serv.epollRead, events[:], -1)
		if err != nil {
			log.Println("websocket: server.reader: unix.EpollWait: " + err.Error())
			break
		}

		for x := 0; x < nevents; x++ {
			conn := int(events[x].Fd)

			packet, err := Recv(conn)
			if err != nil || len(packet) == 0 {
				serv.ClientRemove(conn)
				continue
			}

			frames := Unpack(packet)
			if frames == nil {
				serv.ClientRemove(conn)
				continue
			}

			isClosing = false
			for _, frame := range frames.v {
				if frame.masked != frameIsMasked {
					serv.handleBadRequest(conn)
					isClosing = true
					break
				}

				switch frame.opcode {
				case opcodeCont:
					serv.handleFragment(conn, frame)
				case opcodeText:
					serv.handleFragment(conn, frame)
				case opcodeBin:
					serv.handleFragment(conn, frame)
				case opcodeClose:
					serv.handleClose(conn, frame)
					isClosing = true
				case opcodePing:
					serv.handlePing(conn, frame)
				case opcodePong:
					// Ignore pong from client.
				}
				if isClosing {
					break
				}
			}

			if !isClosing {
				// See https://idea.popcount.org/2017-02-20-epoll-is-fundamentally-broken-12/
				events[x].Events = unix.EPOLLIN | unix.EPOLLONESHOT

				err = unix.EpollCtl(serv.epollRead, unix.EPOLL_CTL_MOD, conn, &events[x])
				if err != nil {
					log.Println("websocket: server.reader: unix.EpollCtl: " + err.Error())
					continue
				}
			}
		}
	}
}

//
// pinger is a routine that send control PING frame to all client connections
// every N seconds.
//
func (serv *Server) pinger() {
	ticker := time.NewTicker(16 * time.Second)
	framePing := NewFramePing(false, nil)

	for range ticker.C {
		serv.Clients.Lock()

		if len(serv.Clients.all) == 0 {
			serv.Clients.Unlock()
			continue
		}

		// Make a copy of all client connections to prevent race
		// condition.
		conns := make([]int, len(serv.Clients.all))
		copy(conns, serv.Clients.all)

		serv.Clients.Unlock()

		for _, conn := range conns {
			_, err := unix.Write(conn, framePing)
			if err != nil {
				// Error on sending PING will be assumed as
				// bad connection.
				serv.ClientRemove(conn)
			}
		}
	}
}

//
// Start accepting incoming connection from clients.
//
func (serv *Server) Start() {
	go serv.upgrader()
	go serv.reader()
	go serv.pinger()

	for {
		conn, _, err := unix.Accept(serv.sock)
		if err != nil {
			log.Println("websocket: unix.Accept: " + err.Error())
			return
		}

		serv.chUpgrade <- conn
	}
}

//
// SendResponse to client.
//
func (serv *Server) SendResponse(conn int, res *Response) (err error) {
	resb, err := json.Marshal(res)
	if err != nil {
		log.Println("websocket: server.SendResponse: " + err.Error())
		return
	}

	_, err = unix.Write(conn, resb)
	if err != nil {
		log.Println("websocket: server.SendResponse: " + err.Error())
	}

	return
}
