// SPDX-FileCopyrightText: 2015 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package smote resamples a dataset by applying the Synthetic Minority
// Oversampling TEchnique (SMOTE). The original dataset must fit entirely in
// memory.  The amount of SMOTE and number of nearest neighbors may be specified.
// For more information, see
//
// Nitesh V. Chawla et. al. (2002). Synthetic Minority Over-sampling
// Technique. Journal of Artificial Intelligence Research. 16:321-357.
package smote

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"

	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/knn"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/resampling"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

// Runtime for input and output.
type Runtime struct {
	// Synthetics contain output of resampling as synthetic samples.
	Synthetics tabula.Dataset

	// SyntheticFile is a filename where synthetic samples will be written.
	SyntheticFile string `json:"SyntheticFile"`

	// Runtime the K-Nearest-Neighbourhood parameters.
	knn.Runtime

	// PercentOver input for oversampling percentage.
	PercentOver int `json:"PercentOver"`

	// NSynthetic input for number of new synthetic per sample.
	NSynthetic int
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
	var (
		logp         = `populate`
		randMax      = big.NewInt(int64(neighbors.Len()))
		randMaxInt64 = big.NewInt(math.MaxInt64)
		lenAttr      = len(*instance)

		randv *big.Int
		err   error
	)

	for range smote.NSynthetic {
		// choose one of the K nearest neighbors
		randv, err = rand.Int(rand.Reader, randMax)
		if err != nil {
			log.Panicf(`%s: %s`, logp, err)
		}
		n := int(randv.Int64())
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

			randv, err = rand.Int(rand.Reader, randMaxInt64)
			if err != nil {
				log.Panicf(`%s: %s`, logp, err)
			}

			var (
				f64 = float64(randv.Int64())
				gap float64
			)
			if f64 > 0 {
				gap = f64 / float64(math.MaxInt64)
			}

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
