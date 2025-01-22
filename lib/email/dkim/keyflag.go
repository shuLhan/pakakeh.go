// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dkim

import (
	"bytes"
)

// KeyFlag define a type of key flag in DKIM key record.
type KeyFlag byte

// List of valid key flags.
const (
	// KeyFlagTesting or "y" in text, indicate that domain is for testing
	// DKIM.
	KeyFlagTesting KeyFlag = iota

	// KeyFlagStrict or "s" in text, means that the domain in AUID ("i=")
	// tag value MUST equal or subdomain of SDID "d=" tag value.
	KeyFlagStrict
)

// keyFlagNames contains mapping between key flag and their text
// representation.
var keyFlagNames = map[KeyFlag]byte{
	KeyFlagTesting: 'y',
	KeyFlagStrict:  's',
}

func unpackKeyFlags(in []byte) (out []KeyFlag) {
	flags := bytes.Split(in, sepColon)
	for x := range len(flags) {
		if len(flags[x]) != 1 {
			continue
		}
		for k, v := range keyFlagNames {
			if flags[x][0] == v {
				insertKeyFlag(&out, k)
				break
			}
		}
	}
	return out
}

func insertKeyFlag(flags *[]KeyFlag, key KeyFlag) {
	for _, v := range *flags {
		if v == key {
			return
		}
	}
	*flags = append(*flags, key)
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
