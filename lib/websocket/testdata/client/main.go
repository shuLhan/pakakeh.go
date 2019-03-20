// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package client provide a program to test WebSocket client implementation
// with autobahn testsuite.
package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/shuLhan/share/lib/websocket"
)

func main() {
	var wg sync.WaitGroup
	var done int
	var clients []*websocket.Client = make([]*websocket.Client, 0, 301)

	handleBin := func(cl *websocket.Client, frame *websocket.Frame) (err error) {
		err = cl.SendBin(frame.Payload())
		if err != nil {
			log.Fatal("client: SendBin: " + err.Error())
		}
		return
	}
	handleText := func(cl *websocket.Client, frame *websocket.Frame) (err error) {
		err = cl.SendText(frame.Payload())
		if err != nil {
			log.Fatal("client: SendText: " + err.Error())
		}
		return
	}
	handleQuit := func() {
		done++
		wg.Done()
		fmt.Printf(">>> DONE %d\n", done)
	}

	for x := 1; x <= 301; x++ {
		cl := &websocket.Client{
			Endpoint:   fmt.Sprintf("ws://127.0.0.1:9001/runCase?case=%d&agent=libwebsocket", x),
			HandleBin:  handleBin,
			HandleText: handleText,
			HandleQuit: handleQuit,
		}

		wg.Add(1)
		err := cl.Connect()
		if err != nil {
			log.Fatal(err)
		}

		clients = append(clients, cl)
	}

	wg.Wait()
}
