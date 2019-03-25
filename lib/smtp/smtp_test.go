// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"log"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const (
	testAddress           = "127.0.0.1:2525"
	testDomain            = "mail.kilabit.local"
	testPassword          = "secret"
	testTLSAddress        = "127.0.0.1:2533"
	testClientSMTPAddress = "smtp://127.0.0.1:2525"
	testSMTPSAddress      = "smtps://127.0.0.1:2533"
)

var (
	testClient        *Client  // nolint: gochecknoglobals
	testServer        *Server  // nolint: gochecknoglobals
	testAccountFirst  *Account // nolint: gochecknoglobals
	testAccountSecond *Account // nolint: gochecknoglobals
)

func TestMain(m *testing.M) {
	var err error

	testAccountFirst, err = NewAccount("First Tester", "first", testDomain, testPassword)
	if err != nil {
		log.Fatal(err)
	}
	testAccountSecond, err = NewAccount("Second Tester", "second", testDomain, testPassword)
	if err != nil {
		log.Fatal(err)
	}

	primaryDomain := NewDomain(testDomain)
	primaryDomain.Accounts["first"] = testAccountFirst
	primaryDomain.Accounts["second"] = testAccountSecond

	env := &Environment{
		PrimaryDomain: primaryDomain,
	}

	testServer = &Server{
		Address:    testAddress,
		TLSAddress: testTLSAddress,
		Env:        env,
		Handler:    NewLocalHandler(env),
	}

	err = testServer.LoadCertificate(
		"testdata/mail.kilabit.local.chain.cert.pem",
		"testdata/mail.kilabit.local.key.pem",
	)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		e := testServer.Start()
		if e != nil {
			log.Fatal("ListenAndServe:", e.Error())
		}
	}()

	testClient, err = NewClient(testSMTPSAddress, true)
	if err != nil {
		log.Fatal(err)
	}

	s := m.Run()

	os.Exit(s)
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
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "mailbox", c.exp, string(got), true)
	}
}
