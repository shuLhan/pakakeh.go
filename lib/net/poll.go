// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

const (
	// maxQueue define the number of events that can be read from poll at
	// one time.
	// Increasing this number also increase the memory consumed by
	// process.
	maxQueue = 2048
)

// Poll represent an interface to network polling.
type Poll interface {
	// Close the poll.
	Close()

	// RegisterRead add the file descriptor to read poll.
	RegisterRead(fd int) (err error)

	// ReregisterEvent register the event back.
	// This method must be called back after calling WaitReadEvent when
	// PollEvent descriptor still being used.
	ReregisterEvent(pe PollEvent) error

	// ReregisterRead register the file descriptor back to events.
	// This method must be called back after calling WaitRead when fd
	// still being used.
	//
	// See https://idea.popcount.org/2017-02-20-epoll-is-fundamentally-broken-12/
	ReregisterRead(idx, fd int)

	// UnregisterRead remove file descriptor from the poll.
	UnregisterRead(fd int) (err error)

	// WaitRead wait for read event received and return list of file
	// descriptor that are ready for reading.
	WaitRead() (fds []int, err error)

	// WaitReadEvents wait for read event received and return list of
	// PollEvent that contains the file descriptor and the underlying
	// OS specific event state.
	WaitReadEvents() (events []PollEvent, err error)
}
