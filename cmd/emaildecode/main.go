// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Program emaildecode convert the email body from quoted-printable to plain
// text.
// Usage,
//
//	emaildecode <file>
//
// The emaildecode accept single file as input, read its content, decode the
// body based on Content-Transfer-Encoding, and then print each body to
// standard output.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.sr.ht/~shulhan/pakakeh.go"
	"git.sr.ht/~shulhan/pakakeh.go/lib/email"
)

var usage = `= emaildecode - CLI to decode email body

== USAGE

	emaildecode <COMMAND | FILE>

List of accepted COMMANDS,

	help - print the usage.

FILE
	The file that contains email to be decoded.

== INFO

Version: ` + pakakeh.Version + `
Website: https://sr.ht/~shulhan/pakakeh.go/
`

const (
	cmdHelp = `help`
)

func main() {
	flag.Parse()

	var fileInput = flag.Arg(0)
	if len(fileInput) == 0 {
		log.Fatalf(`missing file input`)
	}

	if fileInput == cmdHelp {
		fmt.Println(usage)
		os.Exit(0)
	}

	var (
		msg *email.Message
		err error
	)

	msg, _, err = email.ParseFile(fileInput)
	if err != nil {
		log.Fatal(err)
	}

	var (
		mime  *email.MIME
		field *email.Field
		x     int
	)
	for _, field = range msg.Header.Fields {
		fmt.Printf(`%s: %s`, field.Name, field.Value)
	}
	for x, mime = range msg.Body.Parts {
		fmt.Printf("\n-- mime #%d\n\n", x)
		if mime.Header != nil {
			for _, field = range mime.Header.Fields {
				fmt.Printf(`%s: %s`, field.Name, field.Value)
			}
			fmt.Println(``)
		}
		fmt.Println(string(mime.Content))
	}
}
