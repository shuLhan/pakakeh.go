// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux
// +build linux

package net

import (
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

func (poll *epoll) ReregisterEvent(event PollEvent) (err error) {
	var (
		logp = `ReregisterEvent`
		fd   = int(event.Descriptor())
		obj  = event.Event()

		epollEvent unix.EpollEvent
		ok         bool
	)

	epollEvent, ok = obj.(unix.EpollEvent)
	if !ok {
		return fmt.Errorf(`%s: expecting unix.EpollEvent, got %T`, logp, obj)
	}

	epollEvent.Events = unix.EPOLLIN | unix.EPOLLONESHOT

	err = unix.SetNonblock(fd, true)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = unix.EpollCtl(poll.read, unix.EPOLL_CTL_MOD, fd, &epollEvent)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

func (poll *epoll) ReregisterRead(idx, fd int) {
	var err error

	poll.events[idx].Events = unix.EPOLLIN | unix.EPOLLONESHOT

	err = unix.SetNonblock(fd, true)
	if err != nil {
		log.Printf("epoll.ReregisterRead: %s", err.Error())
	}

	err = unix.EpollCtl(poll.read, unix.EPOLL_CTL_MOD, fd, &poll.events[idx])
	if err != nil {
		log.Println("epoll.RegisterRead: unix.EpollCtl: " + err.Error())
		err = poll.UnregisterRead(fd)
		if err != nil {
			log.Println("epoll.RegisterRead: " + err.Error())
		}
	}
}

func (poll *epoll) UnregisterRead(fd int) (err error) {
	err = unix.EpollCtl(poll.read, unix.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return fmt.Errorf("epoll.UnregisterRead: %s", err.Error())
	}

	return nil
}

func (poll *epoll) WaitRead() (fds []int, err error) {
	var n int
	for {
		n, err = unix.EpollWait(poll.read, poll.events[:], -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return nil, fmt.Errorf("epoll.WaitRead: %s", err.Error())
		}
		break
	}

	for x := 0; x < n; x++ {
		fds = append(fds, int(poll.events[x].Fd))
	}

	return fds, nil
}

func (poll *epoll) WaitReadEvents() (events []PollEvent, err error) {
	var (
		logp = `WaitReadEvents`

		n int
		x int
	)

	for n == 0 {
		n, err = unix.EpollWait(poll.read, poll.events[:], -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		break
	}

	for x = 0; x < n; x++ {
		events = append(events, &pollEvent{
			fd:    poll.events[x].Fd,
			event: poll.events[x],
		})
	}

	return events, nil
}
