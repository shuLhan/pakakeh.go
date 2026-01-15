// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.

package dsv

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestReaderWithClaset(t *testing.T) {
	fcfg := "testdata/claset.dsv"

	claset := tabula.Claset{}

	_, e := NewReader(fcfg, &claset)
	if e != nil {
		t.Fatal(e)
	}

	test.Assert(t, "", 3, claset.GetClassIndex())

	claset.SetMajorityClass("regular")
	claset.SetMinorityClass("vandalism")

	clone := claset.Clone().(tabula.ClasetInterface)

	test.Assert(t, "", 3, clone.GetClassIndex())
	test.Assert(t, "", "regular", clone.MajorityClass())
	test.Assert(t, "", "vandalism", clone.MinorityClass())
}
