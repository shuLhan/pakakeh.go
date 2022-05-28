// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestServer_CachesSave(t *testing.T) {
	var (
		srv = &Server{}

		hname   = []byte("caches.save.local")
		address = []byte("127.0.0.1")
		msg     = NewMessageAddress(hname, [][]byte{address})
		answer  = newAnswer(msg, false)

		expAnswers []*Answer
		gotAnswers []*Answer
		err        error
		n          int
	)

	srv.Caches.init(0, 0)

	_ = srv.Caches.upsert(answer)

	var w bytes.Buffer

	n, err = srv.CachesSave(&w)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "CachesSave", 1, n)

	msg = NewMessage()
	msg.packet = answer.msg.packet
	err = msg.Unpack()
	if err != nil {
		t.Fatal(err)
	}
	expAnswers = append(expAnswers, newAnswer(msg, false))

	srv.Caches.init(0, 0)

	gotAnswers, err = srv.CachesLoad(&w)
	if err != nil {
		t.Fatal(err)
	}

	for _, answer = range gotAnswers {
		answer.el = nil
	}

	test.Assert(t, "CachesWrite", expAnswers, gotAnswers)
}
