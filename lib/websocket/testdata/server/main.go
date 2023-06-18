// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package server provide a program for testing WebSocket server implement
// with autobahn testsuite.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"

	"github.com/shuLhan/share/lib/websocket"
)

const (
	cmdShutdown = `shutdown`
)

// handleBin from websocket by echo-ing back the payload.
func main() {
	var (
		srv *websocket.Server
		err error
	)

	flag.Parse()

	var cmd = flag.Arg(0)
	if cmd == cmdShutdown {
		doShutdown()
		return
	}

	log.SetFlags(0)

	var opts = &websocket.ServerOptions{
		Address: `0.0.0.0:9001`,
		HandleBin: func(conn int, payload []byte) {
			var (
				packet []byte = websocket.NewFrameBin(false, payload)
				err    error
			)

			err = websocket.Send(conn, packet)
			if err != nil {
				log.Println("handleBin: " + err.Error())
			}
		},

		HandleText: func(conn int, payload []byte) {
			var (
				packet []byte = websocket.NewFrameText(false, payload)
				err    error
			)
			err = websocket.Send(conn, packet)
			if err != nil {
				log.Println("handleText: " + err.Error())
			}

			if bytes.Equal(payload, []byte(cmdShutdown)) {
				log.Println(`Shutting down server...`)
				srv.Stop()
			}
		},
	}

	srv = websocket.NewServer(opts)

	log.Printf(`Running test server at %s`, opts.Address)

	err = srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func doShutdown() {
	var (
		logp = `doShutdown`
		cl   = &websocket.Client{
			Endpoint: fmt.Sprintf(`ws://0.0.0.0:9001`),
		}
	)

	var err = cl.Connect()
	if err != nil {
		log.Fatalf(`%s: %s`, logp, err)
	}

	err = cl.SendText([]byte(cmdShutdown))
	if err != nil {
		log.Fatalf(`%s: %s`, logp, err)
	}
}
