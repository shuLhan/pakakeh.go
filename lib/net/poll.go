package net

const (
	maxQueue = 128
)

//
// Poll represent an interface to network polling.
//
type Poll interface {
	//
	// Close the poll.
	//
	Close()

	//
	// RegisterRead add the file descriptor to read poll.
	//
	RegisterRead(fd int) (err error)

	//
	// ReregisterRead register the file descriptor back to events.
	// This method must be used on Linux after calling WaitRead.
	//
	// See https://idea.popcount.org/2017-02-20-epoll-is-fundamentally-broken-12/
	//
	ReregisterRead(idx, fd int)

	//
	// UnregisterRead remove file descriptor from the poll.
	//
	UnregisterRead(fd int) (err error)

	//
	// WaitRead wait for read event received and return list of file
	// descriptor that are ready for reading.
	//
	WaitRead() (fds []int, err error)
}
