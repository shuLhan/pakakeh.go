// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package classifier

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestComputeNumeric(t *testing.T) {
	actuals := []int64{1, 1, 1, 0, 0, 0, 0}
	predics := []int64{1, 1, 0, 0, 0, 0, 1}
	vs := []int64{1, 0}
	exp := []int{2, 1, 3, 1}

	cm := &CM{}

	cm.ComputeNumeric(vs, actuals, predics)

	test.Assert(t, "", exp[0], cm.TP(), true)
	test.Assert(t, "", exp[1], cm.FN(), true)
	test.Assert(t, "", exp[2], cm.TN(), true)
	test.Assert(t, "", exp[3], cm.FP(), true)

	fmt.Println(cm)
}

func TestComputeStrings(t *testing.T) {
	actuals := []string{"1", "1", "1", "0", "0", "0", "0"}
	predics := []string{"1", "1", "0", "0", "0", "0", "1"}
	vs := []string{"1", "0"}
	exp := []int{2, 1, 3, 1}

	cm := &CM{}

	cm.ComputeStrings(vs, actuals, predics)

	test.Assert(t, "", exp[0], cm.TP(), true)
	test.Assert(t, "", exp[1], cm.FN(), true)
	test.Assert(t, "", exp[2], cm.TN(), true)
	test.Assert(t, "", exp[3], cm.FP(), true)

	fmt.Println(cm)
}

func TestGroupIndexPredictions(t *testing.T) {
	testIds := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	actuals := []int64{1, 1, 1, 1, 0, 0, 0, 0, 0, 0}
	predics := []int64{1, 1, 0, 1, 0, 0, 0, 0, 1, 0}
	exp := [][]int{
		{0, 1, 3},       // tp
		{2},             // fn
		{8},             // fp
		{4, 5, 6, 7, 9}, // tn
	}

	cm := &CM{}

	cm.GroupIndexPredictions(testIds, actuals, predics)

	test.Assert(t, "", exp[0], cm.TPIndices(), true)
	test.Assert(t, "", exp[1], cm.FNIndices(), true)
	test.Assert(t, "", exp[2], cm.FPIndices(), true)
	test.Assert(t, "", exp[3], cm.TNIndices(), true)
}
