// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"log"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

var (
	dummyConfig        *Config
	testDefaultSection *Section
	testParser         *parser
)

func TestMain(m *testing.M) {
	var err error

	dummyConfig, err = newConfig(``)
	if err != nil {
		log.Fatal(err)
	}

	testParser = newParser(dummyConfig)

	testDefaultSection = NewSection(dummyConfig, ``)
	testDefaultSection.setDefaults()

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

func TestConfigGet(t *testing.T) {
	type testCase struct {
		name string
		exp  string
	}

	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/config_get_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var cfg *Config

	cfg, err = Load(`./testdata/config`)
	if err != nil {
		t.Fatal(err)
	}

	var cases = []testCase{{
		name: ``,
		exp:  string(tdata.Output[`empty`]),
	}, {
		name: `example.local`,
		exp:  string(tdata.Output[`example.local`]),
	}, {
		name: `my.example.local`,
		exp:  string(tdata.Output[`my.example.local`]),
	}, {
		name: `foo.local`,
		exp:  string(tdata.Output[`foo.local`]),
	}, {
		// With Hostname key not set but match Host wildcard.
		name: `my.foo.local`,
		exp:  string(tdata.Output[`my.foo.local`]),
	}}

	var (
		section *Section
		buf     bytes.Buffer
		c       testCase
	)
	for _, c = range cases {
		section = cfg.Get(c.name)
		buf.Reset()
		_, err = section.WriteTo(&buf)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.name, c.exp, buf.String())
	}
}

func TestConfigMerge(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/config_merge_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var cfg *Config

	cfg, err = Load(`./testdata/sub/config`)
	if err != nil {
		t.Fatal(err)
	}

	var topcfg *Config

	topcfg, err = Load(`./testdata/config`)
	if err != nil {
		t.Fatal(err)
	}

	cfg.Merge(topcfg)

	var (
		host       = `my.example.local`
		gotSection = cfg.Get(host)

		buf bytes.Buffer
	)

	_, err = gotSection.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, host, string(tdata.Output[host]), buf.String())
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
