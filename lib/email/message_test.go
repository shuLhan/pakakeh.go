// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"fmt"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/email/dkim"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewMultipart(t *testing.T) {
	dateInUtc = true

	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/message_newmultipart_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		from     = []byte(`a@b.c`)
		to       = []byte(`d@e.f`)
		subject  = []byte(`test`)
		bodyText = []byte(`This is plain text`)
		bodyHTML = []byte(`<b>This is body in HTML</b>`)

		msg  *Message
		msgb []byte
	)

	msg, err = NewMultipart(from, to, subject, bodyText, bodyHTML)
	if err != nil {
		t.Fatal(err)
	}

	msgb, err = msg.Pack()
	if err != nil {
		t.Fatal(err)
	}

	var (
		msgBoundary = msg.Header.Boundary()
		msgID       = msg.Header.ID()
		exp         = string(tdata.Output[`message.txt`])
	)

	exp = fmt.Sprintf(exp, msgBoundary, msgID, msgBoundary, msgBoundary, msgBoundary)

	test.Assert(t, `NewMultipart`, exp, string(msgb))
}

func TestMessageParseMessage(t *testing.T) {
	cases := []struct {
		in      string
		exp     string
		expErr  string
		expRest string
	}{{
		in:  "testdata/empty.txt",
		exp: "\r\n",
	}, {
		in:     "testdata/invalid-header.txt",
		expErr: `ParseMessage: ParseField: parseValue: invalid field value '\n'`,
	}, {
		in: "testdata/rfc5322-A.6.3.txt",
		exp: "from:John Doe <jdoe@machine(comment). example>\r\n" +
			"to:Mary Smith <mary@example.net>\r\n" +
			"subject:Saying Hello\r\n" +
			"date:Fri, 21 Nov 1997 09(comment): 55 : 06 -0600\r\n" +
			"message-id:<1234 @ local(blah) .machine .example>\r\n" +
			"\r\n" +
			"\r\n" +
			"This is a message just to say hello.\r\n" +
			"So, \"Hello\".\r\n",
	}, {
		in: "testdata/multipart-mixed.txt",
		exp: "from:Nathaniel Borenstein <nsb@bellcore.com>\r\n" +
			"to:Ned Freed <ned@innosoft.com>\r\n" +
			"date:Sun, 21 Mar 1993 23:56:48 -0800 (PST)\r\n" +
			"subject:Sample message\r\n" +
			"mime-version:1.0\r\n" +
			"content-type:multipart/mixed; boundary=\"simple boundary\"\r\n" +
			"\r\n" +
			"\r\n" +
			"This is implicitly typed plain US-ASCII text.\r\n" +
			"It does NOT end with a linebreak.\r\n" +
			"content-type:text/plain; charset=us-ascii\r\n" +
			"\r\n" +
			"This is explicitly typed plain US-ASCII text.\r\n" +
			"It DOES end with a linebreak.\r\n" +
			"\r\n",
		expRest: "\r\nThis is the epilogue.  It is also to be ignored.\r\n\r\n",
	}}

	for _, c := range cases {
		t.Log(c.in)

		in, err := os.ReadFile(c.in)
		if err != nil {
			t.Fatal(err)
		}

		msg, rest, err := ParseMessage(in)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}
		if msg == nil {
			continue
		}

		test.Assert(t, "rest", c.expRest, string(rest))
		test.Assert(t, "Message", c.exp, msg.String())
	}
}

func TestMessage_AddCC(t *testing.T) {
	var (
		msg Message
		err error
	)

	cases := []struct {
		desc      string
		mailboxes string
		expMsg    string
		expError  string
	}{{
		desc:      "One mailbox",
		mailboxes: "one <a@b.c>",
		expMsg:    "cc:one <a@b.c>\r\n\r\n",
	}, {
		desc:   "Empty mailbox",
		expMsg: "cc:one <a@b.c>\r\n\r\n",
	}, {
		desc:      "Invalid mailbox",
		mailboxes: "a",
		expError:  `AddCC: ParseMailboxes: empty or invalid address`,
		expMsg:    "cc:one <a@b.c>\r\n\r\n",
	}, {
		desc:      "Multiple mailboxes",
		mailboxes: "two <a@b.c>,   three <a@b.c> ",
		expMsg:    "cc:one <a@b.c>, two <a@b.c>, three <a@b.c>\r\n\r\n",
	}}

	for _, c := range cases {
		err = msg.AddCC(c.mailboxes)
		if err != nil {
			test.Assert(t, c.desc, c.expError, err.Error())
		}
		test.Assert(t, c.desc, c.expMsg, msg.String())
	}
}

