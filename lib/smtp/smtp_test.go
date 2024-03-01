// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"crypto"
	"crypto/rsa"
	"log"
	"os"
	"testing"
	"time"

	libcrypto "git.sr.ht/~shulhan/pakakeh.go/lib/crypto"
	"git.sr.ht/~shulhan/pakakeh.go/lib/email/dkim"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	testAddress         = "127.0.0.1:2525"
	testDomain          = "mail.kilabit.local"
	testPassword        = "secret"
	testTLSAddress      = "127.0.0.1:2533"
	testSMTPSAddress    = "smtps://" + testTLSAddress
	testFileCertificate = "testdata/" + testDomain + ".cert.pem"
	testFilePrivateKey  = "testdata/" + testDomain + ".key.pem"
)

var (
	testServer        *Server
	testAccountFirst  *Account
	testAccountSecond *Account
)

func testRunServer() {
	var err error

	testAccountFirst, err = NewAccount("First Tester", "first", testDomain, testPassword)
	if err != nil {
		log.Fatal(err)
	}
	testAccountSecond, err = NewAccount("Second Tester", "second", testDomain, testPassword)
	if err != nil {
		log.Fatal(err)
	}

	var primaryKey crypto.PrivateKey

	primaryKey, err = libcrypto.LoadPrivateKey(testFilePrivateKey, nil)
	if err != nil {
		log.Fatal(err)
	}

	var (
		rsakey *rsa.PrivateKey
		ok     bool
	)

	rsakey, ok = primaryKey.(*rsa.PrivateKey)
	if !ok {
		log.Panicf(`%s: expecting RSA private key, got %T`, testFilePrivateKey, primaryKey)
	}

	primaryDKIMOpts := &DKIMOptions{
		Signature:  dkim.NewSignature([]byte(testDomain), []byte("default")),
		PrivateKey: rsakey,
	}
	primaryDomain := NewDomain(testDomain, primaryDKIMOpts)
	primaryDomain.Accounts["first"] = testAccountFirst
	primaryDomain.Accounts["second"] = testAccountSecond

	env := &Environment{
		PrimaryDomain: primaryDomain,
	}

	testServer = &Server{
		address:    testAddress,
		tlsAddress: testTLSAddress,
		Env:        env,
		Handler:    NewLocalHandler(env),
	}

	err = testServer.LoadCertificate(
		testFileCertificate,
		testFilePrivateKey,
	)
	if err != nil {
		log.Fatal("testServer.LoadCertificate: " + err.Error())
	}

	go func() {
		err = testServer.Start()
		if err != nil {
			log.Fatal("ListenAndServe:" + err.Error())
		}
	}()
}

func testNewClient(withAuth bool) (cl *Client) {
	var (
		opts = ClientOptions{
			ServerUrl: testSMTPSAddress,
			Insecure:  true,
		}

		err error
	)
	if withAuth {
		opts.AuthUser = testAccountFirst.Short()
		opts.AuthPass = testPassword
		opts.AuthMechanism = SaslMechanismPlain
	}

	cl, err = NewClient(opts)
	if err != nil {
		log.Fatal(err)
	}

	return cl
}

func TestMain(m *testing.M) {
	testRunServer()

	// Wait for server to be alive.
	var (
		timeout = 5 * time.Second
		err     error
	)
	err = libnet.WaitAlive(`tcp`, testAddress, timeout)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestParsePath(t *testing.T) {
	cases := []struct {
		desc   string
		path   string
		exp    string
		expErr string
	}{{
		desc:   "With empth path",
		expErr: "ParsePath: empty path",
	}, {
		desc: "With null path",
		path: "<>",
	}, {
		desc:   "Without '<'",
		path:   "  local@domain>",
		expErr: "ParsePath: missing opening '<'",
	}, {
		desc:   "Without '>'",
		path:   "<local@domain",
		expErr: "ParsePath: missing closing '>'",
	}, {
		desc:   "Without mailbox",
		path:   "<@domain:>",
		expErr: "ParsePath: invalid mailbox format",
	}, {
		desc: "Without domain",
		path: "<@domain:local>",
		exp:  "local",
	}, {
		desc: "With source-route",
		path: "<@domain:local@domain>",
		exp:  "local@domain",
	}, {
		desc: "With two source-routes",
		path: "<@domain,@domain:local@domain>",
		exp:  "local@domain",
	}, {
		desc: "With comment on local-part",
		path: "<@domain,@domain:local(comment)@domain>",
		exp:  "local@domain",
	}, {
		desc: "With comment on domain",
		path: "<@domain,@domain:local(comment)@domain(comment)>",
		exp:  "local@domain",
	}, {
		desc:   "With double dot on local",
		path:   "<@domain,@domain:local..part(comment)@domain(comment)>",
		expErr: "ParsePath: invalid mailbox format",
	}, {
		desc:   "With double dot on domain",
		path:   "<@domain,@domain:local..part(comment)@domain..com(comment)>",
		expErr: "ParsePath: invalid mailbox format",
	}, {
		desc: "With address literal",
		path: "<@domain,@domain:local(comment)@[127.0.0.1]>",
		exp:  "local@[127.0.0.1]",
	}, {
		desc: "With quoted-string",
		path: `<@domain,@domain:"local\ \\\"(comment)"@[127.0.0.1]>`,
		exp:  `"local \"(comment)"@[127.0.0.1]`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := ParsePath([]byte(c.path))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "mailbox", c.exp, string(got))
	}
}
