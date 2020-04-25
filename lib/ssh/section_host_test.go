// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewSectionHost(t *testing.T) {
	cases := []struct {
		rawPattern string
		exp        func(def ConfigSection) *ConfigSection
		expError   string
	}{{
		rawPattern: "",
		exp: func(exp ConfigSection) *ConfigSection {
			exp.patterns = make([]*configPattern, 0)
			return &exp
		},
	}, {
		rawPattern: "*",
		exp: func(exp ConfigSection) *ConfigSection {
			exp.patterns = []*configPattern{{
				pattern: "*",
			}}
			return &exp
		},
	}, {
		rawPattern: "192.168.1.?",
		exp: func(exp ConfigSection) *ConfigSection {
			exp.patterns = []*configPattern{{
				pattern: `192.168.1.?`,
			}}
			return &exp
		},
	}, {
		rawPattern: "!*.co.uk *",
		exp: func(exp ConfigSection) *ConfigSection {
			exp.patterns = []*configPattern{{
				pattern:  `*.co.uk`,
				isNegate: true,
			}, {
				pattern: "*",
			}}
			return &exp
		},
	}}

	for _, c := range cases {
		got, err := newSectionHost(c.rawPattern)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}
		got.postConfig(testParser.homeDir)
		test.Assert(t, "newHost", c.exp(*testDefaultSection), got, true)
	}
}
