// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

import (
	"bytes"
)

// KeyType define a type of algorithm that sign the key.
type KeyType byte

// List of valid key types.
const (
	KeyTypeRSA KeyType = iota // "rsa" (default)
)

// keyTypeNames contains mapping between key type and their text
// representation.
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
