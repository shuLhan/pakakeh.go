// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// SignAlg define the type of signing and verification algorithm.
//
type SignAlg byte

//
// List of valid and known signing/verifying algorithms.
//
const (
	SignAlgRS256 SignAlg = iota // rsa-sha256 (default)
	SignAlgRS1                  // rsa-sha1
)

//
// signAlgNames contains mapping between known algorithm type and their names.
//
//nolint:gochecknoglobals
var signAlgNames = map[SignAlg][]byte{
	SignAlgRS256: []byte("rsa-sha256"),
	SignAlgRS1:   []byte("rsa-sha1"),
}
