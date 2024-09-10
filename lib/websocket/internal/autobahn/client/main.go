// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package client provide a program to test WebSocket client implementation
// with autobahn testsuite.
package main

import (
	"fmt"
	"log"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/websocket"
	"git.sr.ht/~shulhan/pakakeh.go/lib/websocket/internal/autobahn"
)

func main() {
	var (
		x int
	)

	for x = 1; x <= 301; x++ {
		clientTestCase(x)
	}

	clientUpdateReports()
	time.Sleep(1 * time.Second)
	autobahn.PrintReports(`./client/testdata/index.json`)
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
		Endpoint: `ws://0.0.0.0:9001/updateReports?agent=libwebsocket`,
	}

	var err = cl.Connect()
	if err != nil {
		log.Fatal(`clientUpdateReports:`, err)
	}
	_ = cl.Close()
}
