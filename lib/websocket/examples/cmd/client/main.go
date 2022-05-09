// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program client provide an example of chat client that connect to WebSocket
// server.
//
// To run the client as user ID 1 (Groot),
//
//	$ go run . 1
//
// You can open other terminal and run another clients,
//
//	$ go run . 2 # or
//	$ go run . 3
//
// and start chatting with each others.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/websocket"
	"github.com/shuLhan/share/lib/websocket/examples"
)

type ChatClient struct {
	user *examples.Account
	conn *websocket.Client
}

// NewChatClient create new WebSocket client using specific user's account.
func NewChatClient(user *examples.Account) (cc *ChatClient) {
	cc = &ChatClient{
		user: user,
		conn: &websocket.Client{
			Endpoint: "ws://127.0.0.1:9001",
			Headers: http.Header{
				"Key": []string{user.Key},
			},
		},
	}

	cc.conn.HandleText = cc.handleText
	cc.conn.HandleQuit = func() {
		log.Println("connection has been closed...")
		os.Exit(0)
	}

	err := cc.conn.Connect()
	if err != nil {
		log.Fatal("Connect: " + err.Error())
	}

	log.Printf("%s: connected ...", user.Name)

	return cc
}

// Start the chat client.
func (cc *ChatClient) Start() {
	req := &websocket.Request{
		Method: http.MethodPost,
		Target: "/message",
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(cc.user.Name + "> ")

		req.Body, _ = reader.ReadString('\n')
		req.Body = strings.TrimSpace(req.Body)
		if len(req.Body) == 0 {
			continue
		}

		req.ID = uint64(time.Now().Unix())

		packet, err := json.Marshal(req)
		if err != nil {
			log.Fatal(err)
		}

		err = cc.conn.SendText(packet)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

// handleText process response from request or broadcast from server.
func (cc *ChatClient) handleText(cl *websocket.Client, frame *websocket.Frame) (err error) {
	res := &websocket.Response{}

	err = json.Unmarshal(frame.Payload(), res)
	if err != nil {
		return err
	}

	// Print message if its a broadcast message.
	if res.ID == 0 {
		switch res.Message {
		case examples.BroadcastMessage:
			fmt.Printf("\n%s\n%s> ", res.Body, cc.user.Name)
		case examples.BroadcastSystem:
			fmt.Printf("\nsystem: %s\n%s> ", res.Body, cc.user.Name)
		}
	}

	return nil
}

func main() {
	log.SetFlags(0)

	if len(os.Args) <= 1 {
		log.Printf("client <id>")
		return
	}

	uid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	user, ok := examples.Users[int64(uid)]
	if !ok {
		log.Fatalf("invalid user id: %d", uid)
	}

	cc := NewChatClient(user)

	cc.Start()
}
