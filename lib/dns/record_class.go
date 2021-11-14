// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

type RecordClass uint16

// List of known record class, ordered by value.
const (
	RecordClassZERO RecordClass = iota // Empty query class.
	RecordClassIN                      // The Internet
	RecordClassCH                      // The CHAOS class
	RecordClassHS                      // Hesiod [Dyer 87]

	RecordClassANY RecordClass = 255 // Any class
)

//
// RecordClasses contains a mapping between string representation of record
// class to their numeric value, ordered by key alphabetically.
//
var RecordClasses = map[string]RecordClass{
	"CH": RecordClassCH,
	"HS": RecordClassHS,
	"IN": RecordClassIN,
}

//
// RecordClassName contains a mapping between the record class value and its
// string representation, ordered by key alphabetically.
//
var RecordClassName = map[RecordClass]string{
	RecordClassCH: "CH",
	RecordClassHS: "HS",
	RecordClassIN: "IN",
}
