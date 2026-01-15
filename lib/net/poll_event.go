// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package net

// PollEvent define an interface for poll event, [unix.EpollEvent] on Linux or
// [unix.Kevent_t] on BSD.
type PollEvent interface {
	// Descriptor return the file descriptor associated with poll.
	Descriptor() uint64

	// Event return the underlying event structure.
	// It can be cast to actual type, unix.EpollEvent in Linux or
	// unix.Kevent_t on BSD.
	Event() any
}
