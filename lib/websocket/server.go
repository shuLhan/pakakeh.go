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
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const (
	_maxQueueUpgrade = 128

	_resUpgradeOK = "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-Websocket-Accept: %s\r\n\r\n"
)

//
// Server for websocket.
//
type Server struct {
	sock      int
	chUpgrade chan int
	epollRead int
	clients   sync.Map
	fragments map[int]*Frame
	routes    *rootRoute

	HandleText         HandlerFn
	HandleBin          HandlerFn
	HandleClose        HandlerFn
	HandlePing         HandlerFn
	HandleAuth         HandlerAuthFn
	HandleClientAdd    HandlerClientFn
	HandleClientRemove HandlerClientFn
}

//
// NewServer will create new web-socket server that listen on port number.
//
func NewServer(port int) (serv *Server, err error) {
	serv = &Server{
		chUpgrade: make(chan int, _maxQueueUpgrade),
		fragments: make(map[int]*Frame),
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
	serv.HandleClose = serv.handleClose
	serv.HandlePing = serv.handlePing
	serv.HandleClientAdd = func(ctx context.Context, conn int) {}
	serv.HandleClientRemove = func(ctx context.Context, conn int) {}

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

	err = unix.Listen(serv.sock, _maxQueueUpgrade)

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
		fmt.Fprintln(os.Stderr, err.Error())
	}

	unix.Close(conn)
}

//
// handleUpgrade parse and validate websocket HTTP handshake.
//
func (serv *Server) handleUpgrade(httpRequest []byte) (
	ctx context.Context, req *Handshake, err error,
) {
	req = _handshakePool.Get().(*Handshake)

	err = req.Parse(httpRequest)
	if err != nil {
		_handshakePool.Put(req)
		req = nil
		return
	}

	// (7)
	if serv.HandleAuth != nil {
		ctx, err = serv.HandleAuth(req)
		if err != nil {
			_handshakePool.Put(req)
			req = nil
			return
		}
	} else {
		ctx = context.Background()
	}

	return
}

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

	serv.clients.Store(conn, ctx)

	return
}

func (serv *Server) clientRemove(conn int) {
	v, ok := serv.clients.Load(conn)
	if ok {
		ctx := v.(context.Context)
		go serv.HandleClientRemove(ctx, conn)
	}

	serv.clients.Delete(conn)
	delete(serv.fragments, conn)

	err := unix.EpollCtl(serv.epollRead, unix.EPOLL_CTL_DEL, conn, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	serv.clientClose(conn)
}

func (serv *Server) clientClose(conn int) {
	err := unix.Close(conn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (serv *Server) upgrader() {
	for conn := range serv.chUpgrade {
		packet, err := Recv(conn)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			unix.Close(conn)
			continue
		}
		if len(packet) == 0 {
			unix.Close(conn)
			continue
		}

		ctx, req, err := serv.handleUpgrade(packet)
		if err != nil {
			serv.handleError(conn, http.StatusBadRequest, err.Error())
			continue
		}

		wsAccept := generateHandshakeAccept(req.Key)
		_handshakePool.Put(req)

		bb := _bbPool.Get().(*bytes.Buffer)
		bb.Reset()

		fmt.Fprintf(bb, _resUpgradeOK, wsAccept)

		_, err = unix.Write(conn, bb.Bytes())
		_bbPool.Put(bb)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			unix.Close(conn)
			continue
		}

		err = serv.clientAdd(ctx, conn)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			unix.Close(conn)
			continue
		}

		serv.HandleClientAdd(ctx, conn)
	}
}

//
// handleFragment will handle continuation frame (fragmentation).
//
//
//	(5.4-P34)
//	o  A fragmented message consists of a single frame with the FIN bit
//	   clear and an opcode other than 0, followed by zero or more frames
//	   with the FIN bit clear and the opcode set to 0, and terminated by
//	   a single frame with the FIN bit set and an opcode of 0.  A
//	   fragmented message is conceptually equivalent to a single larger
//	   message whose payload is equal to the concatenation of the
//	   payloads of the fragments in order; however, in the presence of
//	   extensions, this may not hold true as the extension defines the
//	   interpretation of the "Extension data" present.  For instance,
//	   "Extension data" may only be present at the beginning of the first
//	   fragment and apply to subsequent fragments, or there may be
//	   "Extension data" present in each of the fragments that applies
//	   only to that particular fragment.  In the absence of "Extension
//	   data", the following example demonstrates how fragmentation works.
//
//	   EXAMPLE: For a text message sent as three fragments, the first
//	   fragment would have an opcode of 0x1 and a FIN bit clear, the
//	   second fragment would have an opcode of 0x0 and a FIN bit clear,
//	   and the third fragment would have an opcode of 0x0 and a FIN bit
//	   that is set.
//
// The first frame and their continuation is saved on map of socket connection
// and frame: fragments.
//
// For each request frame, there are three possible cases:
//
// (1) request is the first frame (fin = 0 && opcode != 0).
//
// request will replace any previous non-completed fragmentation.
//
// (2) request is the middle frame (fin = 0 && opcode = 0).
//
// (2.1) Check if previous fragmentation exists, if not ignore the request.
// (2.2) Append the request payload with previous frame.
//
// (3) request is the last frame (fin = 1 && opcode = 0)
//
// (3.1) Check if previous fragmentation exists, if not ignore the request.
// (3.2) Append the request payload with previous frame.
// (3.3) Handle request
// (3.4) Clear cache of fragmentations
//
func (serv *Server) handleFragment(conn int, req *Frame) {
	// (1)
	if req.opcode != opcodeCont {
		serv.fragments[conn] = req
		return
	}

	// (2.1) (3.1)
	f, ok := serv.fragments[conn]
	if !ok {
		return
	}

	// (2.2) (3.2)
	f.payload = append(f.payload, req.payload...)
	f.len += req.len

	// (2)
	if req.fin == 0 {
		return
	}

	req.fin = frameIsFinished

	// (3.3)
	if f.opcode == opcodeText {
		go serv.HandleText(conn, f)
	} else if f.opcode == opcodeBin {
		go serv.HandleBin(conn, f)
	}

	// (3.4)
	serv.fragments[conn] = nil
	delete(serv.fragments, conn)
}

