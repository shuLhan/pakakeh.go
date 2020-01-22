// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Program server provide an example of WebSocket server as group chat.
// The client that connect to the server must be authenticated using key.
//
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/shuLhan/share/lib/websocket"
	"github.com/shuLhan/share/lib/websocket/examples"
)

var server *websocket.Server

func main() {
	log.SetFlags(0)

	server = websocket.NewServer(9001)

	// Register the authentication handler.
	server.HandleAuth = handleAuth

	server.HandleClientAdd = handleClientAdd
	server.HandleClientRemove = handleClientRemove

	// Register the message handler
	err := server.RegisterTextHandler(http.MethodPost, "/message", handlePostMessage)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server: starting ...")

	server.Start()
}

//
// handleAuth authenticated the new connection by checking the Header "Key"
// value.
//
func handleAuth(req *websocket.Handshake) (ctx context.Context, err error) {
	key := req.Header.Get("Key")

	for id, user := range examples.Users {
		if user.Key == key {
			ctx = context.WithValue(context.Background(),
				websocket.CtxKeyUID, id)
			return ctx, nil
		}
	}

	return nil, fmt.Errorf("user's key not found")
}

//
// handleClientAdd do something when new connection accepted by server.
//
func handleClientAdd(ctx context.Context, conn int) {
	log.Printf("server: adding client %d ...", conn)

	uid := ctx.Value(websocket.CtxKeyUID).(int64)
	user := examples.Users[uid]

	// Broadcast to other connections that new user is connected.
	body := user.Name + " joining conversation..."
	packet, err := websocket.NewBroadcast(examples.BroadcastSystem, body)
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range server.Clients.All() {
		if c == conn {
			continue
		}
		err := websocket.Send(c, packet)
		if err != nil {
			log.Println(err)
		}
	}
}

//
// handleClientRemove do something when connection removed by server, either
// by client disconnected or manually removed by server itself.
//
func handleClientRemove(ctx context.Context, conn int) {
	log.Printf("server: client %d has been disconnected ...", conn)

	uid := ctx.Value(websocket.CtxKeyUID).(int64)
	user := examples.Users[uid]

	// Broadcast to other connections that user is disconnected.
	body := user.Name + " leaving conversation..."
	packet, err := websocket.NewBroadcast(examples.BroadcastSystem, body)
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range server.Clients.All() {
		if c == conn {
			continue
		}
		err := websocket.Send(c, packet)
		if err != nil {
			log.Println(err)
		}
	}
}

//
// handlePostMessage handle message that is send to server by client.
//
func handlePostMessage(ctx context.Context, req *websocket.Request) (res websocket.Response) {
	uid := ctx.Value(websocket.CtxKeyUID).(int64)
	user := examples.Users[uid]

	log.Printf("server: message from %s: %q\n", user.Name, req.Body)

	body := user.Name + ": " + req.Body
	packet, err := websocket.NewBroadcast(examples.BroadcastMessage, body)
	if err != nil {
		res.Code = http.StatusInternalServerError
		res.Body = err.Error()
		return res
	}

	// Broadcast the message to all connected clients, including our
	// connections. Remember, that user could connected through many
	// application.
	for _, conn := range server.Clients.All() {
		if conn == req.Conn {
			continue
		}
		err = websocket.Send(conn, packet)
		if err != nil {
			log.Println(err)
		}
	}

	// Set the response status to success.
	res.Code = http.StatusOK

	return res
}
