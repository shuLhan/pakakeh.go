// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package dns

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"
)

const (
	testServerAddress    = "127.0.0.1:5300"
	testDoTServerAddress = "127.0.0.1:18053"
	testTLSPort          = 18053

	// Equal to 2023-08-05 07:53:20 +0000 UTC.
	testNowEpoch = 1691222000
)

var (
	_testServer *Server
)

func TestMain(m *testing.M) {
	log.SetFlags(0)

	var flagNoServer bool

	flag.BoolVar(&flagNoServer, `no-server`, false, `Skip running servers`)
	flag.Parse()

	timeNow = func() time.Time {
		return time.Unix(testNowEpoch, 0)
	}

	if !flagNoServer {
		runServer()
	}

	os.Exit(m.Run())
}

func runServer() {
	var (
		serverOptions = &ServerOptions{
			ListenAddress:    "127.0.0.1:5300",
			HTTPPort:         8443,
			TLSPort:          testTLSPort,
			TLSCertFile:      "testdata/domain.crt",
			TLSPrivateKey:    "testdata/domain.key",
			TLSAllowInsecure: true,
		}

		err error
	)

	_testServer, err = NewServer(serverOptions)
	if err != nil {
		log.Fatal(err)
	}

	var zoneFile *Zone

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
}
