// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command ini provide a command line interface to get and set values in the
// [INI file format].
//
// [INI file format]: https://godocs.io/git.sr.ht/~shulhan/pakakeh.go/lib/ini
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go"
	"git.sr.ht/~shulhan/pakakeh.go/lib/ini"
)

const usage = `= ini

A command line interface to get and set values in the INI file format [1].

== SYNOPSIS

     ini <get | help | set | version> <TAGS> [VALUE] <FILE | "-">

== ARGUMENTS

get
	Print the key's value to the stdout.
	If the key not found it will print empty line.

help
	Print the command usage.

set
	Changes the key's value in the INI file to VALUE.

version
	Print the current command version.

<TAGS>
	Tag is combination of section, subsection, and key string separated
	by colon ":", using the following format,

		SECTION ":" SUBSECTION ":" KEY ":" DEFAULT

	At least one of the tag should be defined.

<VALUE>
	New value for key when calling set.

<FILE | "-">
	Path to the INI file.
	If set to "-" and the command is "get", it will read the INI content
	from standard input.

== EXAMPLES

Get key's value,

	$ ini get "user::name" lib/ini/testdata/input.ini
	Shulhan

	$ cat lib/ini/testdata/input.ini | ini get "user::name" -
	Shulhan

Set key's value,

	$ ini set "user::name" "my name" lib/ini/testdata/input.ini

	$ ini get "user::name" lib/ini/testdata/input.ini
	my name

== REFERENCES

[1] https://godocs.io/git.sr.ht/~shulhan/pakakeh.go/lib/ini`

const (
	cmdGet     = "get"
	cmdHelp    = "help"
	cmdSet     = "set"
	cmdVersion = "version"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("ini: ")

	flag.Parse()

	var (
		args = flag.Args()

		cmd string
	)

	if len(args) == 0 {
		log.Fatal(usage)
	}

	cmd = strings.ToLower(args[0])

	switch cmd {
	case cmdGet:
		doGet(args[1:])

	case cmdHelp:
		fmt.Println(usage)

	case cmdSet:
		doSet(args[1:])

	case cmdVersion:
		fmt.Println("ini v" + share.Version)

	default:
		log.Println("ini: unknown command:", cmd)
	}
}

func doGet(args []string) {
	var (
		cfg  *ini.Ini
		tags []string
		raw  []byte
		vstr string
		err  error
	)

	if len(args) < 2 {
		log.Fatalf("%s: missing arguments", cmdGet)
	}

	tags = ini.ParseTag(args[0])

	vstr = args[1]
	if vstr == "-" {
		raw, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("%s: %s", cmdGet, err)
		}

		cfg, err = ini.Parse(raw)
	} else {
		cfg, err = ini.Open(vstr)
	}
	if err != nil {
		log.Fatalf("%s: %s", cmdGet, err)
	}

	vstr, _ = cfg.Get(tags[0], tags[1], tags[2], tags[3])
	fmt.Println(vstr)
}

func doSet(args []string) {
	var (
		cfg  *ini.Ini
		tags []string
		err  error
	)

	if len(args) < 3 {
		log.Fatalf("%s: missing arguments", cmdSet)
	}

	tags = ini.ParseTag(args[0])

	cfg, err = ini.Open(args[2])
	if err != nil {
		log.Fatalf("%s: %s", cmdSet, err)
	}

	cfg.Set(tags[0], tags[1], tags[2], args[1])

	err = cfg.Save(args[2])
	if err != nil {
		log.Fatalf("%s: %s", cmdSet, err)
	}
}
