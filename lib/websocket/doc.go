// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package websocket provide a library for creating WebSocket server or
// client.
//
// The websocket server is implemented with epoll and kqueue, which means it's
// only run on Linux, Darwin, or BSD.
//
// Pub-Sub Example
//
// The following example show how create an authenticated WebSocket server
// that echo the data frame TEXT back to client.
//
//	import (
//		...
//
//		"github.com/shuLhan/share/lib/websocket"
//	)
//
//	var srv *websocket.Server
//
//	func handleAuth(req *Handshake) (ctx context.Context, err error) {
//		URL, err := url.ParseRequestURI(string(req.URI))
//		if err != nil {
//			return nil, err
//		}
//
//		q := URL.Query()
//
//		extJWT := q.Get("ticket")
//		if len(extJWT) == 0 {
//			return nil, fmt.Errorf("Missing authorization")
//		}
//
//		ctx = context.WithValue(context.Background(), CtxKeyExternalJWT, extJWT)
//		ctx = context.WithValue(ctx, CtxKeyInternalJWT, _testInternalJWT)
//		ctx = context.WithValue(ctx, CtxKeyUID, _testUID)
//
//		return ctx, nil
//	}
//
//	func handleText(conn int, payload []byte) {
//		packet := websocket.NewFrameText(false, payload)
//
//		ctx := srv.Clients.Context(conn)
//
//		// ... do something with connection context "ctx"
//
//		err := websocket.Send(conn, packet)
//		if err != nil {
//			log.Println("handleText: " + err.Error())
//		}
//	}
//
//	func main() {
//		opts := &ServerOptions{
//			Address: ":9001",
//			HandleAuth: handleAuth,
//			HandleText: handleText,
//		}
//		srv, err := websocket.NewServer(opts)
//		if err != nil {
//			log.Println("websocket: " + err.Error())
//			os.Exit(2)
//		}
//
//		srv.Start()
//	}
//
// References
//
// - https://tools.ietf.org/html/rfc6455
//
// - https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers
//
// - http://man7.org/linux/man-pages/man7/epoll.7.html
//
package websocket
