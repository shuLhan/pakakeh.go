// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestMailTx_isTerminated(t *testing.T) {
	cases := []struct {
		desc string
		mail MailTx
		exp  bool
	}{{
		desc: "With length < 5",
		mail: MailTx{
			Data: []byte("\n.\r\n"),
		},
	}, {
		desc: "With empty data",
		mail: MailTx{
			Data: []byte("\r\n.\r\n"),
		},
		exp: true,
	}, {
		desc: "With data",
		mail: MailTx{
			Data: []byte("Hello, there\r\n.\r\n"),
		},
		exp: true,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		test.Assert(t, "isTerminated", c.exp, c.mail.isTerminated())
	}
}

func TestFormat(t *testing.T) {
	var (
		in  = ".\n..\r\na.text.\n.message\r\n.\r\n"
		exp = "..\r\n...\r\na.text.\r\n..message\r\n..\r\n"
		got []byte
	)
	got = format([]byte(in))
	test.Assert(t, `format`, []byte(exp), got)
}
