// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/test"
)

const (
	testdataInputIni          = "testdata/input.ini"
	testdataVarWithoutSection = "testdata/var_without_section.ini"
)

type StructA struct {
	X int  `ini:"a::x"`
	Y bool `ini:"a::y"`
}

type StructB struct {
	StructA
	Z float64 `ini:"b::z"`
}

type StructC struct {
	StructB
	XX byte `ini:"c::xx"`
}

type StructMap struct {
	Amap map[string]string `ini:"test:map"`
}

type Y struct {
	String string `ini:"::string"`
	Int    int    `ini:"::int"`
}

type X struct {
	Time time.Time `ini:"section::time" layout:"2006-01-02 15:04:05"`

	PtrBool     *bool          `ini:"section:pointer:bool"`
	PtrDuration *time.Duration `ini:"section:pointer:duration"`
	PtrInt      *int           `ini:"section:pointer:int"`
	PtrString   *string        `ini:"section:pointer:string"`
	PtrTime     *time.Time     `ini:"section:pointer:time" layout:"2006-01-02 15:04:05"`

	PtrStruct    *Y `ini:"section:ptr_struct"`
	PtrStructNil *Y `ini:"section:ptr_struct_nil"`

	Struct Y `ini:"section:struct"`

	String string `ini:"section::string"`

	SliceStruct []Y `ini:"slice:struct"`

	SlicePtrBool     []*bool          `ini:"slice:ptr:bool"`
	SlicePtrDuration []*time.Duration `ini:"slice:ptr:duration"`
	SlicePtrInt      []*int           `ini:"slice:ptr:int"`
	SlicePtrString   []*string        `ini:"slice:ptr:string"`
	SlicePtrStruct   []*Y             `ini:"slice:ptr_struct"`
	SlicePtrTime     []*time.Time     `ini:"slice:ptr:time" layout:"2006-01-02 15:04:05"`

	SliceBool     []bool          `ini:"slice::bool"`
	SliceDuration []time.Duration `ini:"slice::duration"`
	SliceInt      []int           `ini:"slice::int"`
	SliceString   []string        `ini:"slice::string"`
	SliceTime     []time.Time     `ini:"slice::time" layout:"2006-01-02 15:04:05"`

	Duration time.Duration `ini:"section::duration"`
	Int      int           `ini:"section::int"`
	Bool     bool          `ini:"section::bool"`
}

func TestData(t *testing.T) {
	var (
		listTestData []*test.Data
		tdata        *test.Data
		err          error
	)

	listTestData, err = test.LoadDataDir("testdata/struct")
	if err != nil {
		t.Fatal(err)
	}

	for _, tdata = range listTestData {
		t.Run(tdata.Name, func(t *testing.T) {
			var (
				kind   = tdata.Flag["kind"]
				input  = tdata.Input["default"]
				expOut = tdata.Output["default"]
				gotX   = &X{}
				gotC   = &StructC{}
				gotMap = &StructMap{}

				obj    interface{}
				gotOut []byte
				err    error
			)

			switch kind {
			case "":
				return
			case "embedded":
				obj = gotC
			case "map":
				obj = gotMap
			case "struct":
				obj = gotX
			}

			err = Unmarshal(input, obj)
			if err != nil {
				t.Fatal(err)
			}

			gotOut, err = Marshal(obj)
			if err != nil {
				t.Fatal(err)
			}

			test.Assert(t, string(tdata.Desc), string(expOut), string(gotOut))
		})
	}
}

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
		t.Log(c.desc)

		_, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
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
		t.Logf(c.desc)

		cfg, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		err = cfg.Save(c.outFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
		}
	}
}

func TestAddSection(t *testing.T) {
	in := &Ini{}

	cases := []struct {
		sec    *Section
		expIni *Ini
		desc   string
	}{{
		desc:   "With nil section",
		expIni: &Ini{},
	}, {
		desc: "With valid section",
		sec: &Section{
			mode:      lineModeSection,
			name:      "Test",
			nameLower: "test",
		},
		expIni: &Ini{
			secs: []*Section{{
				mode:      lineModeSection,
				name:      "Test",
				nameLower: "test",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		in.addSection(c.sec)

		test.Assert(t, "ini", c.expIni, in)
	}
}

func TestGet(t *testing.T) {
	var (
		err error
		got string
		ok  bool
	)

	inputIni, err := Open(testdataInputIni)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc   string
		sec    string
		sub    string
		key    string
		expVal string
		expOk  bool
	}{{
		desc: `With empty section`,
		sub:  "devel",
		key:  "remote",
	}, {
		desc:   `With empty subsection`,
		sec:    "user",
		key:    "name",
		expVal: "Shulhan",
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
		desc: `With empty value`,
		sec:  "http",
		key:  "sslVerify",
	}}

	for _, c := range cases {
		t.Logf("%+v", c)

		got, ok = inputIni.Get(c.sec, c.sub, c.key, "")
		if !ok {
			test.Assert(t, "ok", c.expOk, ok)
			continue
		}

		test.Assert(t, "value", c.expVal, got)
	}
}

func TestGetInputIni(t *testing.T) {
	inputIni, err := Open(testdataInputIni)
	if err != nil {
		t.Fatal(err)
	}

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
			"",
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
	}, {
		sec: "last",
		keys: []string{
			"valid0",
			"valid1",
			"valid2",
			"valid3",
			"valid4",
		},
		expVals: []string{
			"",
			"",
			"",
			"",
			"",
		},
	}}

	var (
		got string
		ok  bool
	)

	for _, c := range cases {
		t.Log(c)

		if debug.Value >= 3 {
			t.Logf("Section header: [%s %s]", c.sec, c.sub)
			t.Logf(">>> keys: %s", c.keys)
			t.Logf(">>> expVals: %s", c.expVals)
		}

		for x, k := range c.keys {
			t.Log("  Get:", k)

			got, ok = inputIni.Get(c.sec, c.sub, k, "")
			if !ok {
				t.Logf("Get: %s > %s > %s", c.sec, c.sub, k)
				test.Assert(t, "ok", true, ok)
				t.FailNow()
			}

			test.Assert(t, "value", c.expVals[x], got)
		}
	}
}

func TestIni_Get(t *testing.T) {
	var (
		cfg   *Ini
		tdata *test.Data
		got   string
		def   string
		tags  []string
		keys  [][]byte
		exps  [][]byte
		key   []byte
		err   error
		x     int
		ok    bool
	)

	tdata, err = test.LoadData("testdata/get.txt")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err = Parse(tdata.Input["default"])
	if err != nil {
		t.Fatal(err)
	}

	keys = bytes.Split(tdata.Input["keys"], []byte("\n"))
	exps = bytes.Split(tdata.Output["default"], []byte("\n"))

	if len(keys) != len(exps) {
		t.Fatalf("%s: input keys length %d does not match with output %d",
			tdata.Name, len(keys), len(exps))
	}

	for x, key = range keys {
		if len(key) == 0 {
			test.Assert(t, "Get", string(exps[x]), "")
			continue
		}

		tags = parseTag(string(key))
		def = tags[3]

		got, ok = cfg.Get(tags[0], tags[1], tags[2], def)
		got = fmt.Sprintf("%t %s.", ok, got)

		test.Assert(t, "Get", string(exps[x]), got)
	}
}
