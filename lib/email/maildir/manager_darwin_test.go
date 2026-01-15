// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package maildir

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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
	expError = `NewManager: initDirs: mkdir /cur: read-only file system`
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
