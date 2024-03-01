// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crf implement the cascaded random forest algorithm, proposed by
// Baumann et.al in their paper:
//
// Baumann, Florian, et al. "Cascaded Random Forest for Fast Object
// Detection." Image Analysis. Springer Berlin Heidelberg, 2013. 131-142.
package crf

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"git.sr.ht/~shulhan/pakakeh.go/lib/floats64"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier/rf"
	"git.sr.ht/~shulhan/pakakeh.go/lib/numbers"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

const (
	tag = "[crf]"

	// DefStage default number of stage
	DefStage = 200
	// DefTPRate default threshold for true-positive rate.
	DefTPRate = 0.9
	// DefTNRate default threshold for true-negative rate.
	DefTNRate = 0.7

	// DefNumTree default number of tree.
	DefNumTree = 1
	// DefPercentBoot default percentage of sample that will be used for
	// bootstraping a tree.
	DefPercentBoot = 66
	// DefPerfFile default performance file output.
	DefPerfFile = "crf.perf"
	// DefStatFile default statistic file output.
	DefStatFile = "crf.stat"
)

var (
	// ErrNoInput will tell you when no input is given.
	ErrNoInput = errors.New("rf: input samples is empty")
)

// Runtime define the cascaded random forest runtime input and output.
type Runtime struct {
	// tnset contain sample of all true-negative in each iteration.
	tnset *tabula.Claset

	// forests contain forest for each stage.
	forests []*rf.Runtime

	// weights contain weight for each stage.
	weights []float64

	// Runtime embed common fields for classifier.
	classifier.Runtime

	// NStage number of stage.
	NStage int `json:"NStage"`

	// NTree number of tree in each stage.
	NTree int `json:"NTree"`

	// TPRate threshold for true positive rate per stage.
	TPRate float64 `json:"TPRate"`

	// TNRate threshold for true negative rate per stage.
	TNRate float64 `json:"TNRate"`

	// NRandomFeature number of features used to split the dataset.
	NRandomFeature int `json:"NRandomFeature"`

	// PercentBoot percentage of bootstrap.
	PercentBoot int `json:"PercentBoot"`
}

// New create and return new input for cascaded random-forest.
func New(nstage, ntree, percentboot, nfeature int, tprate, tnrate float64) (crf *Runtime) {
	crf = &Runtime{
		NStage:         nstage,
		NTree:          ntree,
		PercentBoot:    percentboot,
		NRandomFeature: nfeature,
		TPRate:         tprate,
		TNRate:         tnrate,
	}

	return crf
}

// AddForest will append new forest.
func (crf *Runtime) AddForest(forest *rf.Runtime) {
	crf.forests = append(crf.forests, forest)
}

// Initialize will check crf inputs and set it to default values if its
// invalid.
func (crf *Runtime) Initialize(samples tabula.ClasetInterface) error {
	if crf.NStage <= 0 {
		crf.NStage = DefStage
	}
	if crf.TPRate <= 0 || crf.TPRate >= 1 {
		crf.TPRate = DefTPRate
	}
	if crf.TNRate <= 0 || crf.TNRate >= 1 {
		crf.TNRate = DefTNRate
	}
	if crf.NTree <= 0 {
		crf.NTree = DefNumTree
	}
	if crf.PercentBoot <= 0 {
		crf.PercentBoot = DefPercentBoot
	}
	if crf.NRandomFeature <= 0 {
		// Set default value to square-root of features.
		ncol := samples.GetNColumn() - 1
		crf.NRandomFeature = int(math.Sqrt(float64(ncol)))
	}
	if crf.PerfFile == "" {
		crf.PerfFile = DefPerfFile
	}
	if crf.StatFile == "" {
		crf.StatFile = DefStatFile
	}
	crf.tnset = samples.Clone().(*tabula.Claset)

	return crf.Runtime.Initialize()
}

// Build given a sample dataset, build the stage with randomforest.
func (crf *Runtime) Build(samples tabula.ClasetInterface) (e error) {
	if samples == nil {
		return ErrNoInput
	}

	e = crf.Initialize(samples)
	if e != nil {
		return
	}

	fmt.Println(tag, "Training samples:", samples)
	fmt.Println(tag, "Sample (one row):", samples.GetRow(0))
	fmt.Println(tag, "Config:", crf)

	for x := 0; x < crf.NStage; x++ {
		forest, e := crf.createForest(samples)
		if e != nil {
			return e
		}

		e = crf.finalizeStage(forest)
		if e != nil {
			return e
		}
	}

	return crf.Finalize()
}

// createForest will create and return a forest and run the training `samples`
// on it.
//
// Algorithm,
// (1) Initialize forest.
// (2) For 0 to maximum number of tree in forest,
// (2.1) grow one tree until success.
// (2.2) If tree tp-rate and tn-rate greater than threshold, stop growing.
// (3) Calculate weight.
// (4) TODO: Move true-negative from samples. The collection of true-negative
// will be used again to test the model and after test and the sample with FP
// will be moved to training samples again.
// (5) Refill samples with false-positive.
func (crf *Runtime) createForest(samples tabula.ClasetInterface) (
	forest *rf.Runtime, e error,
) {
	var cm *classifier.CM
	var stat *classifier.Stat

	fmt.Println(tag, "Forest samples:", samples)

	// (1)
	forest = &rf.Runtime{
		Runtime: classifier.Runtime{
			RunOOB: true,
		},
		NTree:          crf.NTree,
		NRandomFeature: crf.NRandomFeature,
	}

	e = forest.Initialize(samples)
	if e != nil {
		return nil, e
	}

	// (2)
	for t := 0; t < crf.NTree; t++ {
		// (2.1)
		for {
			cm, stat, e = forest.GrowTree(samples)
			if e == nil {
				break
			}
		}

		// (2.2)
		if stat.TPRate > crf.TPRate &&
			stat.TNRate > crf.TNRate {
			break
		}
	}

	e = forest.Finalize()
	if e != nil {
		return nil, e
	}

	// (3)
	crf.computeWeight(stat)

	// (4)
	crf.deleteTrueNegative(samples, cm)

	// (5)
	crf.runTPSet(samples)

	samples.RecountMajorMinor()

	return forest, nil
}

