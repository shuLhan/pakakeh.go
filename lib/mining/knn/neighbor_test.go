// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package knn

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/tabula"
	"github.com/shuLhan/share/lib/test"
)

var dataFloat64 = [][]float64{ // nolint: gochecknoglobals
	{0.243474, 0.505146, 0.472892, 1.34802, -0.844252, 1},
	{0.202343, 0.485983, 0.527533, 1.47307, -0.809672, 1},
	{0.215496, 0.523418, 0.517190, 1.43548, -0.933981, 1},
	{0.214331, 0.546086, 0.414773, 1.38542, -0.702336, 1},
	{0.301676, 0.554505, 0.594757, 1.21258, -0.873084, 1},
}

var distances = []int{4, 3, 2, 1, 0} // nolint: gochecknoglobals

func createNeigbours() (neighbors Neighbors) {
	for x, d := range dataFloat64 {
		row := tabula.Row{}

		for _, v := range d {
			rec := tabula.NewRecordReal(v)
			row.PushBack(rec)
		}

		neighbors.Add(&row, float64(distances[x]))
	}
	return
}

func createNeigboursByIdx(indices []int) (neighbors Neighbors) {
	for x, idx := range indices {
		row := tabula.Row{}

		for _, v := range dataFloat64[idx] {
			rec := tabula.NewRecordReal(v)
			row.PushBack(rec)
		}

		neighbors.Add(&row, float64(distances[x]))
	}
	return
}

func TestContain(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	neighbors := createNeigbours()

	// pick random sample from neighbors
	pickIdx := rand.Intn(neighbors.Len())
	randSample := neighbors.Row(pickIdx).Clone()

	isin, idx := neighbors.Contain(randSample)

	test.Assert(t, "", true, isin, true)
	test.Assert(t, "", pickIdx, idx, true)

	// change one of record value to check for false.
	(*randSample)[0].SetFloat(0)

	isin, _ = neighbors.Contain(randSample)

	test.Assert(t, "", false, isin, true)
}

func TestSort(t *testing.T) {
	neighbors := createNeigbours()
	exp := createNeigboursByIdx(distances)

	sort.Sort(&neighbors)

	test.Assert(t, "", exp.Rows(), neighbors.Rows(), true)
}
