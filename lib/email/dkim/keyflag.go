// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"bytes"
)

//
// KeyFlag define a type of key flag in DKIM key record.
//
type KeyFlag byte

//
// List of valid key flags.
//
const (
	// KeyFlagTesting or "y" in text, indicate that domain is for testing
	// DKIM.
	KeyFlagTesting KeyFlag = iota

	// KeyFlagStrict or "s" in text, means that the domain in AUID ("i=")
	// tag value MUST equal or subdomain of SDID "d=" tag value.
	KeyFlagStrict
)

//
// keyFlagNames contains mapping between key flag and their text
// representation.
//
var keyFlagNames = map[KeyFlag]byte{ // nolint: gochecknoglobals
	KeyFlagTesting: 'y',
	KeyFlagStrict:  's',
}

func unpackKeyFlags(in []byte) (out []KeyFlag) {
	flags := bytes.Split(in, sepColon)
	for x := 0; x < len(flags); x++ {
		for k, v := range keyFlagNames {
			if flags[x][0] == v {
				out = append(out, k)
				break
			}
		}
	}
	return out
}

func packKeyFlags(flags []KeyFlag) []byte {
	var bb bytes.Buffer

	for _, flag := range flags {
		if bb.Len() > 0 {
			bb.Write(sepColon)
		}
		bb.WriteByte(keyFlagNames[flag])
	}

	return bb.Bytes()
}
