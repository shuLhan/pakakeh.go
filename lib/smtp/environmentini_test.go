// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewEnvironmentIni(t *testing.T) {
	osHostname, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc        string
		file        string
		expErr      string
		expHostname string
		expDomains  []string
	}{{
		desc:   "With file not exist",
		file:   "testdata/notexist",
		expErr: "NewEnvironmentIni: open testdata/notexist: no such file or directory",
	}, {
		desc:        "With empty hostname",
		file:        "testdata/smtpd.conf.empty-hostname",
		expHostname: osHostname,
		expDomains:  []string{osHostname},
	}, {
		desc:        "With empty domains",
		file:        "testdata/smtpd.conf.empty-domains",
		expHostname: "local",
		expDomains:  []string{"local"},
	}, {
		desc:        "With duplicate domain",
		file:        "testdata/smtpd.conf.duplicate",
		expHostname: "local",
		expDomains:  []string{"a", "b", "c", "local"},
	}, {
		desc:   "With invalid hostname",
		file:   "testdata/smtpd.conf.invalid-hostname",
		expErr: "EnvironmentIni: invalid hostname 'what.'",
	}, {
		desc:   "With invalid domain",
		file:   "testdata/smtpd.conf.invalid-domain",
		expErr: "EnvironmentIni: invalid domain ''",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := NewEnvironmentIni(c.file)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "hostname", c.expHostname, got.Hostname(), true)
		test.Assert(t, "domains", c.expDomains, got.Domains(), true)
	}
}
