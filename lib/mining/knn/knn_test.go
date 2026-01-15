// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2015 Mhd Sulhan <ms@kilabit.info>

package knn

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestComputeEuclidianDistance(t *testing.T) {
	var exp = []string{
		`[0.302891 0.608544 0.47413 1.42718 -0.811085 1]`,
		`[0.243474 0.505146 0.472892 1.34802 -0.844252 1]` +
			`[0.202343 0.485983 0.527533 1.47307 -0.809672 1]` +
			`[0.215496 0.523418 0.51719 1.43548 -0.933981 1]` +
			`[0.214331 0.546086 0.414773 1.38542 -0.702336 1]` +
			`[0.301676 0.554505 0.594757 1.21258 -0.873084 1]`,
	}
	var expDistances = "[0.5257185558832786" +
		" 0.5690474496911485" +
		" 0.5888777462258191" +
		" 0.6007362149895741" +
		" 0.672666336306493]"

	// Reading data
	dataset := tabula.Dataset{}
	_, e := dsv.SimpleRead("../testdata/phoneme/phoneme.dsv", &dataset)
	if nil != e {
		return
	}

	// Processing
	knnIn := Runtime{
		DistanceMethod: TEuclidianDistance,
		ClassIndex:     5,
		K:              5,
	}

	classes := dataset.GetRows().GroupByValue(knnIn.ClassIndex)

	_, minoritySet := classes.GetMinority()

	kneighbors := knnIn.FindNeighbors(&minoritySet, minoritySet[0])

	var got string
	rows := kneighbors.Rows()
	for _, row := range *rows {
		got += fmt.Sprint(*row)
	}

	test.Assert(t, "", exp[1], got)

	distances := kneighbors.Distances()
	got = fmt.Sprint(*distances)
	test.Assert(t, "", expDistances, got)
}
