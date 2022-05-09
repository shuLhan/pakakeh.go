// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	for x := 0; x < len(algs); x++ {
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
