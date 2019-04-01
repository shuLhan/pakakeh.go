// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command smtpcli provide a command line interface to SMTP client protocol.
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		cli    *client
		err    error
		isQuit bool
	)

	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cli, err = newClient(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	for {
		cli.writePrompt()

		err = cli.readInput()
		if err != nil {
			log.Fatal(err)
		}

		isQuit = cli.handleInput()
		if isQuit {
			return
		}
	}
}

func usage() {
	fmt.Println("smtpcli [(smtp | smtps)://](domain | ip-address [:port])")
}
