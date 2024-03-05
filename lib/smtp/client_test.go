// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"encoding/base64"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

// Test client using live server.
func TestClient_live(t *testing.T) {
	t.Skip(`live testing`)

	var (
		clientOpts = ClientOptions{
			ServerURL:     os.Getenv(`SMTP_SERVER`),
			AuthUser:      os.Getenv(`SMTP_USER`),
			AuthPass:      os.Getenv(`SMTP_PASS`),
			AuthMechanism: SaslMechanismPlain,
			Insecure:      true,
		}

		cl    *Client
		tdata *test.Data
		err   error
	)

	t.Logf(`clientOpts: %+v`, clientOpts)

	tdata, err = test.LoadData(`testdata/client_live_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	cl, err = NewClient(clientOpts)
	if err != nil {
		t.Fatal(err)
	}

	var (
		from     = clientOpts.AuthUser
		to       = []string{`Shulhan <m.shulhan@gmail.com>`}
		subject  = tdata.Input[`subject`]
		bodyText = tdata.Input[`bodyText`]
		bodyHTML = tdata.Input[`bodyHtml`]
	)

	err = cl.SendEmail(from, to, subject, bodyText, bodyHTML)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEhlo(t *testing.T) {
	cases := []struct {
		exp           *Response
		expServerInfo *ServerInfo

		desc string
		arg  string
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
			Info:   "mail.kilabit.local",
			Exts: map[string][]string{
				"dsn": nil,
				"auth": {
					"PLAIN",
				},
			},
		},
	}}

	var (
		cl *Client
	)

	for _, c := range cases {
		t.Log(c.desc)

		cl = testNewClient(false)

		got, err := cl.ehlo(c.arg)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf(`got: %+v`, got)
		test.Assert(t, "Ehlo", c.exp, got)
		test.Assert(t, "ServerInfo.Domain", c.expServerInfo.Domain, cl.ServerInfo.Domain)
		test.Assert(t, "ServerInfo.Info", c.expServerInfo.Info, cl.ServerInfo.Info)
	}
}

func TestAuth(t *testing.T) {
	cases := []struct {
		exp *Response

		desc     string
		username string
		password string
		expErr   string

		mech SaslMechanism
	}{{
		desc:     "With invalid mechanism",
		username: testAccountFirst.Short(),
		password: testPassword,
		expErr:   "client.Authenticate: unknown mechanism",
	}, {
		desc:     "With invalid credential",
		mech:     SaslMechanismPlain,
		username: testAccountFirst.Short(),
		password: "invalid",
		exp: &Response{
			Code:    StatusInvalidCredential,
			Message: "5.7.8 Authentication credentials invalid",
		},
	}, {
		desc:     "With valid credential",
		mech:     SaslMechanismPlain,
		username: testAccountFirst.Short(),
		password: testPassword,
		exp: &Response{
			Code:    StatusAuthenticated,
			Message: "2.7.0 Authentication successful",
		},
	}, {
		desc:     "With valid credential again",
		mech:     SaslMechanismPlain,
		username: testAccountFirst.Short(),
		password: testPassword,
		exp: &Response{
			Code:    StatusCmdBadSequence,
			Message: "Bad sequence of commands",
		},
	}}

	var cl = testNewClient(false)

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Authenticate(c.mech, c.username, c.password)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Response", c.exp, got)
	}
}

func TestAuth2(t *testing.T) {
	var (
		opts = ClientOptions{
			ServerURL: testSMTPSAddress,
			Insecure:  true,
		}

		cl  *Client
		err error
		cmd string
	)

	cl, err = NewClient(opts)
	if err != nil {
		t.Fatal(err)
	}

	cmd = "AUTH PLAIN\r\n"
	res, err := cl.SendCommand([]byte(cmd))
	if err != nil {
		t.Fatal(err)
	}

	exp := &Response{
		Code: StatusAuthReady,
	}
	test.Assert(t, "Response", exp, res)

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
	test.Assert(t, "Response", exp, res)
}

func TestExpand(t *testing.T) {
	cases := []struct {
		exp   *Response
		desc  string
		mlist string
	}{{
		desc:  "With mailing-list",
		mlist: "mailing-list@test",
		exp: &Response{
			Code:    StatusCmdNotImplemented,
			Message: "Command not implemented",
		},
	}}

	var cl = testNewClient(true)

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Expand(c.mlist)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Expand", c.exp, got)
	}
}

func TestHelp(t *testing.T) {
	cases := []struct {
		exp  *Response
		desc string
		arg  string
	}{{
		desc: "Without any argument",
		exp: &Response{
			Code:    StatusHelp,
			Message: "Everything will be alright",
		},
	}}

	var cl = testNewClient(true)

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Help(c.arg)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Help", c.exp, got)
	}
}

func TestSendCommand(t *testing.T) {
	cases := []struct {
		exp  *Response
		desc string
		cmd  []byte
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

	var cl = testNewClient(false)

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.SendCommand(c.cmd)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "SendCommand", c.exp, got)
	}
}

func TestMailTx(t *testing.T) {
	cases := []struct {
		mail *MailTx
		exp  *Response

		desc   string
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

	var cl = testNewClient(true)

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.MailTx(c.mail)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Response", c.exp, got)
	}
}

func TestVerify(t *testing.T) {
	cases := []struct {
		exp *Response

		desc    string
		mailbox string
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

	var cl = testNewClient(true)

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Verify(c.mailbox)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Verify", c.exp, got)
	}
}

func TestQuit(t *testing.T) {
	var (
		cl  = testNewClient(false)
		exp = &Response{
			Code:    StatusClosing,
			Message: "Service closing transmission channel",
		}

		got *Response
		err error
	)

	got, err = cl.Quit()
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Quit", exp, got)
}
