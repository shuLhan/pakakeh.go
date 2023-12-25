// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/ssh/config"
	"github.com/shuLhan/share/lib/test"
)

// TestNewClient_KeyError test SSH to server with host key does not exist in
// known_hosts database.
func TestNewClient_KeyError_notExist(t *testing.T) {
	t.Skip(`Require active SSH server`)

	var (
		section = config.NewSection(nil, `localhost`)

		wd  string
		err error
	)

	wd, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	err = section.Set(config.KeyUser, `ms`)
	if err != nil {
		t.Fatal(err)
	}
	err = section.Set(config.KeyHostname, `localhost`)
	if err != nil {
		t.Fatal(err)
	}

	var knownHostsFile = filepath.Join(wd, `testdata/localhost/known_hosts_empty`)
	err = section.Set(config.KeyUserKnownHostsFile, knownHostsFile)
	if err != nil {
		t.Fatal(err)
	}

	var pkeyFile = filepath.Join(wd, `testdata/localhost/client.key`)
	err = section.Set(config.KeyIdentityFile, pkeyFile)
	if err != nil {
		t.Fatal(err)
	}

	var (
		expError = fmt.Sprintf(`NewClientInteractive: dialWithSigners: ssh: handshake failed: knownhosts: key is unknown from known_hosts files [%s]`, knownHostsFile)
		gotError string
	)

	_, err = NewClientInteractive(section)
	if err != nil {
		gotError = err.Error()
	}
	test.Assert(t, `NewClientInteractive: error`, expError, gotError)
}
