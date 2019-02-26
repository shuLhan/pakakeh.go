// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"encoding/base64"
	"net"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestNewClient(t *testing.T) {
	cases := []struct {
		desc   string
		raddr  string
		expErr string
	}{{
		desc:   "With invalid IP",
		raddr:  "!",
		expErr: "lookup !: no such host",
	}, {
		desc:   "With no MX",
		raddr:  "example.com",
		expErr: "",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_, err := NewClient(c.raddr)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
		}
	}
}

func TestConnect(t *testing.T) {
	time.Sleep(1 * time.Second)

	expRes := &Response{
		Code:    220,
		Message: testServer.Env.PrimaryDomain.Name,
	}

	res, err := testClient.Connect(true)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "connect", expRes, res, true)
}

func TestEhlo(t *testing.T) {
	cases := []struct {
		desc          string
		arg           string
		exp           *Response
		expServerInfo *ServerInfo
	}{{
		desc: "With no argument",
		exp: &Response{
			Code:    StatusOK,
			Message: "mail.kilabit.local",
			Body: []string{
				"DSN",
				"AUTH PLAIN",
			},
		},
		expServerInfo: &ServerInfo{
			Domain: "mail.kilabit.local",
			Exts: []string{
				"dsn",
				"auth",
			},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := testClient.Ehlo(c.arg)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Ehlo", c.exp, got, true)
		test.Assert(t, "ServerInfo", c.expServerInfo,
			testClient.ServerInfo, true)
	}
}

func TestAuth(t *testing.T) {
	cases := []struct {
		desc     string
		mech     Mechanism
		username string
		password string
		expErr   string
		exp      *Response
	}{{
		desc:     "With invalid mechanism",
		username: testAccountFirst.Short(),
		password: testPassword,
		expErr:   "client.Authenticate: unknown mechanism",
	}, {
		desc:     "With invalid credential",
		mech:     MechanismPLAIN,
		username: testAccountFirst.Short(),
		password: "invalid",
		exp: &Response{
			Code:    StatusInvalidCredential,
			Message: "5.7.8 Authentication credentials invalid",
		},
	}, {
		desc:     "With valid credential",
		mech:     MechanismPLAIN,
		username: testAccountFirst.Short(),
		password: testPassword,
		exp: &Response{
			Code:    StatusAuthenticated,
			Message: "2.7.0 Authentication successful",
		},
	}, {
		desc:     "With valid credential again",
		mech:     MechanismPLAIN,
		username: testAccountFirst.Short(),
		password: testPassword,
		exp: &Response{
			Code:    StatusCmdBadSequence,
			Message: "Bad sequence of commands",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := testClient.Authenticate(c.mech, c.username, c.password)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Response", c.exp, got, true)
	}
}

func TestAuth2(t *testing.T) {
	cl, err := NewClient(testTLSAddress)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cl.Connect(true)
	if err != nil {
		t.Fatal(err)
	}

	cmd := "AUTH PLAIN\r\n"
	res, err := cl.SendCommand([]byte(cmd))
	if err != nil {
		t.Fatal(err)
	}

	exp := &Response{
		Code: StatusAuthReady,
	}
	test.Assert(t, "Response", exp, res, true)

	cred := []byte("\x00" + testAccountFirst.Short() + "\x00" + testPassword)
	cmd = base64.StdEncoding.EncodeToString(cred)

	res, err = cl.SendCommand([]byte(cmd))
	if err != nil {
		t.Fatal(err)
	}

	exp = &Response{
		Code:    StatusAuthenticated,
		Message: "2.7.0 Authentication successful",
	}
	test.Assert(t, "Response", exp, res, true)
}

func TestExpand(t *testing.T) {
	cases := []struct {
		desc  string
		mlist string
		exp   *Response
	}{{
		desc:  "With mailing-list",
		mlist: "mailing-list@test",
		exp: &Response{
			Code:    StatusCmdNotImplemented,
			Message: "Command not implemented",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := testClient.Expand(c.mlist)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Expand", c.exp, got, true)
	}
}

func TestHelp(t *testing.T) {
	cases := []struct {
		desc string
		arg  string
		exp  *Response
	}{{
		desc: "Without any argument",
		exp: &Response{
			Code:    StatusHelp,
			Message: "Everything will be alright",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := testClient.Help(c.arg)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Help", c.exp, got, true)
	}
}

func TestSendCommand(t *testing.T) {
	cases := []struct {
		desc string
		cmd  []byte
		exp  *Response
	}{{
		desc: "Send HELO",
		cmd:  []byte("HELO 192.168.10.1\r\n"),
		exp: &Response{
			Code:    StatusOK,
			Message: "mail.kilabit.local",
		},
	}, {
		desc: "Send NOOP",
		cmd:  []byte("NOOP\r\n"),
		exp: &Response{
			Code:    StatusOK,
			Message: "OK",
		},
	}, {
		desc: "Send RSET",
		cmd:  []byte("RSET\r\n"),
		exp: &Response{
			Code:    StatusOK,
			Message: "OK",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := testClient.SendCommand(c.cmd)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "SendCommand", c.exp, got, true)
	}
}

func TestLookup(t *testing.T) {
	cases := []struct {
		desc    string
		address string
		exp     net.IP
		expErr  string
	}{{
		desc:   "With empty address",
		expErr: "lookup : no such host",
	}, {
		desc:    "With MX",
		address: "kilabit.info",
		exp:     net.ParseIP("103.200.4.162"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := lookup(c.address)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "IP", c.exp.String(), got.String(), true)
	}
}

func TestMailTx(t *testing.T) {
	cases := []struct {
		desc   string
		mail   *MailTx
		exp    *Response
		expErr string
	}{{
		desc: "With empty mail",
	}, {
		desc:   "With empty From",
		mail:   &MailTx{},
		expErr: "SendMailTx: empty mail 'From' parameter",
	}, {
		desc: "With empty Recipients",
		mail: &MailTx{
			From: "ms@localhost",
		},
		expErr: "SendMailTx: empty mail 'Recipients' parameter",
	}, {
		desc: "With no data",
		mail: &MailTx{
			From: "ms@localhost",
			Recipients: []string{
				"root@localhost",
			},
		},
		exp: &Response{
			Code:    StatusOK,
			Message: "OK",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := testClient.MailTx(c.mail)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Response", c.exp, got, true)
	}
}

func TestVerify(t *testing.T) {
	cases := []struct {
		desc    string
		mailbox string
		exp     *Response
	}{{
		desc:    "With mailbox exist",
		mailbox: testAccountFirst.Short(),
		exp: &Response{
			Code:    StatusOK,
			Message: "First Tester <first@mail.kilabit.local>",
		},
	}, {
		desc:    "With mailbox not exist",
		mailbox: "notexist@mail",
		exp: &Response{
			Code:    StatusMailboxNotFound,
			Message: "mailbox not found",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := testClient.Verify(c.mailbox)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Verify", c.exp, got, true)
	}
}

func TestQuit(t *testing.T) {
	exp := &Response{
		Code:    StatusClosing,
		Message: "Service closing transmission channel",
	}

	got, err := testClient.Quit()
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Quit", exp, got, true)
}
