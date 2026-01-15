// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

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
