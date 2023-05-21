// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maildir

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewManager(t *testing.T) {
	var (
		md       *Manager
		expError string
		err      error
	)

	_, err = NewManager(" \t\n")
	expError = `NewManager: empty base directory`
	test.Assert(t, `With empty directory`, expError, err.Error())

	_, err = NewManager(`/`)
	expError = `NewManager: initDirs: mkdir /cur: permission denied`
	test.Assert(t, `With no permission`, expError, err.Error())

	var baseDir = t.TempDir()

	md, err = NewManager(baseDir)
	if err != nil {
		t.Fatal(err)
	}

	if md.pid == 0 {
		t.Fatal(`want PID, got 0`)
	}
	if len(md.hostname) == 0 {
		t.Fatal(`want hostname, got empty`)
	}
	if md.counter != 0 {
		t.Fatalf(`want zero counter, got %d`, md.counter)
	}

	assertDirExist(t, md.dirCur)
	assertDirExist(t, md.dirNew)
	assertDirExist(t, md.dirTmp)
}

func TestDelete(t *testing.T) {
	var (
		mg       *Manager
		err      error
		expError error
	)

	mg, err = NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Case: empty file name.
	err = mg.Delete(``)
	test.Assert(t, `With file not exist`, expError, err)

	// Case: file not exist.
	err = mg.Delete(`filenotexist`)
	test.Assert(t, `With file not exist`, expError, err)

	// Case: file exist.
	var (
		msg   = []byte(`new message`)
		fnNew string
		fnCur string
	)

	fnNew, err = mg.Incoming(msg)
	if err != nil {
		t.Fatal(err)
	}
	fnCur, _, err = mg.FetchNew(fnNew)
	if err != nil {
		t.Fatal(err)
	}
	err = mg.Delete(fnCur)
	if err != nil {
		t.Fatal(err)
	}

	var pathCur = filepath.Join(mg.dirCur, fnCur)
	_, err = os.Stat(pathCur)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf(`want error %q, got %v`, os.ErrNotExist, err)
	}
}

func TestFetchNew(t *testing.T) {
	var (
		mg     *Manager
		fnCur  string
		msg    []byte
		expMsg []byte
		err    error
	)

	mg, err = NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Case: empty file name.

	fnCur, msg, err = mg.FetchNew(``)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `With empty file name: file name`, ``, fnCur)
	test.Assert(t, `With empty file name: message`, expMsg, msg)

	// Case: file not exist.

	fnCur, msg, err = mg.FetchNew(`filenotexist`)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `With file not exist: file name`, ``, fnCur)
	test.Assert(t, `With file not exist: message`, expMsg, msg)

	// Case: file exist.

	var fnNew string

	msg = []byte(`new message`)
	fnNew, err = mg.Incoming(msg)
	if err != nil {
		t.Fatal(err)
	}

	expMsg = msg
	fnCur, msg, err = mg.FetchNew(fnNew)
	if err != nil {
		t.Fatal(err)
	}

	var expNameCur = `1684640949.M875494_P1000_V36_I170430_Q0.localhost,S=11:2`
	test.Assert(t, `With file exist: file name`, expNameCur, fnCur)
	test.Assert(t, `With file exist: message`, string(expMsg), string(msg))

	// The file in "new" should not exist.
	var pathNew = filepath.Join(mg.dirNew, fnNew)

	_, err = os.Stat(pathNew)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf(`want error %q, got %q`, os.ErrNotExist, err)
	}
}

func TestIncoming(t *testing.T) {
	var (
		mg       *Manager
		fnNew    string
		expError string
		err      error
	)

	mg, err = NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Case: empty message.

	_, err = mg.Incoming(nil)
	if err != nil {
		expError = `Incoming: empty message`
		test.Assert(t, `With empty message`, expError, err.Error())
	}

	// Case: success.

	var msg = []byte("From: me@localhost")

	fnNew, err = mg.Incoming(msg)
	if err != nil {
		t.Fatal(err)
	}

	var pathNew = filepath.Join(mg.dirNew, filepath.Base(fnNew))
	assertFileContent(t, pathNew, msg)

	var expCounter int64 = 1
	test.Assert(t, `counter should increase`, expCounter, mg.counter)

	// Case: failed due to the file exist in new.

	mg.counter = 0

	_, err = mg.Incoming(msg)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			t.Fatalf(`want %q, got %q`, os.ErrExist, err)
		}
	}

	// Case: the generateUniqueName fail due the file already exist.

	var pathTmp = filepath.Join(mg.dirTmp, filepath.Base(fnNew))

	err = os.WriteFile(pathTmp, msg, 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mg.Incoming(msg)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			t.Fatalf(`want error %q, got %v`, os.ErrExist, err)
		}
	}
}

func TestOutgoingQueue(t *testing.T) {
	var (
		mg  *Manager
		err error
	)

	mg, err = NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Case: empty message.
	var expError string

	_, err = mg.OutgoingQueue(nil)
	if err != nil {
		expError = `OutgoingQueue: empty message`
		test.Assert(t, `With empty message`, expError, err.Error())
	}

	// Case: success.
	var (
		msg   = []byte("From: me@localhost")
		fnTmp string
	)

	fnTmp, err = mg.OutgoingQueue(msg)
	if err != nil {
		t.Fatal(err)
	}

	var pathTmp = filepath.Join(mg.dirTmp, fnTmp)
	assertFileContent(t, pathTmp, msg)
}

func assertDirExist(t *testing.T, dir string) {
	var (
		fi  os.FileInfo
		err error
	)
	fi, err = os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.IsDir() {
		t.Fatalf(`%s: expecting a directory, got %s`, dir, fi.Mode())
	}
}

func assertFileContent(t *testing.T, file string, exp []byte) {
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `assertFileContent: `+file, string(exp), string(got))
}
