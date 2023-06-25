// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package server provide a program for testing WebSocket server implement
// with autobahn testsuite.
package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"time"

	"github.com/shuLhan/share/lib/websocket"
)

const (
	cmdShutdown = `shutdown`
)

// handleBin from websocket by echo-ing back the payload.
func main() {
	var (
		timeout = 30 * time.Second

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
				timeStart        = time.Now()
				packet    []byte = websocket.NewFrameBin(false, payload)
				err       error
			)

			err = websocket.Send(conn, packet, timeout)
			if err != nil {
				log.Println("handleBin: " + err.Error())
			}

			var elapsed = time.Since(timeStart)
			if elapsed >= timeout {
				log.Printf(`HandleBin: %s`, elapsed)
			}
		},

		HandleText: func(conn int, payload []byte) {
			var (
				timeStart        = time.Now()
				packet    []byte = websocket.NewFrameText(false, payload)
				err       error
			)

			err = websocket.Send(conn, packet, timeout)
			if err != nil {
				log.Println("handleText: " + err.Error())
			}

			var elapsed = time.Since(timeStart)
			if elapsed >= timeout {
				log.Printf(`HandleText: %s`, elapsed)
			}

			if bytes.Equal(payload, []byte(cmdShutdown)) {
				log.Println(`Shutting down server...`)
				os.Exit(0)
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
			Endpoint: `ws://0.0.0.0:9001`,
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
