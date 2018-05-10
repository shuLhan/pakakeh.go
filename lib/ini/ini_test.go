package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const (
	testdataInputIni          = "testdata/input.ini"
	testdataVarWithoutSection = "testdata/var_without_section.ini"
)

var (
	inputIni *Ini
)

func TestOpen(t *testing.T) {
	cases := []struct {
		desc       string
		inFile     string
		expErr     string
		expErrSave string
	}{{
		desc:   "With no file",
		expErr: "open : no such file or directory",
	}, {
		desc:   "With variable without section",
		inFile: testdataVarWithoutSection,
		expErr: "variable without section, line 7 at testdata/var_without_section.ini",
	}, {
		desc:   "With valid file",
		inFile: "testdata/input.ini",
	}}

	for _, c := range cases {
		t.Logf("%+v", c)

		in, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		err = in.Save(c.inFile + ".save")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSave(t *testing.T) {
	cases := []struct {
		desc   string
		inFile string
		expErr string
	}{{
		desc:   "With no file",
		expErr: "open : no such file or directory",
	}, {
		desc:   "With variable without section",
		inFile: testdataVarWithoutSection,
		expErr: "variable without section, line 7 at testdata/var_without_section.ini",
	}, {
		desc:   "With valid file",
		inFile: "testdata/input.ini",
	}}

	for _, c := range cases {
		t.Logf("%+v", c)

		ini, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		err = ini.Save(c.inFile + ".save")
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
		}
	}
}

func TestGet(t *testing.T) {
	var (
		err error
		got []byte
		ok  bool
	)

	inputIni, err = Open(testdataInputIni)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc   string
		sec    string
		sub    string
		key    string
		expVal []byte
		expOk  bool
	}{{
		desc: `With empty section`,
		sub:  "devel",
		key:  "remote",
	}, {
		desc:   `With empty subsection`,
		sec:    "user",
		key:    "name",
		expVal: []byte("Shulhan"),
		expOk:  true,
	}, {
		desc: `With empty key`,
		sec:  "user",
	}, {
		desc: `With invalid section`,
		sec:  "sectionnotexist",
		key:  "name",
	}, {
		desc: `With invalid subsection`,
		sec:  "branch",
		sub:  "notexist",
		key:  "remote",
	}, {
		desc: `With invalid key`,
		sec:  "branch",
		sub:  "devel",
		key:  "user",
	}, {
		desc:   `With empty value`,
		sec:    "http",
		key:    "sslVerify",
		expVal: []byte("true"),
	}}

	for _, c := range cases {
		t.Logf("%+v", c)

		got, ok = inputIni.Get(c.sec, c.sub, c.key)
		if !ok {
			test.Assert(t, "ok", c.expOk, ok, true)
			continue
		}

		test.Assert(t, "value", c.expVal, got, true)
	}
}

func TestGetInputIni(t *testing.T) {
	cases := []struct {
		sec     string
		sub     string
		keys    []string
		expVals []string
	}{{
		sec: "core",
		keys: []string{
			"filemode",
			"gitProxy",
			"pager",
			"editor",
			"autocrlf",
		},
		expVals: []string{
			"true",
			"default-proxy",
			"less -R",
			"nvim",
			"false",
		},
	}, {
		sec: "diff",
		keys: []string{
			"external",
			"renames",
		},
		expVals: []string{
			"/usr/local/bin/diff-wrapper",
			"true",
		},
	}, {
		sec: "user",
		keys: []string{
			"name",
			"email",
		},
		expVals: []string{
			"Shulhan",
			"ms@kilabit.info",
		},
	}, {
		sec: "http",
		keys: []string{
			"sslVerify",
			"cookiefile",
		},
		expVals: []string{
			"true",
			"/home/ms/.gitcookies",
		},
	}, {
		sec: "http",
		sub: "https://weak.example.com",
		keys: []string{
			"sslVerify",
			"cookiefile",
		},
		expVals: []string{
			"false",
			"/tmp/cookie.txt",
		},
	}, {
		sec: "branch",
		sub: "devel",
		keys: []string{
			"remote",
			"merge",
		},
		expVals: []string{
			"origin",
			"refs/heads/devel",
		},
	}, {
		sec: "include",
		keys: []string{
			"path",
		},
		expVals: []string{
			"~/foo.inc",
		},
	}, {
		sec: "includeIf",
		sub: "gitdir:/path/to/foo/.git",
		keys: []string{
			"path",
		},
		expVals: []string{
			"/path/to/foo.inc",
		},
	}, {
		sec: "includeIf",
		sub: "gitdir:/path/to/group/",
		keys: []string{
			"path",
		},
		expVals: []string{
			"foo.inc",
		},
	}, {
		sec: "includeIf",
		sub: "gitdir:~/to/group/",
		keys: []string{
			"path",
		},
		expVals: []string{
			"/path/to/foo.inc",
		},
	}, {
		sec:     "color",
		keys:    []string{"ui"},
		expVals: []string{"true"},
	}, {
		sec: "gui",
		keys: []string{
			"fontui",
			"fontdiff",
			"diffcontext",
			"spellingdictionary",
		},
		expVals: []string{
			"-family \"xos4 Terminus\" -size 10 -weight normal -slant roman -underline 0 -overstrike 0",
			"-family \"xos4 Terminus\" -size 10 -weight normal -slant roman -underline 0 -overstrike 0",
			"4",
			"none",
		},
	}, {
		sec: "svn",
		keys: []string{
			"rmdir",
		},
		expVals: []string{
			"true",
		},
	}, {
		sec: "alias",
		keys: []string{
			"change",
			"gofmt",
			"mail",
			"pending",
			"submit",
			"sync",
			"tree",
			"to-master",
			"to-prod",
		},
		expVals: []string{
			"codereview change",
			"codereview gofmt",
			"codereview mail",
			"codereview pending",
			"codereview submit",
			"codereview sync",
			`!git --no-pager log --graph 		--date=format:'%Y-%m-%d' 		--pretty=format:'%C(auto,dim)%ad %<(7,trunc) %an %Creset%m %h %s %Cgreen%d%Creset' 		--exclude=*/production 		--exclude=*/dev-* 		--all -n 20`,
			`!git stash -u 		&& git fetch origin 		&& git rebase origin/master 		&& git stash pop 		&& git --no-pager log --graph --decorate --pretty=oneline 			--abbrev-commit origin/master~1..HEAD`,
			`!git stash -u 		&& git fetch origin 		&& git rebase origin/production 		&& git stash pop 		&& git --no-pager log --graph --decorate --pretty=oneline 			--abbrev-commit origin/production~1..HEAD`,
		},
	}, {
		sec: "url",
		sub: "git@github.com:",
		keys: []string{
			"insteadOf",
		},
		expVals: []string{
			"https://github.com/",
		},
	}}

	var (
		got []byte
		ok  bool
	)

	for _, c := range cases {
		t.Log(c)

		if debug >= debugL2 {
			t.Logf("Section header: [%s %s]", c.sec, c.sub)
			t.Logf(">>> keys: %s", c.keys)
			t.Logf(">>> expVals: %s", c.expVals)
		}

		for x, k := range c.keys {
			t.Log("  Get:", k)

			got, ok = inputIni.Get(c.sec, c.sub, k)
			if !ok {
				t.Logf("Get: %s > %s > %s", c.sec, c.sub, k)
				test.Assert(t, "ok", true, ok, true)
				t.FailNow()
			}

			test.Assert(t, "value", []byte(c.expVals[x]), got, true)
		}
	}
}
