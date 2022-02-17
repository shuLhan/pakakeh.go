// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestStore(t *testing.T) {
	testFS, err := NewLocalStorage("./testdata")
	if err != nil {
		t.Fatalf("NewLocalStorage: %s\n", err)
	}

	cases := []struct {
		expErr error
		mail   *MailTx
		desc   string
	}{{
		desc: "With nil",
	}, {
		desc: "Without received",
		mail: NewMailTx("me", []string{"a@box.com"}, []byte("Hello, world!")),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := testFS.MailSave(c.mail)
		if err != nil {
			test.Assert(t, "error", c.expErr, err)
			continue
		}

		var got *MailTx
		if c.mail != nil {
			got, err = testFS.MailLoad(c.mail.ID)
			if err != nil {
				t.Fatal(err)
			}
		}

		test.Assert(t, "mail", c.mail, got)
	}
}

func TestDelete(t *testing.T) {
	testFS, err := NewLocalStorage("./testdata")
	if err != nil {
		t.Fatalf("NewLocalStorage: %s\n", err)
	}

	mails, err := testFS.MailLoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %s\n", err)
	}

	for _, mail := range mails {
		err = testFS.MailDelete(mail.ID)
		if err != nil {
			t.Fatalf("LocalStorage.Delete: %s\n", err)
		}
	}
}
