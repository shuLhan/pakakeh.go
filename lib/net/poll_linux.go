// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package net

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

type epoll struct {
	events [maxQueue]unix.EpollEvent
	read   int
}

// NewPoll create and initialize new poll using epoll for Linux system or
// kqueue for BSD or Darwin (macOS).
func NewPoll() (Poll, error) {
	var err error

	ep := &epoll{}

	ep.read, err = unix.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("epoll.NewPoll: %s", err.Error())
	}

	return ep, nil
}

func (poll *epoll) Close() {
	unix.Close(poll.read)
}

func (poll *epoll) RegisterRead(fd int) (err error) {
	event := unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLONESHOT,
		Fd:     int32(fd),
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		return fmt.Errorf("epoll.RegisterRead: %s", err.Error())
	}

	err = unix.EpollCtl(poll.read, unix.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		return fmt.Errorf("epoll.RegisterRead: %s", err.Error())
	}

	return nil
}

func (poll *epoll) UnregisterRead(fd int) (err error) {
	err = unix.EpollCtl(poll.read, unix.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return fmt.Errorf("epoll.UnregisterRead: %s", err.Error())
	}

	return nil
}

func (poll *epoll) WaitRead() (fds []int, err error) {
	var (
		logp = `WaitRead`

		n  int
		fd int
	)
	for {
		n, err = unix.EpollWait(poll.read, poll.events[:], -1)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		break
	}

	for x := range n {
		fd = int(poll.events[x].Fd)

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

func (poll *epoll) WaitReadEvents() (events []PollEvent, err error) {
	var (
		logp = `WaitReadEvents`

		n  int
		fd int
	)

	for n == 0 {
		n, err = unix.EpollWait(poll.read, poll.events[:], -1)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		break
	}

	for x := range n {
		fd = int(poll.events[x].Fd)

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
			fd:    poll.events[x].Fd,
			event: poll.events[x],
		}
		events = append(events, event)
	}

	return events, nil
}
