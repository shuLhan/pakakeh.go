// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package smote resamples a dataset by applying the Synthetic Minority
// Oversampling TEchnique (SMOTE). The original dataset must fit entirely in
// memory.  The amount of SMOTE and number of nearest neighbors may be specified.
// For more information, see
//
// Nitesh V. Chawla et. al. (2002). Synthetic Minority Over-sampling
// Technique. Journal of Artificial Intelligence Research. 16:321-357.
package smote

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shuLhan/share/lib/mining/knn"
	"github.com/shuLhan/share/lib/mining/resampling"
	"github.com/shuLhan/share/lib/tabula"
)

// Runtime for input and output.
type Runtime struct {
	// Runtime the K-Nearest-Neighbourhood parameters.
	knn.Runtime
	// PercentOver input for oversampling percentage.
	PercentOver int `json:"PercentOver"`
	// SyntheticFile is a filename where synthetic samples will be written.
	SyntheticFile string `json:"SyntheticFile"`
	// NSynthetic input for number of new synthetic per sample.
	NSynthetic int
	// Synthetics contain output of resampling as synthetic samples.
	Synthetics tabula.Dataset
}

// New create and return new smote runtime.
func New(percentOver, k, classIndex int) (smoteRun *Runtime) {
	smoteRun = &Runtime{
		Runtime: knn.Runtime{
			DistanceMethod: knn.TEuclidianDistance,
			ClassIndex:     classIndex,
			K:              k,
		},
		PercentOver: percentOver,
	}
	return
}

// Init will recheck input and set to default value if its not valid.
func (smote *Runtime) Init() {
	rand.Seed(time.Now().UnixNano())

	if smote.K <= 0 {
		smote.K = resampling.DefaultK
	}
	if smote.PercentOver <= 0 {
		smote.PercentOver = resampling.DefaultPercentOver
	}
}

// GetSynthetics return synthetic samples.
func (smote *Runtime) GetSynthetics() tabula.DatasetInterface {
	return &smote.Synthetics
}

// populate will generate new synthetic sample using nearest neighbors.
func (smote *Runtime) populate(instance *tabula.Row, neighbors knn.Neighbors) {
	lenAttr := len(*instance)

	for x := 0; x < smote.NSynthetic; x++ {
		// choose one of the K nearest neighbors
		n := rand.Intn(neighbors.Len())
		sample := neighbors.Row(n)

		newSynt := make(tabula.Row, lenAttr)

		// Compute new synthetic attributes.
		for attr, sr := range *sample {
			if attr == smote.ClassIndex {
				continue
			}

			ir := (*instance)[attr]

			iv := ir.Float()
			sv := sr.Float()

			dif := sv - iv
			gap := rand.Float64()
			newAttr := iv + (gap * dif)

			record := &tabula.Record{}
			record.SetFloat(newAttr)
			newSynt[attr] = record
		}

		newSynt[smote.ClassIndex] = (*instance)[smote.ClassIndex]

		smote.Synthetics.PushRow(&newSynt)
	}
}

// Resampling will run resampling algorithm using values that has been defined
// in `Runtime` and return list of synthetic samples.
//
// The `dataset` must be samples of minority class not the whole dataset.
//
// Algorithms,
//
// (0) If oversampling percentage less than 100, then
// (0.1) replace the input dataset by selecting n random sample from dataset
//
//	      without replacement, where n is
//
//		(percentage-oversampling / 100) * number-of-sample
//
// (1) For each `sample` in dataset,
// (1.1) find k-nearest-neighbors of `sample`,
// (1.2) generate synthetic sample in neighbors.
// (2) Write synthetic samples to file, only if `SyntheticFile` is not empty.
func (smote *Runtime) Resampling(dataset tabula.Rows) (e error) {
	smote.Init()

	if smote.PercentOver < 100 {
		// (0.1)
		smote.NSynthetic = (smote.PercentOver / 100.0) * len(dataset)
		dataset, _, _, _ = dataset.RandomPick(smote.NSynthetic, false)
	} else {
		smote.NSynthetic = smote.PercentOver / 100.0
	}

	// (1)
	for x := range dataset {
		sample := dataset[x]

		// (1.1)
		neighbors := smote.FindNeighbors(&dataset, sample)

		// (1.2)
		smote.populate(sample, neighbors)
	}

	// (2)
	if smote.SyntheticFile != "" {
		e = resampling.WriteSynthetics(smote, smote.SyntheticFile)
	}

	return
}

// Write will write synthetic samples to file defined in `file`.
func (smote *Runtime) Write(file string) error {
	return resampling.WriteSynthetics(smote, file)
}

func (smote *Runtime) String() (s string) {
	s = fmt.Sprintf("'smote' : {\n"+
		"		'ClassIndex'     :%d\n"+
		"	,	'K'              :%d\n"+
		"	,	'PercentOver'    :%d\n"+
		"	,	'DistanceMethod' :%d\n"+
		"}", smote.ClassIndex, smote.K, smote.PercentOver,
		smote.DistanceMethod)

	return
}
