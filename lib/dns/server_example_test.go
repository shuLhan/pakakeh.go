// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"log"
	"time"
)

func clientLookup(nameserver string) {
	var (
		q = MessageQuestion{
			Name: "kilabit.info",
		}

		cl  *UDPClient
		msg *Message
		rr  ResourceRecord
		err error
		x   int
	)

	cl, err = NewUDPClient(nameserver)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err = cl.Lookup(q, false)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Receiving DNS message: %s\n", msg)
	for x, rr = range msg.Answer {
		fmt.Printf("Answer %d: %s\n", x, rr.Value)
	}
	for x, rr = range msg.Authority {
		fmt.Printf("Authority %d: %s\n", x, rr.Value)
	}
	for x, rr = range msg.Additional {
		fmt.Printf("Additional %d: %s\n", x, rr.Value)
	}
}

func ExampleServer() {
	var (
		serverAddress = "127.0.0.1:5300"
		serverOptions = &ServerOptions{
			ListenAddress:    "127.0.0.1:5300",
			HTTPPort:         8443,
			TLSCertFile:      "testdata/domain.crt",
			TLSPrivateKey:    "testdata/domain.key",
			TLSAllowInsecure: true,
		}

		server   *Server
		zoneFile *Zone
		err      error
	)

	server, err = NewServer(serverOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Load records to be served from zone file.
	zoneFile, err = ParseZoneFile("testdata/kilabit.info", "", 0)
	if err != nil {
		log.Fatal(err)
	}

	server.PopulateCaches(zoneFile.Messages(), zoneFile.Path)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for all listeners running.
	time.Sleep(500 * time.Millisecond)

	clientLookup(serverAddress)

	server.Stop()
}
