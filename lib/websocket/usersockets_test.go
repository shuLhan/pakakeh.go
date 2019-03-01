// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestUserSocketsAdd(t *testing.T) {
	userSocks := &UserSockets{}

	cases := []struct {
		desc     string
		uid      uint64
		conn     int
		expConns []int
		expUIDs  string
	}{{
		desc:     `With new connection`,
		uid:      1,
		conn:     1,
		expConns: []int{1},
		expUIDs:  "map[1:1]",
	}, {
		desc:     `With same connection`,
		uid:      1,
		conn:     1,
		expConns: []int{1},
		expUIDs:  "map[1:1]",
	}, {
		desc:     `With different connection`,
		uid:      1,
		conn:     2,
		expConns: []int{1, 2},
		expUIDs:  "map[1:1 2:1]",
	}, {
		desc:     "With same connection different UID",
		uid:      2,
		conn:     1,
		expConns: []int{1},
		expUIDs:  "map[1:2 2:1]",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		userSocks.Add(c.uid, c.conn)

		got, _ := userSocks.Load(c.uid)
		gotUIDs := fmt.Sprintf("%v", userSocks.uid)

		test.Assert(t, "conns", c.expConns, got, true)
		test.Assert(t, "UserSockets.uid", c.expUIDs, gotUIDs, true)
	}
}

func TestUserSocketsRemove(t *testing.T) {
	userSocks := &UserSockets{}
	userSocks.Add(1, 1)
	userSocks.Add(1, 2)
	userSocks.Add(2, 1)

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
		expConns: []int{2},
	}, {
		desc:     `With invalid connection (2)`,
		uid:      1,
		conn:     1,
		expOK:    false,
		expConns: []int{2},
	}, {
		desc: `With valid connection`,
		uid:  1,
		conn: 2,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		userSocks.Remove(c.uid, c.conn)

		v, ok := userSocks.Load(c.uid)
		if ok {
			got := v.([]int)
			test.Assert(t, "conns", c.expConns, got, true)
		} else {
			test.Assert(t, "ok", c.expOK, ok, true)
		}
	}
}
