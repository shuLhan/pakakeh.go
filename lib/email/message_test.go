// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"io/ioutil"
	"testing"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/email/dkim"
	"github.com/shuLhan/share/lib/test"
)

func TestNewMultipart(t *testing.T) {
	if debug.Value == 0 {
		// This test involve random generated string on boundary, so
		// it should be run manually.
		t.Skip()
	}

	cases := []struct {
		from, to, subject []byte
		bodyText          []byte
		bodyHTML          []byte
	}{{
		from:     []byte("a@b.c"),
		to:       []byte("d@e.f"),
		subject:  []byte("test"),
		bodyText: []byte("This is plain text"),
		bodyHTML: []byte("<b>This is body in HTML</b>"),
	}}

	for _, c := range cases {
		got, err := NewMultipart(c.from, c.to, c.subject, c.bodyText, c.bodyHTML)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("NewMultipart:\n%s", got.Pack())
	}
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
		expErr: "ParseMessage: email: invalid field value at 'From  : John Doe <jdoe@machine(comment).  example>'",
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

		in, err := ioutil.ReadFile(c.in)
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

//
// NOTE: this test require call to DNS to get the public key.
//
func TestMessageDKIMVerify(t *testing.T) {
	t.Skip()

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
