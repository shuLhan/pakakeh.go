// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2015 M. Shulhan <ms@kilabit.info>

package dsv

import (
	"testing"

	"kilabit.info/pakakeh.go/lib/tabula"
	"kilabit.info/pakakeh.go/lib/test"
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
