// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"log"
	"os"
	"testing"
	"time"
)

const (
	testServerAddress    = "127.0.0.1:5300"
	testDoTServerAddress = "127.0.0.1:8053"
	testTLSPort          = 8053
)

var (
	_testServer *Server
)

func TestMain(m *testing.M) {
	log.SetFlags(0)

	var (
		serverOptions = &ServerOptions{
			ListenAddress:    "127.0.0.1:5300",
			HTTPPort:         8443,
			TLSPort:          testTLSPort,
			TLSCertFile:      "testdata/domain.crt",
			TLSPrivateKey:    "testdata/domain.key",
			TLSAllowInsecure: true,
		}

		zoneFile *Zone
		err      error
	)

	_testServer, err = NewServer(serverOptions)
	if err != nil {
		log.Fatal(err)
	}

	zoneFile, err = ParseZoneFile("testdata/kilabit.info", "", 0)
	if err != nil {
		log.Fatal(err)
	}

	_testServer.Caches.InternalPopulate(zoneFile.messages, zoneFile.Path)

	go func() {
		err = _testServer.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for all listeners running.
	time.Sleep(500 * time.Millisecond)

	os.Exit(m.Run())
}
