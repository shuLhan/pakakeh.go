// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

//
// ClientConfig contains the configuration to create SSH connection and the
// environment where the SSH program started.
//
type ClientConfig struct {
	// Environments contains system environment variables that will be
	// passed to Execute().
	Environments map[string]string

	// WorkingDir contains the directory where the SSH client started.
	// This value is required when client want to copy file from/to
	// remote.
	// This field is optional, default to current working directory from
	// os.Getwd() or user's home directory.
	WorkingDir string

	// PrivateKeyFile contains path to private key file.
	// This field is optional, default to ".ssh/id_rsa" in user's home
	// directory.
	PrivateKeyFile string

	// RemoteUser contains the user name in remote system.
	// This field is mandatory.
	RemoteUser string

	// RemoteHost contains the IP address or host name of remote system.
	// This field is mandatory.
	RemoteHost string

	// RemotePort contains the port address of remote SSH server.
	// This field is optional, default to 22.
	RemotePort int

	remotePort string
	remoteAddr string
}

//
// initialize the ClientConfig's fields with default value.
//
func (cc *ClientConfig) initialize() (err error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ssh: cannot get user's home directory: " + err.Error())
	}

	if len(cc.WorkingDir) == 0 {
		cc.WorkingDir, err = os.Getwd()
		if err != nil {
			log.Println("ssh: cannot get working directory, default to user's home")
			cc.WorkingDir = userHomeDir
		}
	}

	if len(cc.PrivateKeyFile) == 0 {
		cc.PrivateKeyFile = filepath.Join(userHomeDir, ".ssh", "id_rsa")
	}

	_, err = os.Stat(cc.PrivateKeyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ssh: private key path %q does not exist", cc.PrivateKeyFile)
		}
		return fmt.Errorf("ssh: os.Stat %q: %s", cc.PrivateKeyFile, err)
	}

	if len(cc.RemoteUser) == 0 {
		return fmt.Errorf("ssh: remote user is not defined")
	}
	if len(cc.RemoteHost) == 0 {
		return fmt.Errorf("ssh: remote host is not defined")
	}
	if cc.RemotePort <= 0 || cc.RemotePort >= 65535 {
		log.Printf("ssh: using default port instead of %d\n", cc.RemotePort)
		cc.RemotePort = 22
	}

	cc.remotePort = strconv.Itoa(cc.RemotePort)
	cc.remoteAddr = fmt.Sprintf("%s:%d", cc.RemoteHost, cc.RemotePort)

	return nil
}
