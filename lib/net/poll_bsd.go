// +build darwin dragonfly freebsd netbsd openbsd

package net

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type kqueue struct {
	events [maxQueue]unix.Kevent_t
	read   int
}

//
// NewPoll create and initialize new poll using epoll for Linux system or
// kqueue for BSD or Darwin (macOS).
//
func NewPoll() (Poll, error) {
	var err error

	kq := &kqueue{}

	kq.read, err = unix.Kqueue()
	if err != nil {
		return nil, fmt.Errorf("kqueue.NewPoll: %s", err.Error())
	}

	return kq, nil
}

func (poll *kqueue) Close() {
	// no-op
}

func (poll *kqueue) RegisterRead(fd int) (err error) {
	kevent := unix.Kevent_t{}

	unix.SetKevent(&kevent, fd, unix.EVFILT_READ, unix.EV_ADD)

	err = unix.SetNonblock(fd, true)
	if err != nil {
		return fmt.Errorf("kqueue.RegisterRead: %s", err.Error())
	}

	changes := []unix.Kevent_t{
		kevent,
	}

	_, err = unix.Kevent(poll.read, changes, nil, nil)
	if err != nil {
		return fmt.Errorf("kqueue.RegisterRead: %s", err.Error())
	}

	return nil
}

func (poll *kqueue) ReregisterRead(idx, fd int) {
	// no-op
}

func (poll *kqueue) UnregisterRead(fd int) (err error) {
	kevent := unix.Kevent_t{}

	unix.SetKevent(&kevent, fd, unix.EVFILT_READ, unix.EV_DELETE)

	changes := []unix.Kevent_t{
		kevent,
	}

	_, err = unix.Kevent(poll.read, changes, nil, nil)
	if err != nil {
		return fmt.Errorf("kqueue.UnregisterRead: %s", err.Error())
	}

	return nil
}

func (poll *kqueue) WaitRead() (fds []int, err error) {
	n, err := unix.Kevent(poll.read, nil, poll.events[:], nil)
	if err != nil {
		return nil, fmt.Errorf("kqueue.WaitRead: %s", err.Error())
	}

	for x := 0; x < n; x++ {
		switch poll.events[x].Filter {
		case unix.EVFILT_READ:
			fds = append(fds, int(poll.events[x].Ident))
		}
	}

	return fds, nil
}
