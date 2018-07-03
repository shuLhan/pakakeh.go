// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"sync"
)

//
// UserSockets define a mapping between user ID (uint64) and list of their
// connections ([]int)
//
// Each user may have more than one connection (e.g. from Android, iOS, web,
// etc). By knowing which connections that user have, implementor of websocket
// server can broadcast a message.
//
type UserSockets struct {
	sync.Map
}

//
// Add append new socket `conn` to user ID uid only if the socket is not
// already exist.
//
func (us *UserSockets) Add(uid uint64, conn int) {
	v, ok := us.Load(uid)
	if !ok {
		us.Store(uid, []int{conn})
		return
	}

	conns := v.([]int)

	for x := 0; x < len(conns); x++ {
		if conns[x] == conn {
			return
		}
	}

	conns = append(conns, conn)

	us.Store(uid, conns)
}

//
// Remove socket from list of user's connection.
//
func (us *UserSockets) Remove(uid uint64, conn int) {
	v, ok := us.Load(uid)
	if !ok {
		return
	}

	conns := v.([]int)

	for x := 0; x < len(conns); x++ {
		if conns[x] != conn {
			continue
		}

		conns = append(conns[:x], conns[x+1:]...)
		if len(conns) > 0 {
			us.Store(uid, conns)
		} else {
			us.Delete(uid)
		}
	}
}
