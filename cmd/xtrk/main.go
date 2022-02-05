package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	libos "github.com/shuLhan/share/lib/os"
)

func main() {
	log.SetPrefix("xtrk: ")

	log.SetFlags(0)

	flag.Parse()

	var (
		args   = flag.Args()
		fileIn string
		dirOut string
		err    error
	)

	switch len(args) {
	case 0:
		usage()
		os.Exit(1)

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
	fmt.Printf(`
= xtrk

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
