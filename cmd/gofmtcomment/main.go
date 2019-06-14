// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Command gofmtcomment is a program to convert "/**/" comment into "//".
//
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

const (
	gapb = byte(' ')
)

//nolint:gochecknoglobals
var (
	twoGaps   = []byte{gapb, gapb}
	threeGaps = []byte{gapb, gapb, gapb}
)

func usage() {
	fmt.Printf("%s: <file>\n", os.Args[0])
}

func main() {
	if len(os.Args) != 2 {
		usage()
		return
	}

	log.Printf("Reformat comment on %s\n", os.Args[1])

	f, err := os.OpenFile(os.Args[1], os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	in, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`(?Us)/\*.+\*/`)

	var nshift int
	doubleNewlines := false
	nreplace := 0

	for {
		matchIdx := re.FindIndex(in)
		if matchIdx == nil {
			break
		}

		nreplace++
		startIdx := matchIdx[0]
		endIdx := matchIdx[1]

		matchb := in[startIdx:endIdx]
		l := len(matchb)

		matchb[0] = '/'
		matchb[1] = '/'
		matchb[l-2] = '/'
		matchb[l-1] = '/'

		l -= 3
		for y := 2; y < l; y++ {
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
