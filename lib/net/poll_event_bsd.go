// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package net

import "golang.org/x/sys/unix"

type pollEvent struct {
	fd    uint64
	event unix.Kevent_t
}

func (pe *pollEvent) Descriptor() uint64 {
	return pe.fd
}

func (pe *pollEvent) Event() any {
	return pe.event
}
