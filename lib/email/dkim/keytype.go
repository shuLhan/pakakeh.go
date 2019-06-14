// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"bytes"
)

//
// KeyType define a type of algorithm that sign the key.
//
type KeyType byte

//
// List of valid key types.
//
const (
	KeyTypeRSA KeyType = iota // "rsa" (default)
)

//
// keyTypeNames contains mapping between key type and their text
// representation.
//
//nolint:gochecknoglobals
var keyTypeNames = map[KeyType][]byte{
	KeyTypeRSA: []byte("rsa"),
}

func parseKeyType(in []byte) (t *KeyType) {
	for k, name := range keyTypeNames {
		if bytes.Equal(in, name) {
			k := k
			return &k
		}
	}
	return nil
}
