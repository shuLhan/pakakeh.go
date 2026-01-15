// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2016 Mhd Sulhan <ms@kilabit.info>

// Package rf implement ensemble of classifiers using random forest
// algorithm by Breiman and Cutler.
//
// Breiman, Leo. "Random forests." Machine learning 45.1 (2001): 5-32.
//
// The implementation is based on various sources and using author experience.
package rf

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier/cart"
	libslices "git.sr.ht/~shulhan/pakakeh.go/lib/slices"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

const (
	tag = "[rf]"

	// DefNumTree default number of tree.
	DefNumTree = 100

	// DefPercentBoot default percentage of sample that will be used for
	// bootstraping a tree.
	DefPercentBoot = 66

	// DefOOBStatsFile default statistic file output.
	DefOOBStatsFile = "rf.oob.stat"

	// DefPerfFile default performance file output.
	DefPerfFile = "rf.perf"

	// DefStatFile default statistic file.
	DefStatFile = "rf.stat"
)

var (
	// ErrNoInput will tell you when no input is given.
	ErrNoInput = errors.New("rf: input samples is empty")
)

// Runtime contains input and output configuration when generating random forest.
type Runtime struct {
	// trees contain all tree in the forest.
	trees []cart.Runtime

	// bagIndices contain list of index of selected samples at bootstraping
	// for book-keeping.
	bagIndices [][]int

	// Runtime embed common fields for classifier.
	classifier.Runtime

	// NTree number of tree in forest.
	NTree int `json:"NTree"`

	// NRandomFeature number of feature randomly selected for each tree.
	NRandomFeature int `json:"NRandomFeature"`

	// PercentBoot percentage of sample for bootstraping.
	PercentBoot int `json:"PercentBoot"`

	// nSubsample number of samples used for bootstraping.
	nSubsample int
}

// Trees return all tree in forest.
func (forest *Runtime) Trees() []cart.Runtime {
	return forest.trees
}

// AddCartTree add tree to forest
func (forest *Runtime) AddCartTree(tree cart.Runtime) {
	forest.trees = append(forest.trees, tree)
}

// AddBagIndex add bagging index for book keeping.
func (forest *Runtime) AddBagIndex(bagIndex []int) {
	forest.bagIndices = append(forest.bagIndices, bagIndex)
}

// Initialize will check forest inputs and set it to default values if invalid.
//
// It will also calculate number of random samples for each tree using,
//
//	number-of-sample * percentage-of-bootstrap
func (forest *Runtime) Initialize(samples tabula.ClasetInterface) error {
	if forest.NTree <= 0 {
		forest.NTree = DefNumTree
	}
	if forest.PercentBoot <= 0 {
		forest.PercentBoot = DefPercentBoot
	}
	if forest.NRandomFeature <= 0 {
		// Set default value to square-root of features.
		ncol := samples.GetNColumn() - 1
		forest.NRandomFeature = int(math.Sqrt(float64(ncol)))
	}
	if forest.OOBStatsFile == "" {
		forest.OOBStatsFile = DefOOBStatsFile
	}
	if forest.PerfFile == "" {
		forest.PerfFile = DefPerfFile
	}
	if forest.StatFile == "" {
		forest.StatFile = DefStatFile
	}

	forest.nSubsample = int(float32(samples.GetNRow()) *
		(float32(forest.PercentBoot) / 100.0))

	return forest.Runtime.Initialize()
}

// Build the forest using samples dataset.
//
// Algorithm,
//
// (0) Recheck input value: number of tree, percentage bootstrap, etc; and
//
//	Open statistic file output.
//
// (1) For 0 to NTree,
// (1.1) Create new tree, repeat until all trees has been build.
// (2) Compute and write total statistic.
func (forest *Runtime) Build(samples tabula.ClasetInterface) (e error) {
	// check input samples
	if samples == nil {
		return ErrNoInput
	}

	// (0)
	e = forest.Initialize(samples)
	if e != nil {
		return
	}

	fmt.Println(tag, "Training set    :", samples)
	fmt.Println(tag, "Sample (one row):", samples.GetRow(0))
	fmt.Println(tag, "Forest config   :", forest)

	// (1)
	for t := 0; t < forest.NTree; t++ {
		// (1.1)
		for {
			_, _, e = forest.GrowTree(samples)
			if e == nil {
				break
			}

			fmt.Println(tag, "error:", e)
		}
	}

	// (2)
	return forest.Finalize()
}

