// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	mockStderr bytes.Buffer
	mockStdout bytes.Buffer

	_testRepoDir   string
	_testRemoteURL string
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	_testRemoteURL = "file://" + wd + "/testdata/beku_test.git"
	_testRepoDir = wd + "/testdata/repotest"

	_stderr = &mockStderr
	_stdout = &mockStdout

	fmt.Printf("stdout type: %T\n", _stdout)
	fmt.Printf("stderr type: %T\n", _stderr)
	fmt.Printf("remote URL : %s\n", _testRemoteURL)
	fmt.Printf("repo dir   : %s\n", _testRepoDir)

	s := m.Run()

	os.Exit(s)
}

func TestClone(t *testing.T) {
	cases := []struct {
		desc, dest, expErr, expStderr, expStdout string
	}{{
		desc:      "Clone on non empty directory",
		dest:      "testdata/notempty",
		expErr:    "Clone: exit status 128",
		expStderr: "fatal: destination path '.' already exists and is not an empty directory.\n",
	}, {
		desc: "Clone on non existen directory",
		dest: _testRepoDir,
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()
		os.RemoveAll(_testRepoDir)

		err := Clone(_testRemoteURL, c.dest)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, mockStdout.String())
	}
}

func TestCheckoutRevision(t *testing.T) {
	cases := []struct {
		desc      string
		remote    string
		branch    string
		revision  string
		expErr    string
		expStderr string
		expStdout string
	}{{
		desc: "With empty revision",
	}, {
		desc:     "With unknown revision",
		branch:   "master",
		revision: "xxxyyyzzz",
		expErr:   "CheckoutRevision: exit status 128",
		expStderr: `fatal: ambiguous argument 'xxxyyyzzz': unknown revision or path not in the working tree.
Use '--' to separate paths from revisions, like this:
'git <command> [<revision>...] -- [<file>...]'
`,
	}, {
		desc:     "With valid revision",
		branch:   "master",
		revision: "d6ad9da",
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		err := CheckoutRevision(_testRepoDir, c.remote, c.branch, c.revision)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, mockStdout.String())
	}
}

func TestGetRemoteURL(t *testing.T) {
	cases := []struct {
		desc                 string
		remoteName           string
		expErr               string
		expStderr, expStdout string
		exp                  string
	}{{
		desc:       "With invalid remote name",
		remoteName: "upstream",
		expErr:     "GetRemote: Empty or invalid remote name",
	}, {
		desc: "With empty remote name",
		exp:  _testRemoteURL,
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		got, err := GetRemoteURL(_testRepoDir, c.remoteName)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "url", c.exp, got)
		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, mockStdout.String())
	}
}

func TestGetTag(t *testing.T) {
	cases := []struct {
		desc      string
		revision  string
		expErr    string
		expStderr string
		expStdout string
	}{{
		desc:      "With current HEAD",
		expErr:    "GetTag: exit status 128",
		expStderr: "fatal: no tag exactly matches 'd6ad9dabc61f72558013bb05e91bf273c491e39c'\n",
	}, {
		desc:      "With revision",
		revision:  "ec65455",
		expStdout: "v0.1.0",
	}, {
		desc:      "With revision",
		revision:  "582b912",
		expStdout: "v0.2.0",
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		got, err := GetTag(_testRepoDir, c.revision)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, got)
	}
}

func TestLatestCommit(t *testing.T) {
	cases := []struct {
		desc      string
		ref       string
		expErr    string
		expStderr string
		expStdout string
	}{{
		desc:      "With invalid ref",
		ref:       "upstream/master",
		expErr:    "LatestCommit: exit status 128",
		expStderr: "fatal: Needed a single revision\n",
	}, {
		desc:      "With empty ref",
		expStdout: "c9f69fb",
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		got, err := LatestCommit(_testRepoDir, c.ref)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, got)
	}
}

func TestLatestTag(t *testing.T) {
	cases := []struct {
		desc      string
		expErr    string
		expStderr string
		expStdout string
	}{{
		desc:      "With default ref",
		expStdout: "v0.2.0",
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		got, err := LatestTag(_testRepoDir)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, got)
	}
}

func TestLatestVersion(t *testing.T) {
	cases := []struct {
		desc      string
		expErr    string
		expStderr string
		expStdout string
		exp       string
	}{{
		desc: "With default repo",
		exp:  "v0.2.0",
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		got, err := LatestVersion(_testRepoDir)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "version", c.exp, got)
		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, mockStdout.String())
	}
}

func TestListTag(t *testing.T) {
	cases := []struct {
		desc   string
		exp    []string
		expErr string
	}{{
		desc: "With default repo",
		exp:  []string{"v0.1.0", "v0.2.0"},
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		got, err := ListTags(_testRepoDir)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "tags", c.exp, got)
	}
}

func TestLogRevisions(t *testing.T) {
	cases := []struct {
		desc         string
		prevRevision string
		nextRevision string
		expErr       string
		expStderr    string
		expStdout    string
	}{{
		desc: "With both revision are empty",
	}, {
		desc:         "With previous revision is empty",
		nextRevision: "582b912",
		expStdout: `582b912 Add feature B.
ec65455 Add feature A.
`,
	}, {
		desc:         "With next revision is empty",
		prevRevision: "582b912",
		expStdout: `582b912 Add feature B.
ec65455 Add feature A.
`,
	}, {
		desc:         "With previous and next revisions",
		prevRevision: "582b912",
		nextRevision: "c9f69fb",
		expStdout: `c9f69fb Rename test to main.go
`,
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		err := LogRevisions(_testRepoDir, c.prevRevision, c.nextRevision)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
		}

		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, mockStdout.String())
	}
}

func TestRemoteChange(t *testing.T) {
	cases := []struct {
		desc                     string
		oldName, newName, newURL string
		expErr                   string
		expStderr                string
		expStdout                string
	}{{
		desc:      "With empty oldName",
		expErr:    "RemoteChange: exit status",
		expStderr: "fatal: No such remote: ''\n",
	}, {
		desc:      "With empty newName",
		oldName:   "origin",
		expErr:    "RemoteChange: exit status",
		expStderr: "fatal: '' is not a valid remote name\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)
		mockStderr.Reset()
		mockStdout.Reset()

		err := RemoteChange(_testRepoDir, c.oldName, c.newName, c.newURL)
		if err != nil {
			if strings.Contains(err.Error(), c.expErr) {
				continue
			}
			t.Fatalf("expecting error like %q, got %q", c.expErr, err.Error())
		}

		test.Assert(t, "stderr", c.expStderr, mockStderr.String())
		test.Assert(t, "stdout", c.expStdout, mockStdout.String())
	}
}