// finalizeStage save forest and write the forest statistic to file.
func (crf *Runtime) finalizeStage(forest *rf.Runtime) (e error) {
	stat := forest.StatTotal()
	stat.ID = int64(len(crf.forests))

	e = crf.WriteOOBStat(stat)
	if e != nil {
		return e
	}

	crf.AddStat(stat)
	crf.ComputeStatTotal(stat)

	// (7)
	crf.AddForest(forest)

	return nil
}

// computeWeight will compute the weight of stage based on F-measure of the
// last tree in forest.
func (crf *Runtime) computeWeight(stat *classifier.Stat) {
	crf.weights = append(crf.weights, math.Exp(stat.FMeasure))
}

//
// deleteTrueNegative will delete all samples data where their row index is in
// true-negative values in confusion matrix and move it to TN-set.
//
// (1) Move true negative to tnset on the first iteration, on the next
// iteration it will be full deleted.
// (2) Delete TN from sample set one-by-one with offset, to make sure we
// are not deleting with wrong index.

func (crf *Runtime) deleteTrueNegative(samples tabula.ClasetInterface,
	cm *classifier.CM,
) {
	var row *tabula.Row

	tnids := cm.TNIndices()
	sort.Ints(tnids)

	// (1)
	if len(crf.weights) <= 1 {
		for _, i := range tnids {
			crf.tnset.PushRow(samples.GetRow(i))
		}
	}

	// (2)
	c := 0
	for x, i := range tnids {
		row = samples.DeleteRow(i - x)
		if row != nil {
			c++
		}
	}
}

// refillWithFP will copy the false-positive data in training set `tnset`
// and append it to `samples`.
func (crf *Runtime) refillWithFP(samples, tnset tabula.ClasetInterface,
	cm *classifier.CM,
) {
	// Get and sort FP.
	fpids := cm.FPIndices()
	sort.Ints(fpids)

	// Move FP samples from TN-set to training set samples.
	for _, i := range fpids {
		samples.PushRow(tnset.GetRow(i))
	}

	// Delete FP from training set.
	var row *tabula.Row
	c := 0
	for x, i := range fpids {
		row = tnset.DeleteRow(i - x)
		if row != nil {
			c++
		}
	}
}

// runTPSet will run true-positive set into trained stage, to get the
// false-positive. The FP samples will be added to training set.
func (crf *Runtime) runTPSet(samples tabula.ClasetInterface) {
	// Skip the first stage, because we just got tnset from them.
	if len(crf.weights) <= 1 {
		return
	}

	tnListID := numbers.IntCreateSeq(0, crf.tnset.Len()-1)
	_, cm, _ := crf.ClassifySetByWeight(crf.tnset, tnListID)

	crf.refillWithFP(samples, crf.tnset, cm)
}

// ClassifySetByWeight will classify each instance in samples by weight
// with respect to its single performance.
//
// Algorithm,
// (1) For each instance in samples,
// (1.1) for each stage,
// (1.1.1) collect votes for instance in current stage.
// (1.1.2) Compute probabilities of each classes in votes.
//
//	prob_class = count_of_class / total_votes
//
// (1.1.3) Compute total of probabilities times of stage weight.
//
//	stage_prob = prob_class * stage_weight
//
// (1.2) Divide each class stage probabilities with
//
//	stage_prob = stage_prob /
//		(sum_of_all_weights * number_of_tree_in_forest)
//
// (1.3) Select class label with highest probabilities.
// (1.4) Save stage probabilities for positive class.
// (2) Compute confusion matrix.
func (crf *Runtime) ClassifySetByWeight(samples tabula.ClasetInterface,
	sampleListID []int,
) (
	predicts []string, cm *classifier.CM, probs []float64,
) {
	stat := classifier.Stat{}
	stat.Start()

	vs := samples.GetClassValueSpace()
	stageProbs := make([]float64, len(vs))
	stageSumProbs := make([]float64, len(vs))
	sumWeights := floats64.Sum(crf.weights)

	// (1)
	rows := samples.GetDataAsRows()
	for _, row := range *rows {
		for y := range stageSumProbs {
			stageSumProbs[y] = 0
		}

		// (1.1)
		for y, forest := range crf.forests {
			// (1.1.1)
			votes := forest.Votes(row, -1)

			// (1.1.2)
			var votesProbs = libstrings.FrequencyOfTokens(votes, vs, false)

			// (1.1.3)
			for z := range votesProbs {
				stageSumProbs[z] += votesProbs[z]
				stageProbs[z] += votesProbs[z] * crf.weights[y]
			}
		}

		// (1.2)
		stageWeight := sumWeights * float64(crf.NTree)

		for x := range stageProbs {
			stageProbs[x] /= stageWeight
		}

		// (1.3)
		_, maxi, ok := floats64.Max(stageProbs)
		if ok {
			predicts = append(predicts, vs[maxi])
		}

		probs = append(probs, stageSumProbs[0]/float64(len(crf.forests)))
	}

	// (2)
	actuals := samples.GetClassAsStrings()
	cm = crf.ComputeCM(sampleListID, vs, actuals, predicts)

	crf.ComputeStatFromCM(&stat, cm)
	stat.End()

	_ = stat.Write(crf.StatFile)

	return predicts, cm, probs
}
