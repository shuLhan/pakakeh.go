// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package client provide a program to test WebSocket client implementation
// with autobahn testsuite.
package main

import (
	"fmt"
	"log"

	"github.com/shuLhan/share/lib/websocket"
)

func main() {
	var (
		x int
	)

	for x = 1; x <= 301; x++ {
		clientTestCase(x)
	}

	clientUpdateReports()
}

func clientTestCase(testnum int) {
	log.Printf(`Running test case %d`, testnum)

	var (
		chQuit = make(chan struct{}, 1)
		cl     = &websocket.Client{
			Endpoint: fmt.Sprintf(`ws://0.0.0.0:9001/runCase?agent=libwebsocket&case=%d`, testnum),

			HandleBin: func(cl *websocket.Client, frame *websocket.Frame) (err error) {
				err = cl.SendBin(frame.Payload())
				if err != nil {
					log.Fatal("client: SendBin: " + err.Error())
				}
				return
			},

			HandleText: func(cl *websocket.Client, frame *websocket.Frame) (err error) {
				var payload = frame.Payload()
				err = cl.SendText(payload)
				if err != nil {
					log.Fatal("client: SendText: " + err.Error())
				}
				return
			},

			HandleQuit: func() {
				chQuit <- struct{}{}
			},
		}
	)

	var err = cl.Connect()
	if err != nil {
		log.Fatal(err)
	}
	<-chQuit
	log.Printf(`--- DONE %d`, testnum)
}

func clientUpdateReports() {
	var cl = &websocket.Client{
		Endpoint: fmt.Sprintf(`ws://0.0.0.0:9001/updateReports?agent=libwebsocket`),
	}

	var err = cl.Connect()
	if err != nil {
		log.Fatal(`clientUpdateReports:`, err)
	}
	_ = cl.Close()
}
