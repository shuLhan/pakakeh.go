// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"testing"

	"github.com/shuLhan/share/lib/ascii"
	"github.com/shuLhan/share/lib/test"
)

func TestNewResolvConf(t *testing.T) {
	cases := []struct {
		desc   string
		path   string
		exp    *ResolvConf
		expErr string
	}{{
		desc:   "With invalid file",
		path:   "",
		exp:    nil,
		expErr: "open : no such file or directory",
	}, {
		desc: "From testdata/resolv.conf",
		path: "testdata/resolv.conf",
		exp: &ResolvConf{
			Domain:      "a",
			Search:      []string{"d", "e", "f", "g", "h", "i"},
			NameServers: []string{"127.0.0.1", "1.1.1.1", "2.2.2.2"},
			NDots:       1,
			Timeout:     5,
			Attempts:    2,
			OptMisc:     make(map[string]bool),
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		rc, err := NewResolvConf(c.path)
		if err != nil {
			test.Assert(t, "expError", c.expErr, err.Error(), true)
		}

		test.Assert(t, "exp", c.exp, rc, true)
	}
}

func TestResolvConf_Init(t *testing.T) {
	getHostname = func() (string, error) {
		return "kilabit.info", nil
	}

	veryLongName := string(ascii.Random([]byte(ascii.Letters), 255))

	cases := []struct {
		desc           string
		src            string
		envLocaldomain string
		exp            *ResolvConf
	}{{
		desc: "Duplicate domain",
		src: `
domain a
domain b
`,
		exp: &ResolvConf{
			Domain:   "b",
			NDots:    1,
			Timeout:  5,
			Attempts: 2,
			OptMisc:  make(map[string]bool),
		},
	}, {
		desc: "Without domain",
		src: `
search a b c
`,
		exp: &ResolvConf{
			Domain:   "info",
			Search:   []string{"a", "b", "c"},
			NDots:    1,
			Timeout:  5,
			Attempts: 2,
			OptMisc:  make(map[string]bool),
		},
	}, {
		desc: "Duplicate search",
		src: `
search q w e r t y u
search a b c d e f g
`,
		exp: &ResolvConf{
			Domain:   "info",
			Search:   []string{"a", "b", "c", "d", "e", "f"},
			NDots:    1,
			Timeout:  5,
			Attempts: 2,
			OptMisc:  make(map[string]bool),
		},
	}, {
		desc: "More than 3 nameservers",
		src: `
nameserver 1
nameserver 2
nameserver 3
nameserver 4
`,
		exp: &ResolvConf{
			Domain:      "info",
			NameServers: []string{"1", "2", "3"},
			NDots:       1,
			Timeout:     5,
			Attempts:    2,
			OptMisc:     make(map[string]bool),
		},
	}, {
		desc: "A very long search domain > 255 chars",
		src:  `search aaaaa bbbbb ccccc ddddd ` + veryLongName,
		exp: &ResolvConf{
			Domain:   "info",
			Search:   []string{"aaaaa", "bbbbb", "ccccc", "ddddd"},
			NDots:    1,
			Timeout:  5,
			Attempts: 2,
			OptMisc:  make(map[string]bool),
		},
	}, {
		desc:           "Overriding search with env",
		envLocaldomain: "a b c d e f g h",
		exp: &ResolvConf{
			Domain:   "info",
			Search:   []string{"a", "b", "c", "d", "e", "f"},
			NDots:    1,
			Timeout:  5,
			Attempts: 2,
			OptMisc:  make(map[string]bool),
		},
	}, {
		desc: "Single line options",
		src:  `options debug ndots:3 timeout:31 attempts:44`,
		exp: &ResolvConf{
			Domain:   "info",
			NDots:    3,
			Timeout:  30,
			Attempts: 5,
			OptMisc: map[string]bool{
				"debug": true,
			},
		},
	}, {
		desc: "Multi lines options",
		src: `
options debug
options ndots:3
options timeout:31
options attempts:44
options ndots:33
`,
		exp: &ResolvConf{
			Domain:   "info",
			NDots:    15,
			Timeout:  30,
			Attempts: 5,
			OptMisc: map[string]bool{
				"debug": true,
			},
		},
	}}

	rc := new(ResolvConf)

	for _, c := range cases {
		t.Log(c.desc)

		if len(c.envLocaldomain) > 0 {
			err := os.Setenv(envLocaldomain, c.envLocaldomain)
			if err != nil {
				t.Fatal(err)
			}
		}

		rc.Init(c.src)

		if len(c.envLocaldomain) > 0 {
			err := os.Unsetenv(envLocaldomain)
			if err != nil {
				t.Fatal(err)
			}
		}

		test.Assert(t, "ResolvConf", c.exp, rc, true)
	}
}
