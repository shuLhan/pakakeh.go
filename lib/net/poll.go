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

	// UnregisterRead remove file descriptor from the poll.
	UnregisterRead(fd int) (err error)

	// WaitRead wait and return list of file descriptor (fd) that are
	// ready for reading from the pool.
	// The returned fd is detached from poll to allow concurrent
	// processing of fd at the same time.
	// Once the data has been read from the fd and its still need to be
	// used, one need to put it back to poll using RegisterRead.
	WaitRead() (fds []int, err error)

	// WaitReadEvents wait and return list of PollEvent that contains the
	// file descriptor and the underlying OS specific event state.
	WaitReadEvents() (events []PollEvent, err error)
}
