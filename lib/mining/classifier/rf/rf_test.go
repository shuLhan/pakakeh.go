// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2016 Mhd Sulhan <ms@kilabit.info>

package rf

import (
	"fmt"
	"log"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

// Global options to run for each test.
var (
	// SampleDsvFile is the file that contain samples config.
	SampleDsvFile string
	// DoTest if its true then the dataset will splited into training and
	// test set with random selection without replacement.
	DoTest = false
	// NTree number of tree to generate.
	NTree = 100
	// NBootstrap percentage of sample used as subsample.
	NBootstrap = 66
	// MinFeature number of feature to begin with.
	MinFeature = 1
	// MaxFeature maximum number of feature to test
	MaxFeature = -1
	// RunOOB if its true the the OOB samples will be used to test the
	// model in each iteration.
	RunOOB = true
	// OOBStatsFile is the file where OOB statistic will be saved.
	OOBStatsFile string
	// PerfFile is the file where performance statistic will be saved.
	PerfFile string
	// StatFile is the file where classifying statistic will be saved.
	StatFile string
)

func getSamples() (train, test tabula.ClasetInterface) {
	samples := tabula.Claset{}
	_, e := dsv.SimpleRead(SampleDsvFile, &samples)
	if nil != e {
		log.Fatal(e)
	}

	if !DoTest {
		return &samples, nil
	}

	ntrain := int(float32(samples.Len()) * (float32(NBootstrap) / 100.0))

	bag, oob, _, _ := tabula.RandomPickRows(&samples, ntrain, false)

	train = bag.(tabula.ClasetInterface)
	test = oob.(tabula.ClasetInterface)

	train.SetClassIndex(samples.GetClassIndex())
	test.SetClassIndex(samples.GetClassIndex())

	return train, test
}

func runRandomForest() {
	trainset, testset := getSamples()

	if MaxFeature < 0 {
		MaxFeature = trainset.GetNColumn()
	}

	for nfeature := MinFeature; nfeature < MaxFeature; nfeature++ {
		// Add prefix to OOB stats file.
		oobStatsFile := fmt.Sprintf("N%d.%s", nfeature, OOBStatsFile)

		// Add prefix to performance file.
		perfFile := fmt.Sprintf("N%d.%s", nfeature, PerfFile)

		// Add prefix to stat file.
		statFile := fmt.Sprintf("N%d.%s", nfeature, StatFile)

		// Create and build random forest.
		forest := Runtime{
			Runtime: classifier.Runtime{
				RunOOB:       RunOOB,
				OOBStatsFile: oobStatsFile,
				PerfFile:     perfFile,
				StatFile:     statFile,
			},
			NTree:          NTree,
			NRandomFeature: nfeature,
			PercentBoot:    NBootstrap,
		}

		e := forest.Build(trainset)
		if e != nil {
			log.Fatal(e)
		}

		if DoTest {
			predicts, _, probs := forest.ClassifySet(testset, nil)

			forest.Performance(testset, predicts, probs)
			e = forest.WritePerformance()
			if e != nil {
				log.Fatal(e)
			}
		}
	}
}

func TestEnsemblingGlass(_ *testing.T) {
	SampleDsvFile = "../../testdata/forensic_glass/fgl.dsv"
	RunOOB = false
	OOBStatsFile = "glass.oob"
	StatFile = "glass.stat"
	PerfFile = "glass.perf"
	DoTest = true

	runRandomForest()
}

func TestEnsemblingIris(_ *testing.T) {
	SampleDsvFile = "../../testdata/iris/iris.dsv"
	OOBStatsFile = "iris.oob"

	runRandomForest()
}

func TestEnsemblingPhoneme(_ *testing.T) {
	SampleDsvFile = "../../testdata/phoneme/phoneme.dsv"
	OOBStatsFile = "phoneme.oob.stat"
	StatFile = "phoneme.stat"
	PerfFile = "phoneme.perf"

	NTree = 200
	MinFeature = 3
	MaxFeature = 4
	RunOOB = false
	DoTest = true

	runRandomForest()
}

func TestEnsemblingSmotePhoneme(_ *testing.T) {
	SampleDsvFile = "../../resampling/smote/phoneme_smote.dsv"
	OOBStatsFile = "phonemesmote.oob"

	MinFeature = 3
	MaxFeature = 4

	runRandomForest()
}

func TestEnsemblingLnsmotePhoneme(_ *testing.T) {
	SampleDsvFile = "../../resampling/lnsmote/phoneme_lnsmote.dsv"
	OOBStatsFile = "phonemelnsmote.oob"

	MinFeature = 3
	MaxFeature = 4

	runRandomForest()
}

func TestWvc2010Lnsmote(_ *testing.T) {
	SampleDsvFile = "../../testdata/wvc2010lnsmote/wvc2010_features.lnsmote.dsv"
	OOBStatsFile = "wvc2010lnsmote.oob"

	NTree = 1
	MinFeature = 5
	MaxFeature = 6

	runRandomForest()
}

func TestMain(m *testing.M) {
	envTestRF := os.Getenv("TEST_RF")
	if len(envTestRF) == 0 {
		os.Exit(0)
	}

	os.Exit(m.Run())
}
