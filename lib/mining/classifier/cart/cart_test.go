// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cart

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/dsv"
	"github.com/shuLhan/share/lib/tabula"
	"github.com/shuLhan/share/lib/test"
)

const (
	NRows = 150
)

func TestCART(t *testing.T) {
	fds := "../../testdata/iris/iris.dsv"

	ds := tabula.Claset{}

	_, e := dsv.SimpleRead(fds, &ds)
	if nil != e {
		t.Fatal(e)
	}

	fmt.Println("[cart_test] class index:", ds.GetClassIndex())

	// copy target to be compared later.
	targetv := ds.GetClassAsStrings()

	test.Assert(t, "", NRows, ds.GetNRow(), true)

	// Build CART tree.
	CART, e := New(&ds, SplitMethodGini, 0)
	if e != nil {
		t.Fatal(e)
	}

	fmt.Println("[cart_test] CART Tree:\n", CART)

	// Create test set
	testset := tabula.Claset{}
	_, e = dsv.SimpleRead(fds, &testset)

	if nil != e {
		t.Fatal(e)
	}

	testset.GetClassColumn().ClearValues()

	// Classify test set.
	e = CART.ClassifySet(&testset)
	if nil != e {
		t.Fatal(e)
	}

	test.Assert(t, "", targetv, testset.GetClassAsStrings(), true)
}
