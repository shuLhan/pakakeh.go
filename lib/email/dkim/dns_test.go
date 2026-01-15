// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

import (
	"strings"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewDNSClientPool(t *testing.T) {
	cases := []struct {
		desc         string
		expErr       string
		expErrLookup string
		ns           []string
	}{{
		desc: "With empty DefaultNameServers",
		ns:   []string{},
	}, {
		desc: "With invalid IP address on DefaultNameServers",
		ns: []string{
			"invalid.ip",
		},
		expErr: "invalid host address",
	}, {
		desc: "With invalid name server",
		ns: []string{
			"127.0.0.1:5353",
		},
		expErr: "timeout",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		dnsClientPool = nil
		DefaultNameServers = c.ns

		_, err := LookupKey(QueryMethod{}, "example.com")
		if err != nil {
			if strings.Contains(err.Error(), c.expErrLookup) {
				continue
			}
			test.Assert(t, "error lookup", c.expErrLookup, err.Error())
		}
	}

	DefaultNameServers = nil
	dnsClientPool = nil
}
