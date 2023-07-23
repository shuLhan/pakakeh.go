// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

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
		if c.exp != got {
			t.Fatalf("isIncludeDirective: %s: expecting %v, got %v",
				c.line, c.exp, got)
		}
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
		test.Assert(t, c.line, c.exp, got)
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
			`Host foo.local`,
			`Hostname 127.0.0.3`,
			`User foo`,
			`IdentityFile ~/.ssh/foo`,
			`Host *foo.local`,
			`User allfoo`,
			`IdentityFile ~/.ssh/allfoo`,
		},
	}}

	for _, c := range cases {
		got, err := readLines(c.file)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.file, c.exp, got)
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
		expError string
		exp      []string
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
			`Host foo.local`,
			`Hostname 127.0.0.3`,
			`User foo`,
			`IdentityFile ~/.ssh/foo`,
			`Host *foo.local`,
			`User allfoo`,
			`IdentityFile ~/.ssh/allfoo`,
		},
	}}

	for _, c := range cases {
		p, err := newParser()
		if err != nil {
			t.Fatal(err)
		}

		got, err := p.load(c.dir, c.pattern)
		if err != nil {
			if c.expError != err.Error() {
				t.Fatalf("parser.load: expecting error %v, got %v", c.expError, err)
			}
			continue
		}
		test.Assert(t, c.pattern, c.exp, got)
	}
}

func TestParseArgs(t *testing.T) {
	cases := []struct {
		raw string
		exp []string
	}{{
		raw: ``,
		exp: nil,
	}, {
		raw: `aa`,
		exp: []string{"aa"},
	}, {
		raw: `"aa"`,
		exp: []string{"aa"},
	}, {
		raw: `"a"  b  c`,
		exp: []string{"a", "b", "c"},
	}, {
		raw: `a "b c"`,
		exp: []string{"a", "b c"},
	}, {
		raw: `a "b c"  d   `,
		exp: []string{"a", "b c", "d"},
	}, {
		raw: `a "b c`,
		exp: []string{"a", "b c"},
	}}

	for _, c := range cases {
		got := parseArgs(c.raw, ' ')

		test.Assert(t, c.raw, c.exp, got)
	}
}
