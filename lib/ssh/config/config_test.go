// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	testDefaultSection = newSection(``)
	testParser         *parser
)

func TestMain(m *testing.M) {
	var err error

	testParser, err = newParser()
	if err != nil {
		log.Fatal(err)
	}

	testDefaultSection.init(testParser.workDir, testParser.homeDir)

	os.Exit(m.Run())
}

func TestPatternToRegex(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in:  "*",
		exp: ".*",
	}, {
		in:  "?",
		exp: ".?",
	}, {
		in:  "192.*",
		exp: `192\..*`,
	}}

	for _, c := range cases {
		got := patternToRegex(c.in)
		if c.exp != got {
			t.Fatalf("patternToRegex: expecting %s, got %s", c.exp, got)
		}
	}
}

func TestConfig_Get(t *testing.T) {
	cfg, err := Load("./testdata/config")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		exp func(def Section) *Section
		s   string
	}{{
		s: "",
		exp: func(def Section) *Section {
			return nil
		},
	}, {
		s: "example.local",
		exp: func(def Section) *Section {
			def.name = `example.local`
			def.Hostname = "127.0.0.1"
			def.User = "test"
			def.PrivateKeyFile = ""
			def.IdentityFile = []string{
				filepath.Join(def.homeDir, ".ssh", "notexist"),
			}
			def.useDefaultIdentityFile = false
			def.Field = map[string]string{
				`hostname`:     `127.0.0.1`,
				`user`:         `test`,
				`identityfile`: `~/.ssh/notexist`,
			}
			return &def
		},
	}, {
		s: "my.example.local",
		exp: func(def Section) *Section {
			def.name = `*.example.local`
			def.Hostname = "127.0.0.2"
			def.User = "wildcard"
			def.PrivateKeyFile = ""
			def.IdentityFile = []string{
				filepath.Join(def.homeDir, ".ssh", "notexist"),
			}
			def.useDefaultIdentityFile = false
			def.Field = map[string]string{
				`hostname`:     `127.0.0.2`,
				`user`:         `wildcard`,
				`identityfile`: `~/.ssh/notexist`,
			}
			return &def
		},
	}}

	for _, c := range cases {
		got := cfg.Get(c.s)

		// Clear the patterns and criteria for comparison.
		if got != nil {
			got.patterns = nil
			got.criteria = nil
			got.init(testParser.workDir, testParser.homeDir)
		}

		exp := c.exp(*testDefaultSection)
		if exp != nil {
			exp.init(testParser.workDir, testParser.homeDir)
		} else if got == nil {
			continue
		}
		test.Assert(t, c.s, *exp, *got)
	}
}

func TestParseKeyValue(t *testing.T) {
	cases := []struct {
		line     string
		expKey   string
		expValue string
		expError string
	}{{
		line:     `a b`,
		expKey:   "a",
		expValue: "b",
	}, {
		line:     `a    b`,
		expKey:   "a",
		expValue: "b",
	}, {
		line:     `a   =b`,
		expKey:   "a",
		expValue: "b",
	}, {
		line:     `a   "b c"`,
		expKey:   "a",
		expValue: "b c",
	}, {
		line:     `a   ="b c"`,
		expKey:   "a",
		expValue: "b c",
	}, {
		line:     `a   ==b`,
		expError: errMultipleEqual.Error(),
	}}

	for _, c := range cases {
		key, value, err := parseKeyValue(c.line)
		if err != nil {
			if c.expError != err.Error() {
				t.Fatalf("parseKeyValue: expecting error %s, got %s",
					c.expError, err)
			}
			continue
		}

		if c.expKey != key {
			t.Fatalf("parseKeyValue: expecting key %s, got %s",
				c.expKey, key)
		}
		if c.expValue != value {
			t.Fatalf("parseKeyValue: expecting value %s, got %s",
				c.expValue, value)
		}
	}
}
