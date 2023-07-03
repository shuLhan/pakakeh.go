// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program client provide an example of chat client that connect to WebSocket
// server.
//
// To run the client as user ID 1 (Groot),
//
//	$ go run . chat 1
//
// You can open other terminal and run another clients,
//
//	$ go run . chat 2 # or
//	$ go run . chat 3
//
// and start chatting with each others.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shuLhan/share/lib/websocket"
	"github.com/shuLhan/share/lib/websocket/examples"
)

const (
	cmdChat    = `chat`
	cmdChatbot = `chatbot`
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
			Endpoint: `ws://127.0.0.1:9101`,
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

	var err error = cc.conn.Connect()
	if err != nil {
		log.Fatal("Connect: " + err.Error())
	}

	log.Printf("%s: connected ...", user.Name)

	return cc
}

// Start the chat client.
func (cc *ChatClient) Start() {
	var (
		req = &websocket.Request{
			Method: http.MethodPost,
			Target: "/message",
		}

		reader *bufio.Reader
		packet []byte
		err    error
	)

	reader = bufio.NewReader(os.Stdin)
	for {
		fmt.Print(cc.user.Name + "> ")

		req.Body, _ = reader.ReadString('\n')
		req.Body = strings.TrimSpace(req.Body)
		if len(req.Body) == 0 {
			continue
		}

		req.ID = uint64(time.Now().Unix())

		packet, err = json.Marshal(req)
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
	var (
		res = &websocket.Response{}
	)

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

func usage() {
	fmt.Println(`= WebSocket client example

	client <chat | chatbot> <args...>

== USAGE

client ` + cmdChat + ` <id>
	Connect to chat with others using specific ID: 1, 2, or 3.

client ` + cmdChatbot + ` <N>
	Connect to the server and sent N messages for each user
	simultaneously.`)
}

func main() {
	flag.Parse()

	if len(os.Args) <= 2 {
		usage()
		return
	}

	var cmd = strings.ToLower(flag.Arg(0))
	switch cmd {
	case cmdChat:
		doChat(flag.Arg(1))
	case cmdChatbot:
		doChatbot(flag.Arg(1))
	default:
		log.Fatalf(`unknown command: %s`, cmd)
	}
}

func doChat(userIDStr string) {
	var (
		user *examples.Account
		cc   *ChatClient
		err  error
		uid  int
		ok   bool
	)

	uid, err = strconv.Atoi(userIDStr)
	if err != nil {
		log.Fatal(err)
	}

	user, ok = examples.Users[int64(uid)]
	if !ok {
		log.Fatalf("invalid user id: %d", uid)
	}

	cc = NewChatClient(user)

	cc.Start()
}

func doChatbot(nStr string) {
	var (
		wg   sync.WaitGroup
		user *examples.Account
		err  error
		n    int64
	)

	n, err = strconv.ParseInt(nStr, 10, 64)
	if err != nil {
		log.Fatalf(`invalid N: %s`, err)
	}

	for _, user = range examples.Users {
		wg.Add(1)
		go runChatbot(&wg, user, n)
	}
	wg.Wait()
}

func runChatbot(wg *sync.WaitGroup, user *examples.Account, n int64) {
	var (
		req = &websocket.Request{
			Method: http.MethodPost,
			Target: `/message`,
		}

		err    error
		packet []byte
		x      int64
	)

	var cc = NewChatClient(user)

	for ; x < n; x++ {
		req.ID = uint64(time.Now().UnixNano())
		req.Body = fmt.Sprintf(`#%d Hello from %s at %d`, x, user.Name, req.ID)

		packet, err = json.Marshal(req)
		if err != nil {
			log.Fatal(err)
		}

		err = cc.conn.SendText(packet)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}
	err = cc.conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	wg.Done()
}
