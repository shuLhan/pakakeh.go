// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

// SignAlg define the type of signing and verification algorithm.
type SignAlg byte

// List of valid and known signing/verifying algorithms.
const (
	SignAlgRS256 SignAlg = iota // rsa-sha256 (default)
	SignAlgRS1                  // rsa-sha1
)

// signAlgNames contains mapping between known algorithm type and their names.
var signAlgNames = map[SignAlg][]byte{
	SignAlgRS256: []byte("rsa-sha256"),
	SignAlgRS1:   []byte("rsa-sha1"),
}