func TestMessage_AddTo(t *testing.T) {
	var (
		msg Message
		err error
	)

	cases := []struct {
		desc      string
		mailboxes string
		expMsg    string
		expError  string
	}{{
		desc:      "One mailbox",
		mailboxes: "one <a@b.c>",
		expMsg:    "to:one <a@b.c>\r\n\r\n",
	}, {
		desc:   "Empty mailbox",
		expMsg: "to:one <a@b.c>\r\n\r\n",
	}, {
		desc:      "Invalid mailbox",
		mailboxes: "a",
		expError:  `AddTo: ParseMailboxes: empty or invalid address`,
		expMsg:    "to:one <a@b.c>\r\n\r\n",
	}, {
		desc:      "Multiple mailboxes",
		mailboxes: "two <a@b.c>,   three <a@b.c> ",
		expMsg:    "to:one <a@b.c>, two <a@b.c>, three <a@b.c>\r\n\r\n",
	}}

	for _, c := range cases {
		err = msg.AddTo(c.mailboxes)
		if err != nil {
			test.Assert(t, c.desc, c.expError, err.Error())
		}
		test.Assert(t, c.desc, c.expMsg, msg.String())
	}
}

// NOTE: this test require call to DNS to get the public key.
func TestMessageDKIMVerify(t *testing.T) {
	t.Skip("TODO: use local DNS")

	cases := []struct {
		expStatus *dkim.Status
		inFile    string
		expErr    string
	}{{
		inFile: "testdata/message-dkimverify-00.txt",
		expStatus: &dkim.Status{
			Type: dkim.StatusOK,
			SDID: []byte("googlegroups.com"),
		},
	}, {
		inFile: "testdata/message-dkimverify-01.txt",
		expStatus: &dkim.Status{
			Type: dkim.StatusOK,
			SDID: []byte("mg.papercall.io"),
		},
	}}

	for _, c := range cases {
		t.Log(c.inFile)

		msg, _, err := ParseFile(c.inFile)
		if err != nil {
			t.Fatal(err)
		}

		gotStatus, err := msg.DKIMVerify()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "dkim.Status", c.expStatus, gotStatus)
	}
}

func TestMessageDKIMSign(t *testing.T) {
	if privateKey == nil || publicKey == nil {
		initKeys(t)
	}

	canonSimple := dkim.CanonSimple

	cases := []struct {
		sig       *dkim.Signature
		expStatus *dkim.Status

		inFile       string
		expBodyHash  string
		expSignature string
	}{{
		inFile: "testdata/message-dkimsign-00.txt",
		sig: &dkim.Signature{
			SDID:        []byte("example.com"),
			Selector:    []byte("brisbane"),
			CanonHeader: &canonSimple,
			CanonBody:   &canonSimple,
			AUID:        []byte("joe@football.example.com"),
			QMethod:     &dkim.QueryMethod{},
		},
		expBodyHash:  "2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=",
		expSignature: "r4xRAHbEEmL8BwGSZkYzCmDT2Y6ttIEc8boo0UZSENC0unBX4JjjaGALuBjlUiTw6t78PeMx3kgIoX3sjkcquw4TvZgfJNKPEDhTq11IU+2QPJSQa245Tjs3eMZCq/cooax4vEPiJIN9UDNT1BNqbF7cMPGjn5RQQtjbHXxRHjI=",
		expStatus: &dkim.Status{
			Type: dkim.StatusOK,
			SDID: []byte("example.com"),
		},
	}}

	for _, c := range cases {
		t.Log(c.inFile)

		msg, _, err := ParseFile(c.inFile)
		if err != nil {
			t.Fatal(err)
		}

		err = msg.DKIMSign(privateKey, c.sig)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "BodyHash", c.expBodyHash, string(msg.DKIMSignature.BodyHash))
		test.Assert(t, "Signature", c.expSignature, string(msg.DKIMSignature.Value))

		gotStatus, err := msg.DKIMVerify()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "dkim.Status", c.expStatus, gotStatus)
	}
}

func TestMessage_packSingle(t *testing.T) {
	dateInUtc = true

	var (
		tdata *test.Data
		msg   Message
		err   error
		exp   string
		got   []byte
	)

	type testCase struct {
		bodyText  string
		bodyHTML  string
		outputTag string
	}

	tdata, err = test.LoadData(`testdata/message_packsingle_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var cases = []testCase{{
		bodyText:  `this is a body text`,
		outputTag: `body.txt`,
	}, {
		bodyHTML:  `<p>this is an HTML body</p>`,
		outputTag: `body.html`,
	}}

	for _, c := range cases {
		msg.Body.Parts = nil

		if len(c.bodyText) > 0 {
			err = msg.SetBodyText([]byte(c.bodyText))
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(c.bodyHTML) > 0 {
			err = msg.SetBodyHtml([]byte(c.bodyHTML))
			if err != nil {
				t.Fatal(err)
			}
		}

		got, err = msg.Pack()
		if err != nil {
			t.Fatal(err)
		}

		exp = string(tdata.Output[c.outputTag])
		exp = fmt.Sprintf(exp, msg.Header.ID())
		test.Assert(t, c.outputTag, exp, string(got))
	}
}

func TestMessage_SetBodyText(t *testing.T) {
	dateInUtc = true

	var (
		tdata *test.Data
		msg   Message
		err   error
		exp   string
		got   []byte
	)

	tdata, err = test.LoadData(`testdata/message_setbodytext_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		body      string
		outputTag string
	}

	var cases = []testCase{{
		body:      `text body`,
		outputTag: `body1.txt`,
	}, {
		body:      `new text body`,
		outputTag: `body2.txt`,
	}}

	for _, c := range cases {
		err = msg.SetBodyText([]byte(c.body))
		if err != nil {
			t.Fatal(err)
		}

		got, err = msg.Pack()
		if err != nil {
			t.Fatal(err)
		}

		exp = string(tdata.Output[c.outputTag])
		exp = fmt.Sprintf(exp, msg.Header.ID())
		test.Assert(t, c.outputTag, exp, string(got))
	}
}

