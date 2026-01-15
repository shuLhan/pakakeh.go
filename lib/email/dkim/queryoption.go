// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

// QueryOption define an option for query.
type QueryOption byte

// List of valid and known query option.
const (
	QueryOptionTXT QueryOption = iota // "txt" (default)
)

// queryOptionNames contains a mapping between query option and their text
// representation.
var queryOptionNames = map[QueryOption][]byte{
	QueryOptionTXT: []byte("txt"),
}
