// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package dns

// RecordClass represent the class for resource record.
type RecordClass uint16

// List of known record class, ordered by value.
const (
	RecordClassZERO RecordClass = iota // Empty query class.
	RecordClassIN                      // The Internet
	RecordClassCH                      // The CHAOS class
	RecordClassHS                      // Hesiod [Dyer 87]

	RecordClassANY RecordClass = 255 // Any class
)

// RecordClasses contains a mapping between string representation of record
// class to their numeric value, ordered by key alphabetically.
var RecordClasses = map[string]RecordClass{
	"CH": RecordClassCH,
	"HS": RecordClassHS,
	"IN": RecordClassIN,
}

// RecordClassName contains a mapping between the record class value and its
// string representation, ordered by key alphabetically.
var RecordClassName = map[RecordClass]string{
	RecordClassCH: "CH",
	RecordClassHS: "HS",
	RecordClassIN: "IN",
}
