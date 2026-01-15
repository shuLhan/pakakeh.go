// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package maildir

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestManager_scanDir(t *testing.T) {
	var (
		dir = t.TempDir()

		folder *Folder
		err    error
	)

	// Folder without dot.
	var dirFolder = filepath.Join(dir, `nodot`)
	err = os.Mkdir(dirFolder, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Folder without "maildirfolder".
	dirFolder = filepath.Join(dir, `.nomaildir`)
	err = os.Mkdir(dirFolder, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Folder with no permission.
	dirFolder = filepath.Join(dir, `.noperm`)
	err = os.Mkdir(dirFolder, 0400)
	if err != nil {
		t.Fatal(err)
	}

	// Folder with "maildirfolder" but no "cur".
	dirFolder = filepath.Join(dir, `.nocur`)
	err = os.Mkdir(dirFolder, 0400)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(dir, fileMaildirFolder), nil, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Folder with "maildirfolder".
	folder, err = CreateFolder(dir, `.folder`)
	if err != nil {
		t.Fatal(err)
	}

	var md *Manager

	md, err = NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	var expFolders = map[string]*Folder{
		`.folder`: folder,
	}
	test.Assert(t, `folders`, expFolders, md.folders)
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
	fnCur, _, err = mg.Fetch(fnNew)
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

	fnCur, msg, err = mg.Fetch(``)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `With empty file name: file name`, ``, fnCur)
	test.Assert(t, `With empty file name: message`, expMsg, msg)

	// Case: file not exist.

	fnCur, msg, err = mg.Fetch(`filenotexist`)
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
	fnCur, msg, err = mg.Fetch(fnNew)
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
