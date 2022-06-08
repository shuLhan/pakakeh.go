// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestClientManagerAdd(t *testing.T) {
	type testCase struct {
		ctx       context.Context
		desc      string
		expConns  string
		conn      int
		expCtxLen int
	}

	var (
		clients = newClientManager()
		ctx1    = context.WithValue(context.Background(), CtxKeyUID, uint64(1))
		ctx2    = context.WithValue(context.Background(), CtxKeyUID, uint64(2))
	)

	var cases = []testCase{{
		desc:      `With new connection`,
		ctx:       ctx1,
		conn:      1000,
		expConns:  "map[1:[1000]]",
		expCtxLen: 1,
	}, {
		desc:      `With same connection`,
		ctx:       ctx1,
		conn:      1000,
		expConns:  "map[1:[1000]]",
		expCtxLen: 1,
	}, {
		desc:      `With different connection`,
		ctx:       ctx1,
		conn:      2000,
		expConns:  "map[1:[1000 2000]]",
		expCtxLen: 2,
	}, {
		desc:      "With same connection different UID",
		ctx:       ctx2,
		conn:      1000,
		expConns:  "map[1:[2000] 2:[1000]]",
		expCtxLen: 2,
	}}

	var (
		c         testCase
		gotConns  string
		gotCtxLen int
	)

	for _, c = range cases {
		t.Log(c.desc)

		clients.add(c.ctx, c.conn)

		gotConns = fmt.Sprintf("%v", clients.conns)
		gotCtxLen = len(clients.ctx)

		test.Assert(t, "ClientManager.conns", c.expConns, gotConns)
		test.Assert(t, "ClientManager.ctx", c.expCtxLen, gotCtxLen)
	}
}

func TestClientManagerRemove(t *testing.T) {
	type testCase struct {
		desc      string
		expConns  string
		expCtxLen int
		conn      int
	}

	var (
		clients = newClientManager()
		ctx1    = context.WithValue(context.Background(), CtxKeyUID, uint64(1))
		ctx2    = context.WithValue(context.Background(), CtxKeyUID, uint64(2))
	)

	clients.add(ctx1, 1000)
	clients.add(ctx1, 2000)
	clients.add(ctx2, 1000)

	var cases = []testCase{{
		desc:      `With invalid connection`,
		conn:      99,
		expConns:  "map[1:[2000] 2:[1000]]",
		expCtxLen: 2,
	}, {
		desc:      `With valid connection`,
		conn:      1000,
		expConns:  "map[1:[2000]]",
		expCtxLen: 1,
	}, {
		desc:      `With valid connection`,
		conn:      2000,
		expConns:  "map[]",
		expCtxLen: 0,
	}}

	var (
		c         testCase
		gotConns  string
		gotCtxLen int
	)

	for _, c = range cases {
		t.Log(c.desc)

		clients.remove(c.conn)

		gotConns = fmt.Sprintf("%v", clients.conns)
		gotCtxLen = len(clients.ctx)

		test.Assert(t, "conns", c.expConns, gotConns)
		test.Assert(t, "ClientManager.ctx", c.expCtxLen, gotCtxLen)
	}
}
