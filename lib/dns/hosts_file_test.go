// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParseHostsFile(t *testing.T) {
	var (
		hostsFile *HostsFile
		err       error
	)

	hostsFile, err = ParseHostsFile("testdata/hosts")
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Length", 10, len(hostsFile.Records))
}

func TestHostsLoad2(t *testing.T) {
	var (
		err error
	)

	_, err = ParseHostsFile("testdata/hosts.block")
	if err != nil {
		t.Fatal(err)
	}
}
