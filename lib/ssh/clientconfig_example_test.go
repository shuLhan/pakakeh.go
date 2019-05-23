// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"log"
)

func ExampleClientConfig() {
	cfg := &ClientConfig{
		WorkingDir:     "/tmp",
		PrivateKeyFile: "testdata/example.pem",
		RemoteUser:     "hodor",
		RemoteHost:     "127.0.0.1",
	}

	err := cfg.initialize()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("WorkingDir: %s\n", cfg.WorkingDir)
	fmt.Printf("PrivateKeyFile: %s\n", cfg.PrivateKeyFile)
	fmt.Printf("RemoteUser: %s\n", cfg.RemoteUser)
	fmt.Printf("RemoteHost: %s\n", cfg.RemoteHost)
	fmt.Printf("RemotePort: %d\n", cfg.RemotePort)
	fmt.Printf("remoteAddr: %s\n", cfg.remoteAddr)
	// Output:
	// WorkingDir: /tmp
	// PrivateKeyFile: testdata/example.pem
	// RemoteUser: hodor
	// RemoteHost: 127.0.0.1
	// RemotePort: 22
	// remoteAddr: 127.0.0.1:22
}
