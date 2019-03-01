// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"sync"

	"github.com/shuLhan/share/lib/ints"
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
	sync.Mutex

	// conns contains a one-to-many mapping between user ID and their
	// connections.
	conns map[uint64][]int

	// uid contains a one-to-one mapping between socket and user ID.
	// This mapping is to prevent the same file descriptor to be added to
	// other user ID.
	uid map[int]uint64
}

//
// NewUserSockets create and initialize new user sockets.
//
func NewUserSockets() *UserSockets {
	return &UserSockets{
		conns: make(map[uint64][]int),
		uid:   make(map[int]uint64),
	}
}

//
// Add new socket connection to user ID only if the socket is not already
// exist.
//
func (us *UserSockets) Add(uid uint64, conn int) {
	us.Lock()
	// Check if socket already exist.
	prevUID, ok := us.uid[conn]
	us.Unlock()

	if ok {
		// Delete the previous socket reference on other user ID.
		us.Remove(prevUID, conn)
	}

	us.Lock()
	us.uid[conn] = uid

	conns, ok := us.conns[uid]
	if !ok {
		us.conns[uid] = []int{conn}
	} else if !ints.IsExist(conns, conn) {
		conns = append(conns, conn)
		us.conns[uid] = conns
	}
	us.Unlock()
}

//
// Remove socket from list of user's connection.
//
func (us *UserSockets) Remove(uid uint64, conn int) {
	us.Lock()

	delete(us.uid, conn)

	conns, ok := us.conns[uid]
	if ok {
		conns, _ = ints.Remove(conns, conn)

		if len(conns) == 0 {
			delete(us.conns, uid)
		} else {
			us.conns[uid] = conns
		}
	}

	us.Unlock()
}
