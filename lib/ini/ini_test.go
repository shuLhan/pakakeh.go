// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const (
	testdataInputIni          = "testdata/input.ini"
	testdataSectionDupIni     = "testdata/section_dup.ini"
	testdataVarMultiEmpty     = "testdata/var_multi_empty.ini"
	testdataVarMultiSection   = "testdata/var_multi_section.ini"
	testdataVarWithoutSection = "testdata/var_without_section.ini"
)

var (
	inputIni *Ini
)

func TestOpen(t *testing.T) {
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

		_, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}
	}
}

func TestSave(t *testing.T) {
	cases := []struct {
		desc    string
		inFile  string
		outFile string
		expErr  string
	}{{
		desc:   "With no file",
		expErr: "open : no such file or directory",
	}, {
		desc:   "With variable without section",
		inFile: testdataVarWithoutSection,
		expErr: "variable without section, line 7 at testdata/var_without_section.ini",
	}, {
		desc:   "With empty output file",
		inFile: testdataInputIni,
		expErr: "open : no such file or directory",
	}, {
		desc:    "With valid output file",
		inFile:  testdataInputIni,
		outFile: testdataInputIni + ".save",
	}}

	for _, c := range cases {
		t.Logf("%+v", c)

		ini, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		err = ini.Save(c.outFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
		}
	}
}

