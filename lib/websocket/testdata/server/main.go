// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package server provide a program for testing WebSocket server implement
// with autobahn testsuite.
package main

import (
	"log"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/websocket"
)

// handleBin from websocket by echo-ing back the payload.
func handleBin(conn int, payload []byte) {
	var (
		packet []byte = websocket.NewFrameBin(false, payload)
		err    error
	)

	err = websocket.Send(conn, packet)
	if err != nil {
		log.Println("handleBin: " + err.Error())
	}
}

// handleText from websocket by echo-ing back the payload.
func handleText(conn int, payload []byte) {
	var (
		packet []byte = websocket.NewFrameText(false, payload)
		err    error
	)

	if debug.Value >= 3 {
		log.Printf("testdata/server: handleText: {payload.len:%d}\n", len(payload))
	}

	err = websocket.Send(conn, packet)
	if err != nil {
		log.Println("handleText: " + err.Error())
	}
}

func main() {
	var (
		opts = &websocket.ServerOptions{
			Address:    "127.0.0.1:9001",
			HandleBin:  handleBin,
			HandleText: handleText,
		}
		srv = websocket.NewServer(opts)

		err error
	)

	err = srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
