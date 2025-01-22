// SPDX-FileCopyrightText: 2016 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package lnsmote implement the Local-Neighborhood algorithm from the paper,
//
// Maciejewski, Tomasz, and Jerzy Stefanowski. "Local neighbourhood
// extension of SMOTE for mining imbalanced data." Computational
// Intelligence and Data Mining (CIDM), 2011 IEEE Symposium on. IEEE,
// 2011.
package lnsmote

import (
	"crypto/rand"
	"log"
	"math"
	"math/big"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/knn"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/resampling/smote"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

// Runtime parameters for input and output.
type Runtime struct {
	// minorset contain minor class in samples.
	minorset tabula.DatasetInterface

	// datasetRows contain all rows in dataset.
	datasetRows *tabula.Rows

	// outliersRows contain all sample that is detected as outliers.
	outliers tabula.Rows

	// ClassMinor the minority sample in dataset that we want to
	// oversampling.
	ClassMinor string `json:"ClassMinor"`

	// OutliersFile if its not empty then outliers will be saved in file
	// specified by this option.
	OutliersFile string `json:"OutliersFile"`

	// Runtime of SMOTE, since this module extend the SMOTE method.
	smote.Runtime
}

// New create and return new LnSmote object.
func New(percentOver, k, classIndex int, classMinor, outliers string) (
	lnsmoteRun *Runtime,
) {
	lnsmoteRun = &Runtime{
		Runtime: smote.Runtime{
			Runtime: knn.Runtime{
				DistanceMethod: knn.TEuclidianDistance,
				ClassIndex:     classIndex,
				K:              k,
			},
			PercentOver: percentOver,
		},
		ClassMinor:   classMinor,
		OutliersFile: outliers,
	}

	return
}

// Init will initialize LNSmote runtime by checking input values and set it to
// default if not set or invalid.
func (in *Runtime) Init(dataset tabula.DatasetInterface) {
	in.Runtime.Init()

	in.NSynthetic = in.PercentOver / 100.0
	in.datasetRows = dataset.GetDataAsRows()

	in.minorset = tabula.SelectRowsWhere(dataset, in.ClassIndex,
		in.ClassMinor)

	in.outliers = make(tabula.Rows, 0)
}

// Resampling will run resampling process on dataset and return the synthetic
// samples.
func (in *Runtime) Resampling(dataset tabula.DatasetInterface) (
	e error,
) {
	in.Init(dataset)

	minorRows := in.minorset.GetDataAsRows()

	for x := range *minorRows {
		p := (*minorRows)[x]

		neighbors := in.FindNeighbors(in.datasetRows, p)

		for range in.NSynthetic {
			syn := in.createSynthetic(p, neighbors)

			if syn != nil {
				in.Synthetics.PushRow(syn)
			}
		}
	}

	if in.SyntheticFile != "" {
		e = in.Write(in.SyntheticFile)
	}
	if in.OutliersFile != "" && in.outliers.Len() > 0 {
		e = in.writeOutliers()
	}

	return e
}

// createSynthetic will create synthetics row from original row `p` and their
// `neighbors`.
func (in *Runtime) createSynthetic(p *tabula.Row, neighbors knn.Neighbors) (
	synthetic *tabula.Row,
) {
	var (
		logp    = `createSynthetic`
		randMax = big.NewInt(int64(neighbors.Len()))

		randv *big.Int
		err   error
	)

	// choose one of the K nearest neighbors
	randv, err = rand.Int(rand.Reader, randMax)
	if err != nil {
		log.Panicf(`%s: %s`, logp, err)
	}

	randIdx := int(randv.Int64())
	n := neighbors.Row(randIdx)

	// Check if synthetic sample can be created from p and n.
	canit, slp, sln := in.canCreate(p, n)
	if !canit {
		if slp.Len() <= 0 {
			in.outliers.PushBack(p)
		}

		// we can not create from p and synthetic.
		return nil
	}

	synthetic = p.Clone()

	for x, srec := range *synthetic {
		// Skip class attribute.
		if x == in.ClassIndex {
			continue
		}

		delta := in.randomGap(slp.Len(), sln.Len())
		pv := (*p)[x].Float()
		diff := (*n)[x].Float() - pv
		srec.SetFloat(pv + delta*diff)
	}

	return synthetic
}

// canCreate return true if synthetic can be created between two sample `p` and
// `n`. Otherwise it will return false.
func (in *Runtime) canCreate(p, n *tabula.Row) (bool, knn.Neighbors,
	knn.Neighbors,
) {
	slp := in.safeLevel(p)
	sln := in.safeLevel2(p, n)

	return slp.Len() != 0 || sln.Len() != 0, slp, sln
}

// safeLevel return the minority neighbors in sample `p`.
func (in *Runtime) safeLevel(p *tabula.Row) knn.Neighbors {
	neighbors := in.FindNeighbors(in.datasetRows, p)
	minorNeighbors := neighbors.SelectWhere(in.ClassIndex, in.ClassMinor)

	return minorNeighbors
}

// safeLevel2 return the minority neighbors between sample `p` and `n`.
func (in *Runtime) safeLevel2(p, n *tabula.Row) knn.Neighbors {
	neighbors := in.FindNeighbors(in.datasetRows, n)

	// check if n is in minority class.
	nIsMinor := (*n)[in.ClassIndex].IsEqualToString(in.ClassMinor)

	// check if p is in neighbors.
	pInNeighbors, pidx := neighbors.Contain(p)

	// if p in neighbors, replace it with neighbours in K+1
	if nIsMinor && pInNeighbors {
		row := in.AllNeighbors.Row(in.K + 1)
		dist := in.AllNeighbors.Distance(in.K + 1)
		neighbors.Replace(pidx, row, dist)
	}

	minorNeighbors := neighbors.SelectWhere(in.ClassIndex, in.ClassMinor)

	return minorNeighbors
}

// randomGap return the neighbors gap between sample `p` and `n` using safe
// level (number of minority neighbors) of p in `lenslp` and `n` in `lensln`.
func (in *Runtime) randomGap(lenslp, lensln int) (
	delta float64,
) {
	if lensln == 0 && lenslp > 0 {
		return
	}

	var (
		logp    = `randomGap`
		randMax = big.NewInt(math.MaxInt64)
		randv   *big.Int
		err     error
	)

	randv, err = rand.Int(rand.Reader, randMax)
	if err != nil {
		log.Panicf(`%s: %s`, logp, err)
	}

	var randf64 = float64(randv.Int64()) / float64(math.MaxInt64)

	slratio := float64(lenslp) / float64(lensln)
	switch {
	case slratio == 1:
		delta = randf64
	case slratio > 1:
		delta = randf64 * (1 / slratio)
	default:
		delta = 1 - randf64*slratio
	}

	return delta
}

// writeOutliers will save the `outliers` to file specified by
// `OutliersFile`.
func (in *Runtime) writeOutliers() (e error) {
	writer, e := dsv.NewWriter("")
	if nil != e {
		return
	}

	e = writer.OpenOutput(in.OutliersFile)
	if e != nil {
		return
	}

	sep := dsv.DefSeparator
	_, e = writer.WriteRawRows(&in.outliers, &sep)
	if e != nil {
		return
	}

	return writer.Close()
}
