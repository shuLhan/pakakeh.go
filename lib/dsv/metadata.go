// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.

package dsv

import (
	"encoding/json"
	"log"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

// Metadata represent on how to parse each column in record.
type Metadata struct {
	// Name of the column, optional.
	Name string `json:"Name"`

	// Type of the column, default to "string".
	// Valid value are: "string", "integer", "real"
	Type string `json:"Type"`

	// Separator for column in record.
	Separator string `json:"Separator"`

	// LeftQuote define the characters that enclosed the column in the left
	// side.
	LeftQuote string `json:"LeftQuote"`

	// RightQuote define the characters that enclosed the column in the
	// right side.
	RightQuote string `json:"RightQuote"`

	// ValueSpace contain the possible value in records
	ValueSpace []string `json:"ValueSpace"`

	// T type of column in integer.
	T int `json:"T"`

	// Skip, if its true this column will be ignored, not saved in reader
	// object. Default to false.
	Skip bool `json:"Skip"`
}

// NewMetadata create and return new metadata.
func NewMetadata(name, tipe, sep, leftq, rightq string, vs []string) (
	md *Metadata,
) {
	md = &Metadata{
		Name:       name,
		Type:       tipe,
		Separator:  sep,
		LeftQuote:  leftq,
		RightQuote: rightq,
		ValueSpace: vs,
	}

	md.Init()

	return
}

// Init initialize metadata column, i.e. check and set column type.
//
// If type is unknown it will default to string.
func (md *Metadata) Init() {
	switch strings.ToUpper(md.Type) {
	case "INTEGER", "INT":
		md.T = tabula.TInteger
	case "REAL":
		md.T = tabula.TReal
	default:
		md.T = tabula.TString
		md.Type = "string"
	}
}

// GetName return the name of metadata.
func (md *Metadata) GetName() string {
	return md.Name
}

// GetType return type of metadata.
func (md *Metadata) GetType() int {
	return md.T
}

// GetTypeName return string representation of type.
func (md *Metadata) GetTypeName() string {
	return md.Type
}

// GetSeparator return the field separator.
func (md *Metadata) GetSeparator() string {
	return md.Separator
}

// GetLeftQuote return the string used in the beginning of record value.
func (md *Metadata) GetLeftQuote() string {
	return md.LeftQuote
}

// GetRightQuote return string that end in record value.
func (md *Metadata) GetRightQuote() string {
	return md.RightQuote
}

// GetSkip return number of rows that will be skipped when reading data.
func (md *Metadata) GetSkip() bool {
	return md.Skip
}

// GetValueSpace return value space.
func (md *Metadata) GetValueSpace() []string {
	return md.ValueSpace
}

// IsEqual return true if this metadata equal with other instance, return false
// otherwise.
func (md *Metadata) IsEqual(o MetadataInterface) bool {
	if md.Name != o.GetName() {
		return false
	}
	if md.Separator != o.GetSeparator() {
		return false
	}
	if md.LeftQuote != o.GetLeftQuote() {
		return false
	}
	if md.RightQuote != o.GetRightQuote() {
		return false
	}
	return true
}

// String yes, it will print it JSON like format.
func (md *Metadata) String() string {
	r, e := json.MarshalIndent(md, "", "\t")
	if nil != e {
		log.Print(e)
	}
	return string(r)
}
