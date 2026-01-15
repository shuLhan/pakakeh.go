// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package maildir

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewFolder(t *testing.T) {
	// Case: empty name.
	var (
		dir      = t.TempDir()
		name     = " \r\n"
		expError = `NewFolder: empty folder name`

		folder *Folder
		err    error
	)

	_, err = NewFolder(dir, name)
	test.Assert(t, `With empty name`, expError, err.Error())

	// Case: with no "maildirfolder".
	name = `.folder`
	_, err = NewFolder(dir, name)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf(`want error %q, got %q`, os.ErrNotExist, err)
	}

	// Case: with no "cur" folder.
	name = `.maildir_without_cur`

	folder, err = CreateFolder(dir, name)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(folder.dirCur)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewFolder(dir, name)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf(`want error %s, got %s`, os.ErrNotExist, err)
	}

	// Case: with invalid "tmp" folder permission.
	name = `.maildir_with_tmp_permission`

	folder, err = CreateFolder(dir, name)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chmod(folder.dirTmp, 0500)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewFolder(dir, name)
	if !errors.Is(err, os.ErrPermission) {
		t.Fatalf(`want error %s, got %s`, os.ErrPermission, err)
	}

	// Case: success.
	name = `.folder`

	_, err = CreateFolder(dir, name)
	if err != nil {
		t.Fatal(err)
	}

	folder, err = NewFolder(dir, name)
	if err != nil {
		t.Fatal(err)
	}

	err = checkDir(folder.dirCur)
	if err != nil {
		t.Fatal(err)
	}
	err = checkDir(folder.dirNew)
	if err != nil {
		t.Fatal(err)
	}
	err = checkDir(folder.dirTmp)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFolder(t *testing.T) {
	var (
		dir    = t.TempDir()
		name   = `.folder`
		folder *Folder
		err    error
	)

	folder, err = CreateFolder(dir, name)
	if err != nil {
		t.Fatal(err)
	}

	t.Run(`Delete`, func(tt *testing.T) {
		testFolderDelete(tt, folder)
	})
	t.Run(`Fetch`, func(tt *testing.T) {
		testFolderFetch(tt, folder)
	})
	t.Run(`Get`, func(tt *testing.T) {
		testFolderGet(tt, folder)
	})
}

func TestSanitizeFolderName(t *testing.T) {
	type testCase struct {
		name     string
		expError string
		exp      string
	}
	var cases = []testCase{{
		name:     ``,
		expError: `folder name is empty`,
	}, {
		name:     `.`,
		expError: `folder name is empty`,
	}, {
		name:     `notperiod`,
		expError: `folder name must begin with period`,
	}, {
		name:     `..`,
		expError: `folder name must not begin with ".."`,
	}, {
		name:     `..name`,
		expError: `folder name must not begin with ".."`,
	}, {
		name:     ".\u0001name",
		expError: `folder name contains unprintable character '\x01'`,
	}, {
		name:     ".\rname",
		expError: `folder name contains unprintable character '\r'`,
	}, {
		name:     ".\nname",
		expError: `folder name contains unprintable character '\n'`,
	}, {
		name:     ".na/me",
		expError: `folder name must not contains slash '/'`,
	}, {
		name: ".name\n", // The space is trimmed.
		exp:  `.name`,
	}, {
		name: `.folder.sub`,
		exp:  `.folder.sub`,
	}}
	var (
		c        testCase
		gotError string
		got      string
		err      error
	)
	for _, c = range cases {
		got, err = sanitizeFolderName(c.name)
		if err != nil {
			gotError = err.Error()
		}
		test.Assert(t, c.name, c.expError, gotError)
		test.Assert(t, c.name, c.exp, got)
		gotError = ``
	}
}

func testFolderDelete(t *testing.T, folder *Folder) {
	// Case: name is empty.
	var (
		name string
		err  error
	)
	err = folder.Delete(name)
	if err != nil {
		t.Fatalf(`want no error, got %s`, err)
	}

	// Case: file not exist.
	name = `filenotexist`
	err = folder.Delete(name)
	if err != nil {
		t.Fatalf(`want no error, got %s`, err)
	}

	// Case: file exist.
	name = `a:2`
	var fileCur = filepath.Join(folder.dirCur, name)

	err = os.WriteFile(fileCur, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = folder.Delete(name)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(fileCur)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf(`want error %q, got %q`, os.ErrNotExist, err)
	}
}

func testFolderFetch(t *testing.T, folder *Folder) {
	// Case: name is empty.
	var (
		name string
		err  error
	)
	_, _, err = folder.Fetch(name)
	if err != nil {
		t.Fatalf(`want no error, got %s`, err)
	}

	// Case: file not exist.
	name = `filenotexist`
	_, _, err = folder.Fetch(name)
	if err != nil {
		t.Fatalf(`want no error, got %s`, err)
	}

	// Case: file exist.
	name = `a`
	var (
		fileNew = filepath.Join(folder.dirNew, name)
		msg     = []byte(`content of file`)

		fileCur string
	)

	err = os.WriteFile(fileNew, msg, 0600)
	if err != nil {
		t.Fatal(err)
	}

	fileCur, msg, err = folder.Fetch(name)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(fileNew)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf(`want error %q, got %q`, os.ErrNotExist, err)
	}

	test.Assert(t, `content`, `content of file`, string(msg))
	test.Assert(t, `fileCur`, `a:2`, fileCur)
}

func testFolderGet(t *testing.T, folder *Folder) {
	// Case: file not exist.
	var (
		name = `filenotexist`

		expMsg []byte
		gotMsg []byte
		err    error
	)

	gotMsg, err = folder.Get(name)
	if err != nil {
		t.Fatalf(`want no error, got %s`, err)
	}
	test.Assert(t, `msg`, expMsg, gotMsg)

	// Case: file exist.
	name = `a:2`
	expMsg = []byte(`content of file`)

	var fileCur = filepath.Join(folder.dirNew, name)

	err = os.WriteFile(fileCur, expMsg, 0600)
	if err != nil {
		t.Fatal(err)
	}

	gotMsg, err = folder.Get(name)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `content`, string(expMsg), string(gotMsg))
}

func assertFolder(t *testing.T, folder *Folder) {
	var err error

	_, err = os.Stat(folder.dirCur)
	if err != nil {
		t.Fatalf(`want %s, got %s`, folder.dirCur, err)
	}
	_, err = os.Stat(folder.dirNew)
	if err != nil {
		t.Fatalf(`want %s, got %s`, folder.dirNew, err)
	}
	_, err = os.Stat(folder.dirTmp)
	if err != nil {
		t.Fatalf(`want %s, got %s`, folder.dirTmp, err)
	}

	var fileMd = filepath.Join(folder.dir, fileMaildirFolder)

	_, err = os.Stat(fileMd)
	if err != nil {
		t.Fatalf(`want %s, got %s`, fileMd, err)
	}
}
