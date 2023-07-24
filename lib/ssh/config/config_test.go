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
	type testCase struct {
		exp func() Section
		s   string
	}

	var (
		cfg *Config
		err error
	)

	cfg, err = Load(`./testdata/config`)
	if err != nil {
		t.Fatal(err)
	}

	var listTestCase = []testCase{{
		s: ``,
		exp: func() Section {
			var sec = *testDefaultSection
			return sec
		},
	}, {
		s: `example.local`,
		exp: func() Section {
			var sec = *testDefaultSection
			sec.name = `example.local`
			sec.IdentityFile = []string{
				filepath.Join(testDefaultSection.homeDir, `.ssh`, `notexist`),
			}
			sec.Field = map[string]string{
				KeyChallengeResponseAuthentication: ValueYes,
				KeyCheckHostIP:                     ValueYes,
				KeyConnectionAttempts:              DefConnectionAttempts,
				KeyHostname:                        `127.0.0.1`,
				KeyIdentityFile:                    `~/.ssh/notexist`,
				KeyPort:                            DefPort,
				KeyUser:                            `test`,
				KeyXAuthLocation:                   DefXAuthLocation,
			}
			return sec
		},
	}, {
		s: `my.example.local`,
		exp: func() Section {
			var sec = *testDefaultSection
			sec.name = `my.example.local`
			sec.IdentityFile = []string{
				filepath.Join(testDefaultSection.homeDir, `.ssh`, `notexist`),
			}
			sec.Field = map[string]string{
				KeyChallengeResponseAuthentication: ValueYes,
				KeyCheckHostIP:                     ValueYes,
				KeyConnectionAttempts:              DefConnectionAttempts,
				KeyHostname:                        `127.0.0.2`,
				KeyIdentityFile:                    `~/.ssh/notexist`,
				KeyPort:                            DefPort,
				KeyUser:                            `wildcard`,
				KeyXAuthLocation:                   DefXAuthLocation,
			}
			return sec
		},
	}, {
		s: `foo.local`,
		exp: func() Section {
			var sec = *testDefaultSection
			sec.name = `foo.local`
			sec.IdentityFile = []string{
				filepath.Join(testDefaultSection.homeDir, `.ssh`, `foo`),
				filepath.Join(testDefaultSection.homeDir, `.ssh`, `allfoo`),
			}
			sec.Field = map[string]string{
				KeyChallengeResponseAuthentication: ValueYes,
				KeyCheckHostIP:                     ValueYes,
				KeyConnectionAttempts:              DefConnectionAttempts,
				KeyHostname:                        `127.0.0.3`,
				KeyPort:                            DefPort,
				KeyUser:                            `allfoo`,
				KeyIdentityFile:                    `~/.ssh/allfoo`,
				KeyXAuthLocation:                   DefXAuthLocation,
			}
			return sec
		},
	}}

	var (
		c   testCase
		got *Section
	)

	for _, c = range listTestCase {
		got = cfg.Get(c.s)
		test.Assert(t, c.s, c.exp(), *got)
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
