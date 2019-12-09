// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	var wg sync.WaitGroup

	f, err := ioutil.TempFile("", "watcher")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("= Watching: %s\n", f.Name())

	exps := []struct {
		state FileState
		mode  os.FileMode
		size  int64
	}{{
		state: FileStateModified,
		mode:  0700,
	}, {
		state: FileStateModified,
		mode:  0700,
		size:  13,
	}, {
		state: FileStateDeleted,
		mode:  0700,
		size:  13,
	}}

	x := 0
	_, err = NewWatcher(f.Name(), 150*time.Millisecond, func(ns *NodeState) {
		if exps[x].state != ns.State {
			log.Fatalf("Got state %d, want %d\n", ns.State, exps[x].state)
		}
		if exps[x].mode != ns.Node.Mode() {
			log.Fatalf("Got mode %d, want %d\n", ns.Node.Mode(), exps[x].mode)
		}
		if exps[x].size != ns.Node.Size() {
			log.Fatalf("Got size %d, want %d\n", ns.Node.Size(), exps[x].size)
		}
		x++
		wg.Done()
	})
	if err != nil {
		t.Fatal(err)
	}

	// Update file mode
	wg.Add(1)
	err = f.Chmod(0700)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	wg.Add(1)
	_, err = f.WriteString("Write changes")
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	err = os.Remove(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
