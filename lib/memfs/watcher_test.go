// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

type caseWatcher struct {
	state FileState
	mode  os.FileMode
	size  int64
}

func TestWatcher(t *testing.T) {
	var (
		content = "Write changes"

		f       *os.File
		watcher *Watcher
		gotNS   NodeState
		err     error
		x       int
	)

	f, err = ioutil.TempFile("", "watcher")
	if err != nil {
		t.Fatal(err)
	}

	exps := []caseWatcher{{
		state: FileStateUpdateMode,
		mode:  0700,
	}, {
		state: FileStateUpdateContent,
		mode:  0700,
		size:  int64(len(content)),
	}, {
		state: FileStateDeleted,
		mode:  0700,
		size:  int64(len(content)),
	}}

	watcher, err = NewWatcher(f.Name(), 150*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Update file mode
	err = f.Chmod(0700)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-watcher.C
	test.Assert(t, "state", exps[x].state, gotNS.State)
	test.Assert(t, "file mode", exps[x].mode, gotNS.Node.Mode())
	test.Assert(t, "file size", exps[x].size, gotNS.Node.Size())
	x++

	_, err = f.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-watcher.C
	test.Assert(t, "state", exps[x].state, gotNS.State)
	test.Assert(t, "file mode", exps[x].mode, gotNS.Node.Mode())
	test.Assert(t, "file size", exps[x].size, gotNS.Node.Size())
	x++

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-watcher.C
	test.Assert(t, "state", exps[x].state, gotNS.State)
	test.Assert(t, "file mode", exps[x].mode, gotNS.Node.Mode())
	test.Assert(t, "file size", exps[x].size, gotNS.Node.Size())
	x++
}
