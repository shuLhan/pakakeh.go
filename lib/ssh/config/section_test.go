// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"reflect"
	"testing"
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
			exp.patterns = []*pattern{{
				value: "*",
			}}
			return &exp
		},
	}, {
		rawPattern: "192.168.1.?",
		exp: func(exp Section) *Section {
			exp.patterns = []*pattern{{
				value: `192.168.1.?`,
			}}
			return &exp
		},
	}, {
		rawPattern: "!*.co.uk *",
		exp: func(exp Section) *Section {
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
		if !reflect.DeepEqual(*exp, *got) {
			t.Fatalf("newSectionHost: expecting %v, got %v", exp, got)
		}
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
		if !reflect.DeepEqual(exp.IdentityFile, got.IdentityFile) {
			t.Fatalf("init: expecting %v, got %v", exp, got)
		}
	}
}

func TestSection_setEnv(t *testing.T) {
	cfg := &Section{
		Environments: make(map[string]string),
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

		if !reflect.DeepEqual(c.exp, cfg.Environments) {
			t.Fatalf("setEnv: expecting %v, got %v", c.exp, cfg.Environments)
		}
	}
}

func TestSection_setSendEnv(t *testing.T) {
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

	cfg := &Section{
		Environments: make(map[string]string),
	}

	for _, c := range cases {
		cfg.setSendEnv(envs, c.pattern)
		if !reflect.DeepEqual(c.exp, cfg.Environments) {
			t.Fatalf("setSendEnv: %s: expecting %v, got %v", c.pattern, c.exp, cfg.Environments)
		}
	}
}
