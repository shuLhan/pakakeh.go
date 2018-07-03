// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var _userSocks = &UserSockets{}

func testUserSocketsAdd(t *testing.T) {
	cases := []struct {
		desc     string
		uid      uint64
		conn     int
		expConns []int
	}{{
		desc:     `With new connection`,
		uid:      1,
		conn:     1,
		expConns: []int{1},
	}, {
		desc:     `With same connection`,
		uid:      1,
		conn:     1,
		expConns: []int{1},
	}, {
		desc:     `With different connection`,
		uid:      1,
		conn:     2,
		expConns: []int{1, 2},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_userSocks.Add(c.uid, c.conn)

		got, _ := _userSocks.Load(c.uid)

		test.Assert(t, "conns", c.expConns, got, true)
	}
}

func testUserSocketsRemove(t *testing.T) {
	cases := []struct {
		desc     string
		uid      uint64
		conn     int
		expOK    bool
		expConns []int
	}{{
		desc:     `With invalid connection`,
		uid:      1,
		conn:     99,
		expOK:    true,
		expConns: []int{1, 2},
	}, {
		desc:     `With valid connection`,
		uid:      1,
		conn:     1,
		expOK:    true,
		expConns: []int{2},
	}, {
		desc: `With valid connection`,
		uid:  1,
		conn: 2,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_userSocks.Remove(c.uid, c.conn)

		v, ok := _userSocks.Load(c.uid)
		if ok {
			got := v.([]int)
			test.Assert(t, "conns", c.expConns, got, true)
		} else {
			test.Assert(t, "ok", c.expOK, ok, true)
		}
	}
}

func TestUserSockets(t *testing.T) {
	t.Run("add", testUserSocketsAdd)
	t.Run("remove", testUserSocketsRemove)
}
