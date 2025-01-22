// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dkim

import (
	"bytes"
)

// HashAlg define the type for hash algorithm.
type HashAlg byte

// List of valid and known hash algorithms.
const (
	HashAlgALL    HashAlg = iota // (default to allow all)
	HashAlgSHA256                // sha256
	HashAlgSHA1                  // sha1
)

// hashAlgNames contains mapping between type value and their names.
var hashAlgNames = map[HashAlg][]byte{
	HashAlgSHA256: []byte("sha256"),
	HashAlgSHA1:   []byte("sha1"),
}

func unpackHashAlgs(v []byte) (hashAlgs []HashAlg) {
	algs := bytes.Split(v, sepColon)
	for x := range len(algs) {
		for k, v := range hashAlgNames {
			if bytes.Equal(v, algs[x]) {
				hashAlgs = append(hashAlgs, k)
				break
			}
		}
	}
	if len(hashAlgs) == 0 {
		hashAlgs = append(hashAlgs, HashAlgALL)
	}

	return
}

func packHashAlgs(hashAlgs []HashAlg) []byte {
	var bb bytes.Buffer

	for _, v := range hashAlgs {
		if v == HashAlgALL {
			return nil
		}
		if bb.Len() > 0 {
			bb.Write(sepColon)
		}
		bb.Write(hashAlgNames[v])
	}

	return bb.Bytes()
}
