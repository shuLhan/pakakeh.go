// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewSectionHost(t *testing.T) {
	cases := []struct {
		exp        func(def Section) *Section
		rawPattern string
	}{{
		rawPattern: "",
		exp: func(exp Section) *Section {
			exp.patterns = make([]*pattern, 0)
			return &exp
		},
	}, {
		rawPattern: "*",
		exp: func(exp Section) *Section {
			exp.name = `*`
			exp.patterns = []*pattern{{
				value: "*",
			}}
			return &exp
		},
	}, {
		rawPattern: "192.168.1.?",
		exp: func(exp Section) *Section {
			exp.name = `192.168.1.?`
			exp.patterns = []*pattern{{
				value: `192.168.1.?`,
			}}
			return &exp
		},
	}, {
		rawPattern: "!*.co.uk *",
		exp: func(exp Section) *Section {
			exp.name = `!*.co.uk *`
			exp.patterns = []*pattern{{
				value:    `*.co.uk`,
				isNegate: true,
			}, {
				value: "*",
			}}
			return &exp
		},
	}}

	for _, c := range cases {
		got := newSectionHost(c.rawPattern)
		got.init(testParser.workDir, testParser.homeDir)

		exp := c.exp(*testDefaultSection)
		test.Assert(t, c.rawPattern, *exp, *got)
	}
}

func TestSection_init(t *testing.T) {
	cases := []struct {
		section func(def Section) *Section
		exp     func(def Section) *Section
	}{{
		section: func(def Section) *Section {
			return &def
		},
		exp: func(def Section) *Section {
			def.IdentityFile = []string{
				testParser.homeDir + "/.ssh/id_dsa",
				testParser.homeDir + "/.ssh/id_ecdsa",
				testParser.homeDir + "/.ssh/id_ed25519",
				testParser.homeDir + "/.ssh/id_rsa",
			}
			return &def
		},
	}}
	for _, c := range cases {
		got := c.section(*testDefaultSection)
		got.init(testParser.workDir, testParser.homeDir)

		exp := c.exp(*testDefaultSection)
		test.Assert(t, `init`, exp.IdentityFile, got.IdentityFile)
	}
}

func TestSection_setEnv(t *testing.T) {
	cfg := &Section{
		env: make(map[string]string),
	}
	cases := []struct {
		exp   map[string]string
		value string
	}{{
		value: "a",
		exp:   make(map[string]string),
	}, {
		value: "a=b",
		exp: map[string]string{
			"a": "b",
		},
	}}

	for _, c := range cases {
		cfg.setEnv(c.value)

		test.Assert(t, c.value, c.exp, cfg.env)
	}
}

func TestSection_Environments(t *testing.T) {
	envs := map[string]string{
		"key_1": "1",
		"key_2": "2",
		"key3":  "3",
	}

	cases := []struct {
		exp     map[string]string
		pattern string
	}{{
		pattern: "key_1",
		exp: map[string]string{
			"key_1": "1",
		},
	}, {
		pattern: "key_*",
		exp: map[string]string{
			"key_1": "1",
			"key_2": "2",
		},
	}}

	var (
		section = &Section{
			Field: map[string]string{},
			env:   map[string]string{},
		}
	)

	for _, c := range cases {
		section.sendEnv = nil
		_ = section.Set(KeySendEnv, c.pattern)
		got := section.Environments(envs)
		test.Assert(t, c.pattern, c.exp, got)
	}
}

func TestSection_UserKnownHostsFile(t *testing.T) {
	type testCase struct {
		value string
		exp   []string
	}

	var listCase = []testCase{{
		exp: defaultUserKnownHostsFile(),
	}, {
		value: `~/.ssh/myhost ~/.ssh/myhost2`,
		exp: []string{
			`~/.ssh/myhost`,
			`~/.ssh/myhost2`,
		},
	}}

	var (
		section = NewSection(`test`)

		c   testCase
		err error
	)
	for _, c = range listCase {
		err = section.Set(KeyUserKnownHostsFile, c.value)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.value, c.exp, section.UserKnownHostsFile())
	}
}
