// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestConfig_Get(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := NewConfig("./testdata/config")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		s   string
		exp func(def ConfigSection) *ConfigSection
	}{{
		s: "",
		exp: func(def ConfigSection) *ConfigSection {
			return nil
		},
	}, {
		s: "example.local",
		exp: func(def ConfigSection) *ConfigSection {
			def.Hostname = "127.0.0.1"
			def.User = "test"
			def.IdentityFile = []string{
				filepath.Join(homeDir, "/.ssh/test"),
			}
			def.useDefaultIdentityFile = false
			return &def
		},
	}, {
		s: "my.example.local",
		exp: func(def ConfigSection) *ConfigSection {
			def.Hostname = "127.0.0.2"
			def.User = "wildcard"
			def.IdentityFile = []string{
				filepath.Join(homeDir, "/.ssh/wildcard"),
			}
			def.useDefaultIdentityFile = false
			return &def
		},
	}}

	for _, c := range cases {
		got := cfg.Get(c.s)

		// Clear the patterns and criterias for comparison.
		if got != nil {
			got.patterns = nil
			got.criterias = nil
		}
		test.Assert(t, "Get "+c.s, c.exp(*testDefaultSection), got, true)
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
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}
		test.Assert(t, "key:", c.expKey, key, true)
		test.Assert(t, "value:", c.expValue, value, true)
	}
}
