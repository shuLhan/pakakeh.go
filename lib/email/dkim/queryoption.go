// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// QueryOption define an option for query.
//
type QueryOption byte

//
// List of valid and known query option.
//
const (
	QueryOptionTXT QueryOption = iota // "txt" (default)
)

//
// queryOptionNames contains a mapping between query option and their text
// representation.
//
//nolint:gochecknoglobals
var queryOptionNames = map[QueryOption][]byte{
	QueryOptionTXT: []byte("txt"),
}
