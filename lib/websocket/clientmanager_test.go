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
	clients := newClientManager()

	ctx1 := context.WithValue(context.Background(), CtxKeyUID, uint64(1))
	ctx2 := context.WithValue(context.Background(), CtxKeyUID, uint64(2))

	cases := []struct {
		desc     string
		ctx      context.Context
		conn     int
		expConns string
		expCtx   string
	}{{
		desc:     `With new connection`,
		ctx:      ctx1,
		conn:     1000,
		expConns: "map[1:[1000]]",
		expCtx:   "map[1000:context.Background.WithValue(0x4, 0x1)]",
	}, {
		desc:     `With same connection`,
		ctx:      ctx1,
		conn:     1000,
		expConns: "map[1:[1000]]",
		expCtx:   "map[1000:context.Background.WithValue(0x4, 0x1)]",
	}, {
		desc:     `With different connection`,
		ctx:      ctx1,
		conn:     2000,
		expConns: "map[1:[1000 2000]]",
		expCtx:   "map[1000:context.Background.WithValue(0x4, 0x1) 2000:context.Background.WithValue(0x4, 0x1)]",
	}, {
		desc:     "With same connection different UID",
		ctx:      ctx2,
		conn:     1000,
		expConns: "map[1:[2000] 2:[1000]]",
		expCtx:   "map[1000:context.Background.WithValue(0x4, 0x2) 2000:context.Background.WithValue(0x4, 0x1)]",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		clients.add(c.ctx, c.conn)

		gotConns := fmt.Sprintf("%v", clients.conns)
		gotCtx := fmt.Sprintf("%v", clients.ctx)

		test.Assert(t, "ClientManager.conns", c.expConns, gotConns, true)
		test.Assert(t, "ClientManager.ctx", c.expCtx, gotCtx, true)
	}
}

func TestClientManagerRemove(t *testing.T) {
	clients := newClientManager()

	ctx1 := context.WithValue(context.Background(), CtxKeyUID, uint64(1))
	ctx2 := context.WithValue(context.Background(), CtxKeyUID, uint64(2))

	clients.add(ctx1, 1000)
	clients.add(ctx1, 2000)
	clients.add(ctx2, 1000)

	cases := []struct {
		desc     string
		conn     int
		expConns string
		expCtx   string
	}{{
		desc:     `With invalid connection`,
		conn:     99,
		expConns: "map[1:[2000] 2:[1000]]",
		expCtx:   "map[1000:context.Background.WithValue(0x4, 0x2) 2000:context.Background.WithValue(0x4, 0x1)]",
	}, {
		desc:     `With valid connection`,
		conn:     1000,
		expConns: "map[1:[2000]]",
		expCtx:   "map[2000:context.Background.WithValue(0x4, 0x1)]",
	}, {
		desc:     `With valid connection`,
		conn:     2000,
		expConns: "map[]",
		expCtx:   "map[]",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		clients.remove(c.conn)

		gotConns := fmt.Sprintf("%v", clients.conns)
		gotCtx := fmt.Sprintf("%v", clients.ctx)

		test.Assert(t, "conns", c.expConns, gotConns, true)
		test.Assert(t, "ClientManager.ctx", c.expCtx, gotCtx, true)
	}
}
