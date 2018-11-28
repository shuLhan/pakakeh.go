// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	var ok bool
	interval := time.Second * 1
	waitInterval := time.Millisecond * 1500

	f, err := ioutil.TempFile("", "watcher")
	if err != nil {
		t.Fatal(err)
	}

	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	if _debug >= 1 {
		t.Logf("= Watching: %+v\n", fi)
	}

	defer func() {
		if !ok {
			os.Remove(f.Name())
		}
	}()

	watcher, err := NewWatcher(f.Name(), interval)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for fi := range watcher.C {
			if fi == nil {
				if _debug >= 1 {
					t.Log("= Change: file deleted")
				}
				return
			}
			if _debug >= 1 {
				t.Logf("= Change: %+v\n", *fi)
			}
		}
	}()

	// Update file mode
	err = f.Chmod(0700)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(waitInterval)

	_, err = f.WriteString("Write changes")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(waitInterval)

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(waitInterval)

	os.Remove(f.Name())
	time.Sleep(waitInterval)

	ok = true
}
