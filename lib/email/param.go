// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

//
// List of known parameter name in header field's value.
//
var (
	// Parameter for Text Media Type, RFC 2046 section 4.1.
	ParamNameCharset = []byte("charset")

	// Parameters for "application/octet-stream", RFC 2046 section 4.5.1.
	ParamNameType    = []byte("type")
	ParamNamePadding = []byte("padding")

	// Parameter for "multipart", RFC 2046 section 5.1.
	ParamNameBoundary = []byte("boundary")

	// PArameters for "mulitpart/partial", RFC 2046 section 5.2.2.
	ParamNameID     = []byte("id")
	ParamNameNumber = []byte("number")
	ParamNameTotal  = []byte("total")
)

//
// Param represent a key-value in slice of bytes.
//
type Param struct {
	Key    []byte
	Value  []byte
	Quoted bool // Quoted is true if value is contains special characters.
}
