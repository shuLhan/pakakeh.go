// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsIncludeDirective(t *testing.T) {
	cases := []struct {
		line string
		exp  bool
	}{{
		line: "",
	}, {
		line: "includ",
	}, {
		line: "INCLUDE",
	}, {
		line: "INCLUDE ",
		exp:  true,
	}, {
		line: "INCLUDE=",
		exp:  true,
	}}

	for _, c := range cases {
		got := isIncludeDirective(c.line)
		test.Assert(t, "isIncludeDirective: "+c.line, c.exp, got, true)
	}
}

func TestParseInclude(t *testing.T) {
	cases := []struct {
		line string
		exp  []string
	}{{
		line: "include a",
		exp:  []string{"a"},
	}, {
		line: "include a b",
		exp:  []string{"a", "b"},
	}, {
		line: `include "a`,
		exp:  []string{"a"},
	}, {
		line: `include "a b" "c"`,
		exp:  []string{"a b", "c"},
	}}

	for _, c := range cases {
		got := parseInclude(c.line)
		test.Assert(t, "parseInclude: "+c.line, c.exp, got, true)
	}
}

func TestReadLines(t *testing.T) {
	cases := []struct {
		file string
		exp  []string
	}{{
		file: "testdata/config",
		exp: []string{
			`Include config.local`,
			`Host example.local`,
			`Hostname 127.0.0.1`,
			`User test`,
			`IdentityFile ~/.ssh/notexist`,
			`Host *.example.local`,
			`Include sub/config`,
		},
	}}

	for _, c := range cases {
		got, err := readLines(c.file)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "readLines", c.exp, got, true)
	}
}

func TestConfigParser_load(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		dir      string
		pattern  string
		exp      []string
		expError string
	}{{
		dir:     wd,
		pattern: "testdata/config",
		exp: []string{
			`Host local`,
			`Hostname local`,
			`Host example.local`,
			`Hostname 127.0.0.1`,
			`User test`,
			`IdentityFile ~/.ssh/notexist`,
			`Host *.example.local`,
			`Hostname 127.0.0.2`,
			`User wildcard`,
			`IdentityFile ~/.ssh/notexist`,
		},
	}}

	for _, c := range cases {
		parser, err := newConfigParser()
		if err != nil {
			t.Fatal(err)
		}

		got, err := parser.load(c.dir, c.pattern)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}
		test.Assert(t, "load "+c.pattern, c.exp, got, true)
	}
}
