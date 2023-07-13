// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

// List of known parameter name in header field's value.
const (
	// Parameter for Text Media Type, RFC 2046 section 4.1.
	ParamNameCharset = `charset`

	// Parameters for "application/octet-stream", RFC 2046 section 4.5.1.
	ParamNameType    = `type`
	ParamNamePadding = `padding`

	// Parameter for "multipart", RFC 2046 section 5.1.
	ParamNameBoundary = `boundary`

	// Parameters for "multipart/partial", RFC 2046 section 5.2.2.
	ParamNameID     = `id`
	ParamNameNumber = `number`
	ParamNameTotal  = `total`
)

// Param represent a mapping of key with its value.
type Param struct {
	Key    string
	Value  string
	Quoted bool // Quoted is true if value is contains special characters.
}
