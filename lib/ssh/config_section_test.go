// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestConfigSection_postConfig(t *testing.T) {
	cases := []struct {
		section func(def ConfigSection) *ConfigSection
		exp     func(def ConfigSection) *ConfigSection
	}{{
		section: func(def ConfigSection) *ConfigSection {
			return &def
		},
		exp: func(def ConfigSection) *ConfigSection {
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
		got.postConfig(testParser)
		test.Assert(t, "postConfig", c.exp(*testDefaultSection), got, true)
	}
}

func TestConfigSection_setSendEnv(t *testing.T) {
	envs := map[string]string{
		"key_1": "1",
		"key_2": "2",
		"key3":  "3",
	}

	cases := []struct {
		pattern string
		exp     map[string]string
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

	cfg := &ConfigSection{
		Environments: make(map[string]string),
	}

	for _, c := range cases {
		cfg.setSendEnv(envs, c.pattern)

		test.Assert(t, "setSendEnv: "+c.pattern,
			c.exp, cfg.Environments, true)
	}
}