func TestMessage_SetCC(t *testing.T) {
	var (
		msg Message
		err error
	)

	cases := []struct {
		desc      string
		mailboxes string
		expMsg    string
		expError  string
	}{{
		desc:      "One mailbox",
		mailboxes: "test <a@b.c>",
		expMsg:    "cc:test <a@b.c>\r\n\r\n",
	}, {
		desc:   "Empty mailbox",
		expMsg: "cc:test <a@b.c>\r\n\r\n",
	}, {
		desc:      "Invalid mailbox",
		mailboxes: "a",
		expError:  `SetCC: Set: ParseMailboxes: empty or invalid address`,
		expMsg:    "cc:test <a@b.c>\r\n\r\n",
	}, {
		desc:      "Multiple mailboxes",
		mailboxes: "new <a@b.c>, from <a@b.c>",
		expMsg:    "cc:new <a@b.c>, from <a@b.c>\r\n\r\n",
	}}

	for _, c := range cases {
		err = msg.SetCC(c.mailboxes)
		if err != nil {
			test.Assert(t, c.desc, c.expError, err.Error())
		}
		test.Assert(t, c.desc, c.expMsg, msg.String())
	}
}

func TestMessage_SetFrom(t *testing.T) {
	var (
		msg Message
		err error
	)

	cases := []struct {
		desc     string
		mailbox  string
		expMsg   string
		expError string
	}{{
		desc:    "Valid mailbox",
		mailbox: "test <a@b.c>",
		expMsg:  "from:test <a@b.c>\r\n\r\n",
	}, {
		desc:   "Empty mailbox",
		expMsg: "from:test <a@b.c>\r\n\r\n",
	}, {
		desc:     "Invalid mailbox",
		mailbox:  "a",
		expError: `SetFrom: Set: ParseMailboxes: empty or invalid address`,
		expMsg:   "from:test <a@b.c>\r\n\r\n",
	}, {
		desc:    "New mailbox",
		mailbox: "new <a@b.c>",
		expMsg:  "from:new <a@b.c>\r\n\r\n",
	}, {
		desc:    "Multiple mailboxes",
		mailbox: "two <a@b.c>, three <a@b.c>",
		expMsg:  "from:two <a@b.c>, three <a@b.c>\r\n\r\n",
	}}

	for _, c := range cases {
		err = msg.SetFrom(c.mailbox)
		if err != nil {
			test.Assert(t, c.desc, c.expError, err.Error())
		}
		test.Assert(t, c.desc, c.expMsg, msg.String())
	}
}

func TestMessage_SetSubject(t *testing.T) {
	var (
		msg Message
	)
	cases := []struct {
		subject string
		expMsg  string
	}{{
		subject: "a subject",
		expMsg:  "subject:a subject\r\n\r\n",
	}, {
		expMsg: "subject:a subject\r\n\r\n",
	}, {
		subject: "new subject",
		expMsg:  "subject:new subject\r\n\r\n",
	}}

	for _, c := range cases {
		msg.SetSubject(c.subject)

		test.Assert(t, "SetSubject", c.expMsg, msg.String())
	}
}

func TestMessage_SetTo(t *testing.T) {
	var (
		msg Message
		err error
	)

	cases := []struct {
		desc      string
		mailboxes string
		expMsg    string
		expError  string
	}{{
		desc:      "One mailbox",
		mailboxes: "test <a@b.c>",
		expMsg:    "to:test <a@b.c>\r\n\r\n",
	}, {
		desc:   "Empty mailbox",
		expMsg: "to:test <a@b.c>\r\n\r\n",
	}, {
		desc:      "Invalid mailbox",
		mailboxes: "a",
		expMsg:    "to:test <a@b.c>\r\n\r\n",
		expError:  `SetTo: Set: ParseMailboxes: empty or invalid address`,
	}, {
		desc:      "Multiple mailboxes",
		mailboxes: "new <a@b.c>, from <a@b.c>",
		expMsg:    "to:new <a@b.c>, from <a@b.c>\r\n\r\n",
	}}

	for _, c := range cases {
		err = msg.SetTo(c.mailboxes)
		if err != nil {
			test.Assert(t, c.desc, c.expError, err.Error())
		}
		test.Assert(t, c.desc, c.expMsg, msg.String())
	}
}