func TestAddSection(t *testing.T) {
	in := &Ini{}

	cases := []struct {
		desc   string
		sec    *Section
		expIni *Ini
	}{{
		desc:   "With nil section",
		expIni: &Ini{},
	}, {
		desc: "With valid section",
		sec: &Section{
			mode:  varModeSection,
			name:  []byte("Test"),
			_name: []byte("test"),
		},
		expIni: &Ini{
			secs: []*Section{{
				mode:  varModeSection,
				name:  []byte("Test"),
				_name: []byte("test"),
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		in.AddSection(c.sec)

		test.Assert(t, "ini", c.expIni, in, true)
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

func TestGetString(t *testing.T) {
	cfg, err := Open(testdataInputIni)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc string
		sec  string
		sub  string
		key  string
		def  string
		exp  string
	}{{
		desc: "With empty params",
	}, {
		desc: "With non existen key",
		sec:  "test",
		key:  "key",
		def:  "def",
		exp:  "def",
	}, {
		desc: "With valid key, empty default",
		sec:  "user",
		key:  "name",
		exp:  "Shulhan",
	}}

	var got string

	for _, c := range cases {
		got = cfg.GetString(c.sec, c.sub, c.key, c.def)

		test.Assert(t, "string", c.exp, got, true)
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

func TestGetSectionDup(t *testing.T) {
	cases := []struct {
		sec     string
		sub     string
		keys    []string
		expOK   []bool
		expVals []string
	}{{
		sec: "core",
		keys: []string{
			"dupkey",
			"old",
			"new",
			"test",
		},
		expOK: []bool{
			true,
			true,
			true,
			false,
		},
		expVals: []string{
			"2",
			"value",
			"value",
			"",
		},
	}}

	cfg, err := Open(testdataSectionDupIni)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c)

		for x, k := range c.keys {
			t.Log("  Get:", k)

			got, ok := cfg.Get(c.sec, c.sub, k)
			if !ok {
				test.Assert(t, "ok", c.expOK[x], ok, true)
				continue
			}

			test.Assert(t, k, c.expVals[x], string(got), true)
		}
	}
}

func TestGetVarMultiEmpty(t *testing.T) {
	cases := []struct {
		sec     string
		sub     string
		keys    []string
		expOK   []bool
		expVals []string
	}{{
		sec: "alias",
		keys: []string{
			"tree",
			"test",
		},
		expOK: []bool{
			true,
			false,
		},
		expVals: []string{
			"!git --no-pager log --graph ",
			"",
		},
	}, {
		sec: "section",
		keys: []string{
			"tree",
			"test",
		},
		expOK: []bool{
			false,
			true,
		},
		expVals: []string{
			"",
			"true",
		},
	}}

	cfg, err := Open(testdataVarMultiEmpty)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c)

		for x, k := range c.keys {
			t.Log("  Get:", k)

			got, ok := cfg.Get(c.sec, c.sub, k)
			if !ok {
				test.Assert(t, "ok", c.expOK[x], ok, true)
				continue
			}

			test.Assert(t, k, c.expVals[x], string(got), true)
		}
	}
}

func TestGetVarMultiSection(t *testing.T) {
	cases := []struct {
		sec     string
		sub     string
		keys    []string
		expOK   []bool
		expVals []string
	}{{
		sec: "alias",
		keys: []string{
			"tree",
			"test",
		},
		expOK: []bool{
			true,
			true,
		},
		expVals: []string{
			"!git --no-pager log --graph [section]",
			"true",
		},
	}, {
		sec: "section",
		keys: []string{
			"test",
		},
		expOK: []bool{
			false,
		},
		expVals: []string{
			"true",
		},
	}}

	cfg, err := Open(testdataVarMultiSection)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c)

		for x, k := range c.keys {
			t.Log("  Get:", k)

			got, ok := cfg.Get(c.sec, c.sub, k)
			if !ok {
				test.Assert(t, "ok", c.expOK[x], ok, true)
				continue
			}

			test.Assert(t, k, c.expVals[x], string(got), true)
		}
	}
}

func TestGetSections(t *testing.T) {
	cfg, err := Open(testdataInputIni)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc string
		name string
		exp  []*Section
	}{{
		desc: "With empty name",
	}, {
		desc: "With name: unknown",
		name: "unknown",
	}, {
		desc: "With valid name: core",
		name: "core",
		exp: []*Section{{
			mode:    varModeSection,
			LineNum: 8,
			format:  []byte("[%s]\n"),
			name:    []byte("core"),
			_name:   []byte("core"),
			Vars: []*Variable{{
				mode:    varModeComment,
				lineNum: 9,
				format: []byte("	%s\n"),
				others: []byte("; Don't trust file modes"),
			}, {
				mode:    varModeValue,
				lineNum: 10,
				format: []byte("	%s = false\n"),
				Key:   []byte("filemode"),
				_key:  []byte("filemode"),
				Value: []byte("false"),
			}, {
				mode:    varModeEmpty,
				lineNum: 11,
				format:  []byte("\n"),
			}, {
				mode:    varModeComment,
				lineNum: 12,
				format:  []byte("%s\n"),
				others:  []byte("; Our diff algorithm"),
			}},
		}, {
			mode:    varModeSection,
			LineNum: 18,
			format:  []byte("[%s]\n"),
			name:    []byte("core"),
			_name:   []byte("core"),
			Vars: []*Variable{{
				mode:    varModeValue,
				lineNum: 19,
				format: []byte("	%s=\"ssh\" for \"kernel.org\"\n"),
				Key:   []byte("gitProxy"),
				_key:  []byte("gitproxy"),
				Value: []byte("ssh for kernel.org"),
			}, {
				mode:    varModeValue | varModeComment,
				lineNum: 20,
				format: []byte("	%s=default-proxy %s\n"),
				Key:    []byte("gitProxy"),
				_key:   []byte("gitproxy"),
				Value:  []byte("default-proxy"),
				others: []byte("; for the rest"),
			}, {
				mode:    varModeEmpty,
				lineNum: 21,
				format:  []byte("\n"),
			}, {
				mode:    varModeComment,
				lineNum: 22,
				format:  []byte("%s\n"),
				others:  []byte("; User settings"),
			}},
		}, {
			mode:    varModeSection,
			LineNum: 63,
			format:  []byte("[%s]\n"),
			name:    []byte("core"),
			_name:   []byte("core"),
			Vars: []*Variable{{
				mode:    varModeValue,
				lineNum: 64,
				format: []byte("	%s = less -R\n"),
				Key:   []byte("pager"),
				_key:  []byte("pager"),
				Value: []byte("less -R"),
			}, {
				mode:    varModeValue,
				lineNum: 65,
				format: []byte("	%s = nvim\n"),
				Key:   []byte("editor"),
				_key:  []byte("editor"),
				Value: []byte("nvim"),
			}, {
				mode:    varModeValue,
				lineNum: 66,
				format: []byte("	%s = false\n"),
				Key:   []byte("autocrlf"),
				_key:  []byte("autocrlf"),
				Value: []byte("false"),
			}, {
				mode:    varModeValue,
				lineNum: 67,
				format: []byte("	%s = true\n"),
				Key:   []byte("filemode"),
				_key:  []byte("filemode"),
				Value: []byte("true"),
			}},
		}},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := cfg.GetSections(c.name)

		test.Assert(t, "sections length", len(c.exp), len(got), true)

		for x := range c.exp {
			test.Assert(t, "variable length", len(c.exp[x].Vars),
				len(got[x].Vars), true)

			for y := range c.exp[x].Vars {
				t.Logf("var %d: %+v", y, c.exp[x].Vars[y])
				test.Assert(t, "variable", *c.exp[x].Vars[y],
					*got[x].Vars[y], true)
			}

			t.Logf("section %d: %+v", x, c.exp[x])
			test.Assert(t, "section", c.exp[x], got[x], true)
		}
	}
}
