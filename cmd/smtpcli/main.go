// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Command smtpcli provide a command line interface to SMTP client protocol.
// This is an example of implementation Client from
// [lib/smtp].
//
// [lib/smtp]: https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/smtp
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
