// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
	"log"
	"os"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

const (
	testServerAddress = "127.0.0.1:5300"
)

//nolint:gochecknoglobals
var (
	_testServer *Server
)

func TestMain(m *testing.M) {
	var err error

	log.SetFlags(0)

	cert, err := tls.LoadX509KeyPair("testdata/domain.crt", "testdata/domain.key")
	if err != nil {
		log.Fatal("dns: error loading certificate: " + err.Error())
	}

	serverOptions := &ServerOptions{
		ListenAddress:    "127.0.0.1:5300",
		HTTPPort:         8443,
		TLSCertificate:   &cert,
		TLSAllowInsecure: true,
	}

	_testServer, err = NewServer(serverOptions)
	if err != nil {
		log.Fatal(err)
	}

	_testServer.LoadMasterFile("testdata/kilabit.info")

	_testServer.Start()

	// Wait for all listeners running.
	time.Sleep(500 * time.Millisecond)

	s := m.Run()

	os.Exit(s)
}

func TestQueryType(t *testing.T) {
	test.Assert(t, "QueryTypeA", QueryTypeA, uint16(1), true)
	test.Assert(t, "QueryTypeTXT", QueryTypeTXT, uint16(16), true)
	test.Assert(t, "QueryTypeAXFR", QueryTypeAXFR, uint16(252), true)
	test.Assert(t, "QueryTypeALL", QueryTypeALL, uint16(255), true)
}
