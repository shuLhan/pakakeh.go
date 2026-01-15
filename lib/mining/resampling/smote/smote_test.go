// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2015 Mhd Sulhan <ms@kilabit.info>

package smote

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

func TestSmote(t *testing.T) {
	smot := New(100, 5, 5)

	// Read samples.
	dataset := tabula.Claset{}

	_, e := dsv.SimpleRead("../../testdata/phoneme/phoneme.dsv", &dataset)
	if nil != e {
		t.Fatal(e)
	}

	fmt.Println("[smote_test] Total samples:", dataset.Len())

	minorset := dataset.GetMinorityRows()

	fmt.Println("[smote_test] # minority samples:", minorset.Len())

	e = smot.Resampling(*minorset)
	if e != nil {
		t.Fatal(e)
	}

	fmt.Println("[smote_test] # synthetic:", smot.GetSynthetics().Len())

	e = smot.Write("phoneme_smote.csv")
	if e != nil {
		t.Fatal(e)
	}
}
