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
	sync.Mutex
	sync.Map

	// uid contains a one-to-one mapping between socket and user ID.
	// This mapping is to prevent the same file descriptor to be added to
	// other user ID.
	uid map[int]uint64
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
		// Delete the previous reference.
		us.Remove(prevUID, conn)
	}

	if us.uid == nil {
		us.uid = make(map[int]uint64)
	}
	us.uid[conn] = uid

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
	us.Lock()
	delete(us.uid, conn)
	us.Unlock()

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
