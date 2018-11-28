// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"os"
	"time"
)

//
// A Watcher hold a channel that deliver file information when file is
// changed.
// If file is deleted, it will send a nil file information to channel and stop
// watching the file.
//
// This is a naive implementation of file event change notification.
//
type Watcher struct {
	C      <-chan *os.FileInfo
	cin    chan *os.FileInfo
	file   string
	ticker *time.Ticker
}

//
// NewWatcher return a new file watcher that will inspect the file for changes
// with period specified by duration `d` argument.
// If duration is less or equal to zero, it will be set to default duration (5
// seconds).
//
func NewWatcher(file string, d time.Duration) (*Watcher, error) {
	if d <= 0 {
		d = time.Second * 5
	}

	c := make(chan *os.FileInfo, 1)
	watcher := &Watcher{
		C:      c,
		cin:    c,
		file:   file,
		ticker: time.NewTicker(d),
	}

	go watcher.start()

	return watcher, nil
}

func (w *Watcher) start() {
	oldStat, _ := os.Stat(w.file)
	for range w.ticker.C {
		newStat, err := os.Stat(w.file)
		if err != nil {
			w.cin <- nil
			continue
		}
		if oldStat == nil {
			w.cin <- &newStat
			oldStat = newStat
			continue
		}
		if oldStat.Size() != newStat.Size() ||
			oldStat.Mode() != newStat.Mode() ||
			oldStat.ModTime() != newStat.ModTime() {
			w.cin <- &newStat
			oldStat = newStat
			continue
		}
	}
}

//
// Stop watching the file.
//
func (w *Watcher) Stop() {
	w.ticker.Stop()
}
