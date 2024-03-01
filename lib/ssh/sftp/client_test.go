// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

import (
	"io/fs"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestClient_Fsetstat(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	remoteFile := "/tmp/lib-ssh-sftp-fsetstat.test"

	fa := newFileAttrs()
	fa.SetPermissions(0766)

	fh, err := testClient.Create(remoteFile, fa)
	if err != nil {
		t.Fatal(err)
	}

	fa, err = testClient.Fstat(fh)
	if err != nil {
		t.Fatal(err)
	}

	exp := uint32(0o100600)
	fa.SetPermissions(exp)

	err = testClient.Fsetstat(fh, fa)
	if err != nil {
		t.Fatal(err)
	}

	fa, err = testClient.Fstat(fh)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Fsetstat", exp, fa.Permissions())
}

func TestClient_Fstat(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	remoteFile := "/etc/hosts"
	fh, err := testClient.Open(remoteFile)
	if err != nil {
		t.Fatal(err)
	}

	fa, err := testClient.Fstat(fh)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Fstat %s: %+v", remoteFile, fa)
}

func TestClient_Get(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	err := testClient.Get("/etc/hosts", "testdata/etc-hosts.get")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Lstat(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	remoteFile := "/etc/hosts"
	fa, err := testClient.Lstat(remoteFile)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Lstat: %s: %+v", remoteFile, fa)
}

func TestClient_Mkdir(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	type testCase struct {
		path     string
		expError string
	}

	var cases = []testCase{{
		path: `/tmp/lib-ssh-sftp-mkdir`,
	}, {
		path:     `/perm`,
		expError: fs.ErrPermission.Error(),
	}}

	var (
		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log(c.path)

		err = testClient.Mkdir(c.path, nil)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}

		err = testClient.Rmdir(c.path)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestClient_MkdirAll(t *testing.T) {
	if !isTestManual {
		t.Skipf(`%s not set`, envNameTestManual)
	}

	type testCase struct {
		expStat  *FileAttrs
		path     string
		expError string
	}

	var cases = []testCase{{
		path: `/tmp/a/b/c`,
		expStat: &FileAttrs{
			name:        `/tmp/a/b/c`,
			fsMode:      fs.ModeDir | 0700,
			flags:       attrSize | attrUIDGID | attrPermissions | attrAcModtime,
			size:        40,
			permissions: fileTypeDirectory | 0700,
			uid:         1000,
			gid:         1000,
		},
	}, {
		// Creating the same directory should not return an error.
		path: `/tmp/a/b/c`,
		expStat: &FileAttrs{
			name:        `/tmp/a/b/c`,
			fsMode:      fs.ModeDir | 0700,
			flags:       attrSize | attrUIDGID | attrPermissions | attrAcModtime,
			size:        40,
			permissions: fileTypeDirectory | 0700,
			uid:         1000,
			gid:         1000,
		},
	}, {
		path: ``,
		expStat: &FileAttrs{
			name:        `.`,
			fsMode:      fs.ModeDir | 0755,
			flags:       attrSize | attrUIDGID | attrPermissions | attrAcModtime,
			size:        4096,
			permissions: fileTypeDirectory | 0755,
			uid:         1000,
			gid:         33,
		},
	}, {
		path: `.cache/a/b/c`,
		expStat: &FileAttrs{
			name:        `.cache/a/b/c`,
			fsMode:      fs.ModeDir | 0700,
			flags:       attrSize | attrUIDGID | attrPermissions | attrAcModtime,
			size:        4096,
			permissions: fileTypeDirectory | 0700,
			uid:         1000,
			gid:         1000,
		},
	}, {
		// Creating the same directory should not return an error.
		path: `.cache/a/b/c`,
		expStat: &FileAttrs{
			name:        `.cache/a/b/c`,
			fsMode:      fs.ModeDir | 0700,
			flags:       attrSize | attrUIDGID | attrPermissions | attrAcModtime,
			size:        4096,
			permissions: fileTypeDirectory | 0700,
			uid:         1000,
			gid:         1000,
		},
	}, {
		path: `.cache/a b/c`,
		expStat: &FileAttrs{
			name:        `.cache/a b/c`,
			fsMode:      fs.ModeDir | 0700,
			flags:       attrSize | attrUIDGID | attrPermissions | attrAcModtime,
			size:        4096,
			permissions: fileTypeDirectory | 0700,
			uid:         1000,
			gid:         1000,
		},
	}}

	var (
		c       testCase
		gotStat *FileAttrs
		err     error
	)

	for _, c = range cases {
		t.Log(c.path)

		err = testClient.MkdirAll(c.path, nil)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}

		gotStat, err = testClient.Stat(c.path)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}

		// Exclude access and modification times from being checked.
		gotStat.atime = 0
		gotStat.mtime = 0

		test.Assert(t, `Stat `+c.path, c.expStat, gotStat)
	}
}

func TestClient_Put(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	err := testClient.Put("testdata/id_ed25519.pub", "/tmp/id_ed25519.pub")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Readdir(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	path := "/tmp"
	fh, err := testClient.Opendir(path)
	if err != nil {
		t.Fatal(err)
	}

	nodes, err := testClient.Readdir(fh)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("List of files inside the %s:\n", path)
	for x, node := range nodes {
		fi, _ := node.Info()
		t.Logf("%02d: %s %+v\n", x, fi.Mode().String(), node.Name())
	}
}

func TestClient_Realpath(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	node, err := testClient.Realpath("../../etc/hosts")
	if err != nil {
		t.Fatal(err)
	}

	exp := "/etc/hosts"
	test.Assert(t, "Realpath", exp, node.Name())
}

func TestClient_Rename(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	oldPath := "/tmp/lib-ssh-sftp-rename-old"
	newPath := "/tmp/lib-ssh-sftp-rename-new"

	_ = testClient.Remove(newPath)

	fh, err := testClient.Create(oldPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	expAttrs, err := testClient.Fstat(fh)
	if err != nil {
		t.Fatal(err)
	}

	err = testClient.CloseFile(fh)
	if err != nil {
		t.Fatal(err)
	}

	err = testClient.Rename(oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}

	gotAttrs, err := testClient.Stat(newPath)
	if err != nil {
		t.Fatal(err)
	}

	expAttrs.name = newPath

	test.Assert(t, "Rename", expAttrs, gotAttrs)
}

func TestClient_Setstat(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	remoteFile := "/tmp/lib-ssh-sftp-setstat.test"

	fa := newFileAttrs()
	fa.SetPermissions(0766)

	fh, err := testClient.Create(remoteFile, fa)
	if err != nil {
		t.Fatal(err)
	}

	fa, err = testClient.Fstat(fh)
	if err != nil {
		t.Fatal(err)
	}

	exp := uint32(0o100600)
	fa.SetPermissions(exp)

	err = testClient.Setstat(remoteFile, fa)
	if err != nil {
		t.Fatal(err)
	}

	fa, err = testClient.Fstat(fh)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Setstat", exp, fa.Permissions())
}

func TestClient_Stat(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	remoteFile := "/etc/hosts"
	fh, err := testClient.Open(remoteFile)
	if err != nil {
		t.Fatal(err)
	}

	fa, err := testClient.Fstat(fh)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Stat: %s: %+v", remoteFile, fa)
}

func TestClient_Symlink(t *testing.T) {
	if !isTestManual {
		t.Skipf("%s not set", envNameTestManual)
	}

	targetPath := "/tmp/lib-ssh-sftp-symlink-targetpath"
	linkPath := "/tmp/lib-ssh-sftp-symlink-linkpath"

	_ = testClient.Remove(linkPath)
	_ = testClient.Remove(targetPath)

	fh, err := testClient.Create(targetPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = testClient.CloseFile(fh)
	if err != nil {
		t.Fatal(err)
	}

	err = testClient.Symlink(targetPath, linkPath)
	if err != nil {
		t.Fatal(err)
	}

	node, err := testClient.Readlink(linkPath)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Readlink", targetPath, node.Name())
}
