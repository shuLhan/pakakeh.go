// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
