// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"sync"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ints"
)

// ClientManager manage list of active websocket connections on server.
//
// This library assume that each connection belong to a user in the server,
// where each user is representated by uint64.
//
// For a custom management of user use HandleClientAdd and HandleClientRemove
// on Server.
type ClientManager struct {
	// conns contains a one-to-many mapping between user ID and their
	// connections.
	conns map[uint64][]int

	// ctx contains a one-to-one mapping between a socket and its context.
	// The context value is a result from successful authentication,
	// HandleAuth on Server.
	ctx map[int]context.Context

	// frame contains a one-to-one mapping between a socket and a frame.
	// Its usually used to handle chopped frame.
	frame map[int]*Frame

	// frames contains a one-to-one mapping between a socket and
	// continuous frame.
	frames map[int]*Frames

	// all connections.
	all []int

	sync.Mutex
}

// newClientManager create and initialize new user sockets.
func newClientManager() *ClientManager {
	return &ClientManager{
		conns:  make(map[uint64][]int),
		ctx:    make(map[int]context.Context),
		frame:  make(map[int]*Frame),
		frames: make(map[int]*Frames),
	}
}

// All return a copy of all client connections.
func (cls *ClientManager) All() (conns []int) {
	cls.Lock()
	defer cls.Unlock()

	if len(cls.all) > 0 {
		conns = make([]int, len(cls.all))
		copy(conns, cls.all)
	}
	return
}

// finFrames merge all continuous frames into single frame and clear the
// stored frame and frames on behalf of connection.
func (cls *ClientManager) finFrames(conn int, last *Frame) (f *Frame) {
	cls.Lock()
	defer cls.Unlock()

	var (
		frames *Frames
		ok     bool
	)

	frames, ok = cls.frames[conn]
	if !ok {
		return last
	}

	f = frames.fin(last)

	delete(cls.frames, conn)
	delete(cls.frame, conn)

	return
}

// Conns return list of connections by user ID.
//
// Each user may have more than one connection (e.g. from Android, iOS, or
// web). By knowing which connections that user have, implementor of websocket
// server can broadcast a message to all connections.
func (cls *ClientManager) Conns(uid uint64) (conns []int) {
	cls.Lock()
	defer cls.Unlock()

	conns = cls.conns[uid]
	return
}

// Context return the client context.
func (cls *ClientManager) Context(conn int) (ctx context.Context, ok bool) {
	cls.Lock()
	defer cls.Unlock()

	ctx, ok = cls.ctx[conn]
	return ctx, ok
}

// getFrame return an active frame on a client connection.
func (cls *ClientManager) getFrame(conn int) (frame *Frame, ok bool) {
	cls.Lock()
	defer cls.Unlock()

	frame, ok = cls.frame[conn]
	return
}

// getFrames return continuous frames on behalf of connection.
func (cls *ClientManager) getFrames(conn int) (frames *Frames, ok bool) {
	cls.Lock()
	defer cls.Unlock()

	frames, ok = cls.frames[conn]
	return
}

// setFrame set the active, chopped frame on client connection.  If frame is
// nil, it will delete the stored frame in connection.
func (cls *ClientManager) setFrame(conn int, frame *Frame) {
	cls.Lock()
	defer cls.Unlock()

	if frame == nil {
		delete(cls.frame, conn)
	} else {
		cls.frame[conn] = frame
	}
}

// setFrames set continuous frames on client connection.  If frames is nil it
// will clear the stored frames.
func (cls *ClientManager) setFrames(conn int, frames *Frames) {
	cls.Lock()
	defer cls.Unlock()

	if frames == nil {
		delete(cls.frames, conn)
	} else {
		cls.frames[conn] = frames
	}
}

// add new socket connection to user ID in context.
func (cls *ClientManager) add(ctx context.Context, conn int) {
	var (
		v     interface{}
		uid   uint64
		conns []int
		ok    bool
	)

	// Delete the previous socket reference on other user ID.
	cls.remove(conn)

	cls.Lock()
	defer cls.Unlock()

	if !ints.IsExist(cls.all, conn) {
		cls.all = append(cls.all, conn)
	}

	cls.ctx[conn] = ctx

	v = ctx.Value(CtxKeyUID)
	if v != nil {
		uid, _ = v.(uint64)
	}
	if uid > 0 {
		conns, ok = cls.conns[uid]
		if !ok {
			conns = make([]int, 0, 1)
		}
		conns = append(conns, conn)

		cls.conns[uid] = conns
	}
}

// remove a connection from list of clients.
func (cls *ClientManager) remove(conn int) {
	var (
		ctx   context.Context
		v     interface{}
		conns []int
		ok    bool
	)

	cls.Lock()
	defer cls.Unlock()

	delete(cls.frame, conn)
	delete(cls.frames, conn)
	cls.all, _ = ints.Remove(cls.all, conn)

	ctx, ok = cls.ctx[conn]
	if ok {
		var uid uint64
		v = ctx.Value(CtxKeyUID)
		if v != nil {
			uid, _ = v.(uint64)
		}
		delete(cls.ctx, conn)

		if uid > 0 {
			conns, ok = cls.conns[uid]
			if ok {
				conns, _ = ints.Remove(conns, conn)

				if len(conns) == 0 {
					delete(cls.conns, uid)
				} else {
					cls.conns[uid] = conns
				}
			}
		}
	}
}
