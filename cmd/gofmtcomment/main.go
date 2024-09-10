// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Command gofmtcomment is a program to convert "/**/" comment into "//".
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	pakakeh "git.sr.ht/~shulhan/pakakeh.go"
)

const (
	gapb = byte(' ')
)

var (
	twoGaps   = []byte{gapb, gapb}
	threeGaps = []byte{gapb, gapb, gapb}
)

func usage() {
	fmt.Println(`= gofmtcomment

gofmtcomment is a program to convert multi line comments from "/**/" into
"//".

== SYNOPSIS

	gofmtcomment <file>

== ARGUMENTS

<file>
	Path to file where comment to be read and replaced.

== EXAMPLE

	$ gofmtcomment main.go`)
}

const (
	cmdHelp    = "help"
	cmdVersion = "version"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gofmtcomment: ")

	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	var (
		cmd = strings.ToLower(os.Args[1])
	)

	if cmd == cmdHelp {
		usage()
		return
	}
	if cmd == cmdVersion {
		fmt.Println(`gofmtcomment v` + pakakeh.Version)
		return
	}

	var (
		f   *os.File
		re  *regexp.Regexp
		in  []byte
		err error
	)

	log.Printf("Reformat comment on %s\n", os.Args[1])

	f, err = os.OpenFile(os.Args[1], os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	in, err = io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	re = regexp.MustCompile(`(?Us)/\*.+\*/`)

	var (
		nshift         int
		nreplace       int
		startIdx       int
		endIdx         int
		l              int
		y              int
		matchIdx       []int
		matchb         []byte
		doubleNewlines bool
	)

	for {
		matchIdx = re.FindIndex(in)
		if matchIdx == nil {
			break
		}

		nreplace++
		startIdx = matchIdx[0]
		endIdx = matchIdx[1]

		matchb = in[startIdx:endIdx]
		l = len(matchb)

		matchb[0] = '/'
		matchb[1] = '/'
		matchb[l-2] = '/'
		matchb[l-1] = '/'

		l -= 3
		for y = 2; y < l; y++ {
			if matchb[y] != '\n' {
				continue
			}

			y++

			if matchb[y] == ' ' || matchb[y] == '\t' ||
				matchb[y] == '\n' {
				nshift = 2
				in = append(in, twoGaps...)
				if matchb[y] == '\n' {
					doubleNewlines = true
				}
			} else {
				nshift = 3
				in = append(in, threeGaps...)
			}

			copy(in[startIdx+y+nshift:], in[startIdx+y:])
			in[startIdx+y] = '/'
			in[startIdx+y+1] = '/'

			if !doubleNewlines {
				in[startIdx+y+2] = gapb
				y += 2
			} else {
				doubleNewlines = false
			}

			l += nshift
			endIdx += nshift
			matchb = in[startIdx:endIdx]
		}
	}

	_, err = f.WriteAt(in, 0)
	if err != nil {
		log.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}

	if nreplace > 0 {
		log.Printf(">>> Replacing %d comment blocks\n", nreplace)
	} else {
		log.Println(">>> Nothing changes")
	}
}
