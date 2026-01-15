// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2016 Mhd Sulhan <ms@kilabit.info>

package lnsmote

import (
	"fmt"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

const (
	fcfg = "../../testdata/phoneme/phoneme.dsv"
)

func TestLNSmote(t *testing.T) {
	// Read sample dataset.
	dataset := tabula.Claset{}
	_, e := dsv.SimpleRead(fcfg, &dataset)
	if nil != e {
		t.Fatal(e)
	}

	fmt.Println("[lnsmote_test] Total samples:", dataset.GetNRow())

	// Write original samples.
	writer, e := dsv.NewWriter("")

	if nil != e {
		t.Fatal(e)
	}

	e = writer.OpenOutput("phoneme_lnsmote.csv")
	if e != nil {
		t.Fatal(e)
	}

	sep := dsv.DefSeparator
	_, e = writer.WriteRawRows(dataset.GetRows(), &sep)
	if e != nil {
		t.Fatal(e)
	}

	// Initialize LN-SMOTE.
	lnsmoteRun := New(100, 5, 5, "1", "lnsmote.outliers")

	e = lnsmoteRun.Resampling(&dataset)
	if e != nil {
		t.Fatal(e)
	}

	fmt.Println("[lnsmote_test] # synthetic:", lnsmoteRun.Synthetics.Len())

	sep = dsv.DefSeparator
	_, e = writer.WriteRawRows(lnsmoteRun.Synthetics.GetRows(), &sep)
	if e != nil {
		t.Fatal(e)
	}

	e = writer.Close()
	if e != nil {
		t.Fatal(e)
	}
}

func TestMain(m *testing.M) {
	envTestLnsmote := os.Getenv("TEST_LNSMOTE")
	if len(envTestLnsmote) == 0 {
		os.Exit(0)
	}

	os.Exit(m.Run())
}
