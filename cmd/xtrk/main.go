// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command xtrk is command line interface to uncompress and/or unarchive a
// file.
// Supported format: bzip2, gzip, tar, zip, tar.bz2, tar.gz.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	pakakeh "git.sr.ht/~shulhan/pakakeh.go"
	libos "git.sr.ht/~shulhan/pakakeh.go/lib/os"
)

const (
	cmdHelp    = "help"
	cmdVersion = "version"
)

func main() {
	log.SetPrefix("xtrk: ")
	log.SetFlags(0)

	flag.Parse()

	var (
		args = flag.Args()

		cmd    string
		fileIn string
		dirOut string
		err    error
	)

	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd = strings.ToLower(args[0])
	if cmd == cmdHelp {
		usage()
		return
	}
	if cmd == cmdVersion {
		fmt.Println(`xtrk v` + pakakeh.Version)
		return
	}

	switch len(args) {
	case 1:
		fileIn = args[0]

	case 2:
		fileIn = args[0]
		dirOut = args[1]
	}

	if len(dirOut) == 0 {
		dirOut, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	err = libos.Extract(fileIn, dirOut)
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Printf(`= xtrk

xtrk is command line interface to uncompress and/or unarchive a file.

== Synopsis

	xtrk <file> [dir]

== Description

xtrk accept single file to uncompress and/or archived into a directory output
"dir".
If directory output "dir" is not defined, it will be set to current directoy.

The compression and archive format is detected automatically based on the
following file input extension:

* .bz2: decompress using bzip2.
* .gz: decompress using gzip.
* .tar: unarchive using tar.
* .zip: unarchive using zip.
* .tar.bz2: decompress using bzip2 and unarchive using tar.
* .tar.gz: decompresss using gzip and unarchive using tar.

The input file will be removed on success.

== Examples

	$ xtrk file.gz

Extract file.gz into current directory.

	$ xtrk file.tar.bz2 /tmp

Extract file.tar.bz2 into directory /tmp.
`)
}
