// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// QueryType define type for query.
//
type QueryType byte

//
// List of valid and known query type.
//
const (
	QueryTypeDNS QueryType = iota // "dns" (default)
)

//
// queryTypeNames contains a mapping betweend query type and their text
// representation.
//
var queryTypeNames = map[QueryType][]byte{
	QueryTypeDNS: []byte("dns"),
}
