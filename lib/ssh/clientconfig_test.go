// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestClientConfig_initialize(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		cfg    *ClientConfig
		exp    *ClientConfig
		expErr string
	}{{
		cfg:    &ClientConfig{},
		expErr: "ssh: remote user is not defined",
	}, {
		cfg: &ClientConfig{
			RemoteUser: "hodor",
		},
		expErr: "ssh: remote host is not defined",
	}, {
		cfg: &ClientConfig{
			PrivateKeyFile: "notexist",
		},
		expErr: `ssh: private key path "notexist" does not exist`,
	}, {
		cfg: &ClientConfig{
			RemoteUser: "hodor",
			RemoteHost: "127.0.0.1",
		},
		exp: &ClientConfig{
			WorkingDir:     wd,
			PrivateKeyFile: filepath.Join(userHomeDir, ".ssh", "id_rsa"),
			RemoteUser:     "hodor",
			RemoteHost:     "127.0.0.1",
			RemotePort:     22,
			remotePort:     "22",
			remoteAddr:     "127.0.0.1:22",
		},
	}}

	for _, c := range cases {
		err := c.cfg.initialize()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "exp", c.exp, c.cfg, true)
	}
}
