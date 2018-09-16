// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

//
// MetadataInterface is the interface for field metadata.
// This is to make anyone can extend the DSV library including the metadata.
//
type MetadataInterface interface {
	Init()
	GetName() string
	GetType() int
	GetTypeName() string
	GetLeftQuote() string
	GetRightQuote() string
	GetSeparator() string
	GetSkip() bool
	GetValueSpace() []string

	IsEqual(MetadataInterface) bool
}

//
// FindMetadata Given a slice of metadata, find `mdin` in the slice which has the
// same name, ignoring metadata where Skip value is true.
// If found, return the index and metadata object of matched metadata name.
// If not found return -1 as index and nil in `mdout`.
//
func FindMetadata(mdin MetadataInterface, mds []MetadataInterface) (
	idx int,
	mdout MetadataInterface,
) {
	for _, md := range mds {
		if md.GetName() == mdin.GetName() {
			mdout = md
			break
		}
		if !md.GetSkip() {
			idx++
		}
	}
	return idx, mdout
}
