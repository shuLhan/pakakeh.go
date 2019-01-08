// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"net"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestConnect(t *testing.T) {
	time.Sleep(1 * time.Second)

	expRes := &Response{
		Code:    220,
		Message: testEnv.Hostname(),
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
			},
		},
		expServerInfo: &ServerInfo{
			Domain: "mail.kilabit.local",
			Exts: []string{
				"dsn",
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
			testClient.serverInfo, true)
	}
}

func TestExpand(t *testing.T) {
	cases := []struct {
		desc  string
		mlist string
		exp   *Response
	}{{
		desc:  "With mailing-list exist",
		mlist: "list-exist",
		exp: &Response{
			Code:    StatusOK,
			Message: "List Exist",
			Body: []string{
				"Member A <member-a@mail.local>",
			},
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

		test.Assert(t, "IP", c.exp, got, true)
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
		mailbox: "exist",
		exp: &Response{
			Code:    StatusOK,
			Message: "Exist <exist@mail.local>",
		},
	}, {
		desc:    "With mailbox not exit",
		mailbox: "notexist",
		exp: &Response{
			Code:    StatusMailboxNotFound,
			Message: "No such user here",
		},
	}, {
		desc:    "With ambigous user",
		mailbox: "ambigous",
		exp: &Response{
			Code:    StatusMailboxIncorrect,
			Message: "User ambigous",
			Body: []string{
				"Ambigous A <a@mail.local>",
				"Ambigous B <b@mail.local>",
				"Ambigous C <c@mail.local>",
			},
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
