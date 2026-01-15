// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

// QueryType define type for query.
type QueryType byte

// List of valid and known query type.
const (
	QueryTypeDNS QueryType = iota // "dns" (default)
)

// queryTypeNames contains a mapping betweend query type and their text
// representation.
var queryTypeNames = map[QueryType][]byte{
	QueryTypeDNS: []byte("dns"),
}
