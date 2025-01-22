// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package net

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

type kqueue struct {
	events [maxQueue]unix.Kevent_t
	read   int
}

// NewPoll create and initialize new poll using epoll for Linux system or
// kqueue for BSD or Darwin (macOS).
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
	var (
		logp = `WaitRead`

		n  int
		fd int
	)
	for n == 0 {
		n, err = unix.Kevent(poll.read, nil, poll.events[:], nil)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return nil, fmt.Errorf("kqueue.WaitRead: %s", err.Error())
		}
	}

	for x := range n {
		if poll.events[x].Filter != unix.EVFILT_READ {
			continue
		}

		fd = int(poll.events[x].Ident)

		err = poll.UnregisterRead(fd)
		if err != nil {
			log.Printf(`%s: %s`, logp, err)
			continue
		}

		err = unix.SetNonblock(fd, false)
		if err != nil {
			log.Printf(`%s: %s`, logp, err)
			continue
		}

		fds = append(fds, fd)
	}

	return fds, nil
}

func (poll *kqueue) WaitReadEvents() (events []PollEvent, err error) {
	var (
		logp = `WaitReadEvents`

		n  int
		fd int
	)

	for n == 0 {
		n, err = unix.Kevent(poll.read, nil, poll.events[:], nil)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	for x := range n {
		if poll.events[x].Filter != unix.EVFILT_READ {
			continue
		}

		fd = int(poll.events[x].Ident)

		err = poll.UnregisterRead(fd)
		if err != nil {
			log.Printf(`%s: %s`, logp, err)
			continue
		}

		err = unix.SetNonblock(fd, false)
		if err != nil {
			log.Printf(`%s: %s`, logp, err)
			continue
		}

		var event = &pollEvent{
			fd:    poll.events[x].Ident,
			event: poll.events[x],
		}
		events = append(events, event)
	}

	return events, nil
}