//
// handleText message from client.
//
func (serv *Server) handleText(conn int, f *Frame) {
	var (
		handler RouteHandler
		err     error
		ctx     context.Context
		req     *Request
	)

	res := _resPool.Get().(*Response)
	res.Reset()

	v, ok := serv.clients.Load(conn)
	if !ok {
		err = errors.New("client context not found")
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		goto out
	}

	ctx = v.(context.Context)

	req = _reqPool.Get().(*Request)
	req.Reset()

	err = json.Unmarshal(f.payload, req)
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
		serv.clientRemove(conn)
	}

	if req != nil {
		_reqPool.Put(req)
	}
	_resPool.Put(res)
}

//
// handleBin message from client.  This is the dummy handler, that can be
// overwriten by implementor.
//
func (serv *Server) handleBin(conn int, req *Frame) {
}

//
// handleClose request from client.
//
func (serv *Server) handleClose(conn int, req *Frame) {
	req.opcode = opcodeClose
	req.masked = 0

	res := req.Pack(false)

	_, err := unix.Write(conn, res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	serv.clientRemove(conn)
}

//
// handleBadRequest by removing client connection first and then by sending
// Close frame.
//
func (serv *Server) handleBadRequest(conn int) {
	v, ok := serv.clients.Load(conn)
	if ok {
		ctx := v.(context.Context)
		go serv.HandleClientRemove(ctx, conn)
	}

	serv.clients.Delete(conn)
	delete(serv.fragments, conn)

	err := unix.EpollCtl(serv.epollRead, unix.EPOLL_CTL_DEL, conn, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		serv.clientClose(conn)
		return
	}

	resClose := NewFrameClose(false, StatusBadRequest, nil)

	_, err = unix.Write(conn, resClose)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	serv.clientClose(conn)
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

	res := req.Pack(false)

	_, err := unix.Write(conn, res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		serv.clientRemove(conn)
		return
	}
}

//
// reader read request from client.
//
//```RFC6455
// (5.1-P27)
// To avoid confusing network intermediaries (such as
// intercepting proxies) and for security reasons that are further
// discussed in Section 10.3, a client MUST mask all frames that it
// sends to the server (see Section 5.3 for further details).  (Note
// that masking is done whether or not the WebSocket Protocol is running
// over TLS.)  The server MUST close the connection upon receiving a
// frame that is not masked.  In this case, a server MAY send a Close
// frame with a status code of 1002 (protocol error) as defined in
// Section 7.4.1.
//```
//
// (1) See https://idea.popcount.org/2017-02-20-epoll-is-fundamentally-broken-12/
//
func (serv *Server) reader() {
	var (
		events [128]unix.EpollEvent
	)

	for {
		nevents, err := unix.EpollWait(serv.epollRead, events[:], -1)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			break
		}

		for x := 0; x < nevents; x++ {
			conn := int(events[x].Fd)

			packet, err := Recv(conn)
			if err != nil || len(packet) == 0 {
				serv.clientRemove(conn)
				continue
			}

			// (1)
			events[x].Events = unix.EPOLLIN | unix.EPOLLONESHOT

			err = unix.EpollCtl(serv.epollRead, unix.EPOLL_CTL_MOD, conn, &events[x])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			reqs := Unpack(packet)

			if len(reqs) == 0 {
				serv.clientRemove(conn)
				continue
			}

			for _, req := range reqs {
				// (5.1-P27)
				if req.masked != frameIsMasked {
					serv.handleBadRequest(conn)
					break
				}

				switch req.opcode {
				case opcodeCont:
					serv.handleFragment(conn, req)
				case opcodeText:
					if req.fin != frameIsFinished {
						serv.handleFragment(conn, req)
					} else {
						go serv.HandleText(conn, req)
					}
				case opcodeBin:
					if req.fin != frameIsFinished {
						serv.handleFragment(conn, req)
					} else {
						go serv.HandleBin(conn, req)
					}
				case opcodeClose:
					serv.HandleClose(conn, req)
				case opcodePing:
					serv.HandlePing(conn, req)
				case opcodePong:
					continue
				}
			}
		}
	}
}

//
// pinger iterate on all clients and send control Ping frame every N seconds.
//
func (serv *Server) pinger() {
	ticker := time.NewTicker(16 * time.Second)

	for range ticker.C {
		serv.clients.Range(func(k, _ interface{}) bool {
			conn, ok := k.(int)
			if ok {
				_, err := unix.Write(conn, NewFramePing(false, nil))
				if err != nil {
					serv.clientRemove(conn)
				}
			}

			return true
		})
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
			fmt.Fprintln(os.Stderr, err)
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
		fmt.Fprintln(os.Stderr, "SendResponse:", err.Error())
		return
	}

	_, err = unix.Write(conn, resb)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendResponse:", err.Error())
	}

	return
}
