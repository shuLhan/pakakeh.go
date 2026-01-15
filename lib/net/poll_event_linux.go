// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

//go:build linux

package net

import "golang.org/x/sys/unix"

type pollEvent struct {
	fd    int32
	event unix.EpollEvent
}

func (pe *pollEvent) Descriptor() uint64 {
	return uint64(pe.fd)
}

func (pe *pollEvent) Event() any {
	return pe.event
}