// GrowTree build a new tree in forest, return OOB error value or error if tree
// can not grow.
//
// Algorithm,
//
// (1) Select random samples with replacement, also with OOB.
// (2) Build tree using CART, without pruning.
// (3) Add tree to forest.
// (4) Save index of random samples for calculating error rate later.
// (5) Run OOB on forest.
// (6) Calculate OOB error rate and statistic values.
func (forest *Runtime) GrowTree(samples tabula.ClasetInterface) (
	cm *classifier.CM, stat *classifier.Stat, e error,
) {
	stat = &classifier.Stat{}
	stat.ID = int64(len(forest.trees))
	stat.Start()

	// (1)
	bag, oob, bagIdx, oobIdx := tabula.RandomPickRows(
		samples.(tabula.DatasetInterface),
		forest.nSubsample, true)

	bagset := bag.(tabula.ClasetInterface)

	// (2)
	cart, e := cart.New(bagset, cart.SplitMethodGini,
		forest.NRandomFeature)
	if e != nil {
		return nil, nil, e
	}

	// (3)
	forest.AddCartTree(*cart)

	// (4)
	forest.AddBagIndex(bagIdx)

	// (5)
	if forest.RunOOB {
		oobset := oob.(tabula.ClasetInterface)
		_, cm, _ = forest.ClassifySet(oobset, oobIdx)

		forest.AddOOBCM(cm)
	}

	stat.End()

	forest.AddStat(stat)

	// (6)
	if forest.RunOOB {
		forest.ComputeStatFromCM(stat, cm)
	}

	forest.ComputeStatTotal(stat)
	e = forest.WriteOOBStat(stat)

	return cm, stat, e
}

// ClassifySet given a samples predict their class by running each sample in
// forest, and return their class prediction with confusion matrix.
// `samples` is the sample that will be predicted, `sampleListID` is the index of
// samples.
// If `sampleListID` is not nil, then sample index will be checked in each tree,
// if the sample is used for training, their vote is not counted.
//
// Algorithm,
//
// (0) Get value space (possible class values in dataset)
// (1) For each row in test-set,
// (1.1) collect votes in all trees,
// (1.2) select majority class vote, and
// (1.3) compute and save the actual class probabilities.
// (2) Compute confusion matrix from predictions.
// (3) Compute stat from confusion matrix.
// (4) Write the stat to file only if sampleListID is empty, which mean its run
// not from OOB set.
func (forest *Runtime) ClassifySet(samples tabula.ClasetInterface,
	sampleListID []int,
) (
	predicts []string, cm *classifier.CM, probs []float64,
) {
	stat := classifier.Stat{}
	stat.Start()

	if len(sampleListID) == 0 {
		fmt.Println(tag, "Classify set:", samples)
		fmt.Println(tag, "Classify set sample (one row):",
			samples.GetRow(0))
	}

	// (0)
	vs := samples.GetClassValueSpace()
	actuals := samples.GetClassAsStrings()
	sampleIdx := -1

	// (1)
	rows := samples.GetRows()
	for x, row := range *rows {
		// (1.1)
		if len(sampleListID) > 0 {
			sampleIdx = sampleListID[x]
		}
		votes := forest.Votes(row, sampleIdx)

		// (1.2)
		classProbs := libstrings.FrequencyOfTokens(votes, vs, false)

		_, idx := libslices.Max2(classProbs)
		if idx >= 0 {
			predicts = append(predicts, vs[idx])
		}

		// (1.3)
		probs = append(probs, classProbs[0])
	}

	// (2)
	cm = forest.ComputeCM(sampleListID, vs, actuals, predicts)

	// (3)
	forest.ComputeStatFromCM(&stat, cm)
	stat.End()

	if len(sampleListID) == 0 {
		fmt.Println(tag, "CM:", cm)
		fmt.Println(tag, "Classifying stat:", stat)
		_ = stat.Write(forest.StatFile)
	}

	return predicts, cm, probs
}

// Votes will return votes, or classes, in each tree based on sample.
// If checkIdx is true then the `sampleIdx` will be checked in if it has been used
// when training the tree, if its exist then the sample will be skipped.
//
// (1) If row is used to build the tree then skip it,
// (2) classify row in tree,
// (3) save tree class value.
func (forest *Runtime) Votes(sample *tabula.Row, sampleIdx int) (
	votes []string,
) {
	for x, tree := range forest.trees {
		// (1)
		if sampleIdx >= 0 {
			exist := slices.Contains(forest.bagIndices[x],
				sampleIdx)
			if exist {
				continue
			}
		}

		// (2)
		class := tree.Classify(sample)

		// (3)
		votes = append(votes, class)
	}
	return votes
}
