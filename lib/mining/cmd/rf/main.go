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

type options struct {
	// nTree number of tree.
	nTree int
	// nRandomFeature number of feature to compute.
	nRandomFeature int
	// percentBoot percentage of sample for bootstraping.
	percentBoot int
	// oobStatsFile where statistic will be written.
	oobStatsFile string
	// perfFile where performance of classifier will be written.
	perfFile string
	// trainCfg point to the configuration file for training or creating
	// a model
	trainCfg string
	// testCfg point to the configuration file for testing
	testCfg string
}

func usage() {
	flag.PrintDefaults()
}

func initFlags() (opts options) {
	flagUsage := []string{
		"Number of tree in forest (default 100)",
		"Number of feature to compute (default 0)",
		"Percentage of bootstrap (default 64%)",
		"OOB statistic file, where OOB data will be written",
		"Performance file, where statistic of classifying data set will be written",
		"Training configuration",
		"Test configuration",
	}

	flag.IntVar(&opts.nTree, "ntree", -1, flagUsage[0])
	flag.IntVar(&opts.nRandomFeature, "nrandomfeature", -1, flagUsage[1])
	flag.IntVar(&opts.percentBoot, "percentboot", -1, flagUsage[2])

	flag.StringVar(&opts.oobStatsFile, "oobstatsfile", "", flagUsage[3])
	flag.StringVar(&opts.perfFile, "perffile", "", flagUsage[4])

	flag.StringVar(&opts.trainCfg, "train", "", flagUsage[5])
	flag.StringVar(&opts.testCfg, "test", "", flagUsage[6])

	flag.Parse()

	return opts
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
func createRandomForest(opts *options) (forest *rf.Runtime, e error) {
	// (1)
	config, e := ioutil.ReadFile(opts.trainCfg)
	if e != nil {
		return nil, e
	}

	forest = &rf.Runtime{}

	e = json.Unmarshal(config, &forest)
	if e != nil {
		return nil, e
	}

	// (2)
	if opts.nTree > 0 {
		forest.NTree = opts.nTree
	}
	if opts.nRandomFeature > 0 {
		forest.NRandomFeature = opts.nRandomFeature
	}
	if opts.percentBoot > 0 {
		forest.PercentBoot = opts.percentBoot
	}
	if opts.oobStatsFile != "" {
		forest.OOBStatsFile = opts.oobStatsFile
	}
	if opts.perfFile != "" {
		forest.PerfFile = opts.perfFile
	}

	return forest, nil
}

func train(opts *options) (forest *rf.Runtime) {
	var e error

	forest, e = createRandomForest(opts)
	if e != nil {
		panic(e)
	}

	trainset := tabula.Claset{}

	_, e = dsv.SimpleRead(opts.trainCfg, &trainset)
	if e != nil {
		panic(e)
	}

	e = forest.Build(&trainset)
	if e != nil {
		panic(e)
	}

	return forest
}

func test(forest *rf.Runtime, opts *options) {
	testset := tabula.Claset{}
	_, e := dsv.SimpleRead(opts.testCfg, &testset)
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
	var forest *rf.Runtime

	defer un(trace())

	// (0)
	opts := initFlags()

	// (1)
	if opts.trainCfg != "" {
		// (1.1)
		forest = train(&opts)
	} else {
		// (1.2)
		if len(flag.Args()) == 0 {
			usage()
			os.Exit(1)
		}
	}

	// (2)
	if opts.testCfg != "" {
		test(forest, &opts)
	}
}
