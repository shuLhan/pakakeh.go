// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maildir

import (
	"log"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const testDir = "testdata"

func lsDir(dir string) (ls []os.FileInfo) {
	d, err := os.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatal(err)
	}

	ls, err = d.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}

	_ = d.Close()

	return ls
}

func TestOutQueue(t *testing.T) {
	mg, err := New(testDir)
	if err != nil {
		t.Fatal(err)
	}

	mg.RemoveAll(mg.dirOut)

	cases := []struct {
		desc     string
		email    []byte
		expNList int
		expErr   string
	}{{
		desc:   "With empty email",
		expErr: "email/maildir: OutQueue: empty email",
	}, {
		desc:     "With valid inputs",
		email:    []byte("From: me@localhost"),
		expNList: 1,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err = mg.OutQueue(c.email)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		ls := lsDir(mg.dirOut)

		test.Assert(t, "n List", c.expNList, len(ls))
	}
}

func TestDeleteOutQueue(t *testing.T) {
	mg, err := New(testDir)
	if err != nil {
		t.Fatal(err)
	}

	listOut := lsDir(mg.dirOut)

	test.Assert(t, "n List", 1, len(listOut))

	cases := []struct {
		desc     string
		fname    string
		expErr   string
		expNList int
	}{{
		desc:     "With empty filename",
		expErr:   "email/maildir: DeleteOutQueue: empty file name",
		expNList: 1,
	}, {
		desc:     "With valid filename",
		fname:    listOut[0].Name(),
		expNList: 0,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := mg.DeleteOutQueue(c.fname)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		ls := lsDir(mg.dirOut)

		test.Assert(t, "n List", c.expNList, len(ls))
	}
}

func TestIncoming(t *testing.T) {
	mg, err := New(testDir)
	if err != nil {
		t.Fatal(err)
	}

	mg.RemoveAll(mg.dirNew)
	mg.RemoveAll(mg.dirTmp)

	cases := []struct {
		desc    string
		email   []byte
		expErr  string
		expNTmp int
		expNNew int
	}{{
		desc:   "With empty email",
		expErr: "email/maildir: Incoming: empty email",
	}, {
		desc:    "With valid parameters",
		email:   []byte("From: me@localhost"),
		expNTmp: 0,
		expNNew: 1,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := mg.Incoming(c.email)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		lsTmp := lsDir(mg.dirTmp)
		lsNew := lsDir(mg.dirNew)

		test.Assert(t, "n list tmp", c.expNTmp, len(lsTmp))
		test.Assert(t, "n list new", c.expNNew, len(lsNew))
	}
}
