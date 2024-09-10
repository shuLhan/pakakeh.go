// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package websocket provide a library for WebSocket server or client based on
// [RFC6455].
//
// The websocket server is implemented using epoll and kqueue, which means
// it's only run on Linux, Darwin, or BSD.
//
// # Pub-Sub Example
//
// The following code snippet show how to create an authenticated WebSocket
// server that echo the data frame TEXT back to client.
//
//	import (
//		...
//
//		"git.sr.ht/~shulhan/pakakeh.go/lib/websocket"
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
// # Limitation
//
// Only support WebSocket version 13 (the first and most common version used
// in web browser).
//
// # References
//
//   - https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers
//
//   - http://man7.org/linux/man-pages/man7/epoll.7.html
//
// [RFC6455]: https://tools.ietf.org/html/rfc6455
package websocket
