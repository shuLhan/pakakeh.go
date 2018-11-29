// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/shuLhan/share/lib/dsv"
	"github.com/shuLhan/share/lib/mining/classifier/rf"
	"github.com/shuLhan/share/lib/tabula"
)

const (
	tag = "[rf]"
)

var (
	// nTree number of tree.
	nTree = 0
	// nRandomFeature number of feature to compute.
	nRandomFeature = 0
	// percentBoot percentage of sample for bootstraping.
	percentBoot = 0
	// oobStatsFile where statistic will be written.
	oobStatsFile = ""
	// perfFile where performance of classifier will be written.
	perfFile = ""
	// trainCfg point to the configuration file for training or creating
	// a model
	trainCfg = ""
	// testCfg point to the configuration file for testing
	testCfg = ""

	// forest the main object.
	forest rf.Runtime
)

var usage = func() {
	flag.PrintDefaults()
}

func init() {
	flagUsage := []string{
		"Number of tree in forest (default 100)",
		"Number of feature to compute (default 0)",
		"Percentage of bootstrap (default 64%)",
		"OOB statistic file, where OOB data will be written",
		"Performance file, where statistic of classifying data set will be written",
		"Training configuration",
		"Test configuration",
	}

	flag.IntVar(&nTree, "ntree", -1, flagUsage[0])
	flag.IntVar(&nRandomFeature, "nrandomfeature", -1, flagUsage[1])
	flag.IntVar(&percentBoot, "percentboot", -1, flagUsage[2])

	flag.StringVar(&oobStatsFile, "oobstatsfile", "", flagUsage[3])
	flag.StringVar(&perfFile, "perffile", "", flagUsage[4])

	flag.StringVar(&trainCfg, "train", "", flagUsage[5])
	flag.StringVar(&testCfg, "test", "", flagUsage[6])
}

func trace() (start time.Time) {
	start = time.Now()
	fmt.Println(tag, "start", start)
	return
}

func un(startTime time.Time) {
	endTime := time.Now()
	fmt.Println(tag, "elapsed time", endTime.Sub(startTime))
}

//
// createRandomForest will create random forest for training, with the
// following steps,
// (1) load training configuration.
// (2) Overwrite configuration parameter if its set from command line.
//
func createRandomForest() error {
	// (1)
	config, e := ioutil.ReadFile(trainCfg)
	if e != nil {
		return e
	}

	forest = rf.Runtime{}

	e = json.Unmarshal(config, &forest)
	if e != nil {
		return e
	}

	// (2)
	if nTree > 0 {
		forest.NTree = nTree
	}
	if nRandomFeature > 0 {
		forest.NRandomFeature = nRandomFeature
	}
	if percentBoot > 0 {
		forest.PercentBoot = percentBoot
	}
	if oobStatsFile != "" {
		forest.OOBStatsFile = oobStatsFile
	}
	if perfFile != "" {
		forest.PerfFile = perfFile
	}

	return nil
}

func train() {
	e := createRandomForest()
	if e != nil {
		panic(e)
	}

	trainset := tabula.Claset{}

	_, e = dsv.SimpleRead(trainCfg, &trainset)
	if e != nil {
		panic(e)
	}

	e = forest.Build(&trainset)
	if e != nil {
		panic(e)
	}
}

func test() {
	testset := tabula.Claset{}
	_, e := dsv.SimpleRead(testCfg, &testset)
	if e != nil {
		panic(e)
	}

	predicts, _, probs := forest.ClassifySet(&testset, nil)

	forest.Performance(&testset, predicts, probs)

	e = forest.WritePerformance()
	if e != nil {
		panic(e)
	}
}

//
// (0) Parse and check command line parameters.
// (1) If trainCfg parameter is set,
// (1.1) train the model,
// (1.2) TODO: load saved model.
// (2) If testCfg parameter is set,
// (2.1) Test the model using data from testCfg.
//
func main() {
	defer un(trace())

	// (0)
	flag.Parse()

	// (1)
	if trainCfg != "" {
		// (1.1)
		train()
	} else {
		// (1.2)
		if len(flag.Args()) == 0 {
			usage()
			os.Exit(1)
		}
	}

	// (2)
	if testCfg != "" {
		test()
	}
}
