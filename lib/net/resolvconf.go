// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"bytes"

	"github.com/shuLhan/share/lib/file"
)

type ResolvConf struct {
	SearchList  []string
	NameServers []string
}

var (
	newLineTerms = []byte{'\n'}
	spaceSeps    = []byte{'\t', '\n', '\v', '\f', '\r', ' '}
)

//
// NewResolvConf open resolv.conf file in path and return the parsed records.
//
func NewResolvConf(path string) (*ResolvConf, error) {
	rc := &ResolvConf{}

	err := rc.parse(path)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

//
// parse open and parse the resolv.conf file.a
//
// Lines that contain a semicolon (;) or hash character (#) in the first
// column are treated as comments.
//
// The keyword and value must appear on a single line, and the keyword (e.g.,
// nameserver) must start the line.  The value follows the keyword, separated
// by white space.
//
// See `man resolv.conf`
//
func (rc *ResolvConf) parse(path string) error {
	reader, err := file.NewReader(path)
	if err != nil {
		return err
	}

	for {
		c := reader.SkipSpace()
		if c == 0 {
			break
		}
		if c == ';' || c == '#' {
			reader.SkipUntil(newLineTerms)
			continue
		}

		tok, isTerm := reader.ReadUntil(spaceSeps, newLineTerms)
		if isTerm {
			// We found keyword without value.
			continue
		}

		key := string(bytes.ToLower(tok))
		println("domain:", key)
		switch key {
		case "domain", "search":
			rc.parseValue(reader, &rc.SearchList)
		case "nameserver":
			rc.parseValue(reader, &rc.NameServers)
		default:
			reader.SkipUntil(newLineTerms)
		}
	}

	return nil
}

func (rc *ResolvConf) parseValue(reader *file.Reader, out *[]string) {
	for {
		c := reader.SkipHorizontalSpace()
		if c == '\n' || c == 0 {
			break
		}

		tok, isTerm := reader.ReadUntil(spaceSeps, newLineTerms)
		if len(tok) > 0 {
			*out = append(*out, string(tok))
		}
		if isTerm {
			break
		}
	}
}
