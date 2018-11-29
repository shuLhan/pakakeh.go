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
	"github.com/shuLhan/share/lib/mining/classifier/crf"
	"github.com/shuLhan/share/lib/tabula"
)

const (
	tag = "[crf]"
)

var (
	// nStage number of stage.
	nStage = 0
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

	// crforest the main object.
	crforest crf.Runtime
)

var usage = func() {
	flag.PrintDefaults()
}

func initFlags() {
	flagUsage := []string{
		"Number of stage (default 200)",
		"Number of tree in each stage (default 1)",
		"Number of feature to compute (default 0, all features)",
		"Percentage of bootstrap (default 64%)",
		"OOB statistic file, where OOB data will be written",
		"Performance file, where statistic of classifying data set will be written",
		"Training configuration",
		"Test configuration",
	}

	flag.IntVar(&nStage, "nstage", -1, flagUsage[0])
	flag.IntVar(&nTree, "ntree", -1, flagUsage[1])
	flag.IntVar(&nRandomFeature, "nrandomfeature", -1, flagUsage[2])
	flag.IntVar(&percentBoot, "percentboot", -1, flagUsage[3])

	flag.StringVar(&oobStatsFile, "oobstatsfile", "", flagUsage[4])
	flag.StringVar(&perfFile, "perffile", "", flagUsage[5])

	flag.StringVar(&trainCfg, "train", "", flagUsage[6])
	flag.StringVar(&testCfg, "test", "", flagUsage[7])

	flag.Parse()
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
// createCRF will create cascaded random forest for training, with the
// following steps,
// (1) load training configuration.
// (2) Overwrite configuration parameter if its set from command line.
//
func createCRF() error {
	// (1)
	config, e := ioutil.ReadFile(trainCfg)
	if e != nil {
		return e
	}

	crforest = crf.Runtime{}

	e = json.Unmarshal(config, &crforest)
	if e != nil {
		return e
	}

	// (2)
	if nStage > 0 {
		crforest.NStage = nStage
	}
	if nTree > 0 {
		crforest.NTree = nTree
	}
	if nRandomFeature > 0 {
		crforest.NRandomFeature = nRandomFeature
	}
	if percentBoot > 0 {
		crforest.PercentBoot = percentBoot
	}
	if oobStatsFile != "" {
		crforest.OOBStatsFile = oobStatsFile
	}
	if perfFile != "" {
		crforest.PerfFile = perfFile
	}

	crforest.RunOOB = true

	return nil
}

func train() {
	e := createCRF()
	if e != nil {
		panic(e)
	}

	trainset := tabula.Claset{}

	_, e = dsv.SimpleRead(trainCfg, &trainset)
	if e != nil {
		panic(e)
	}

	e = crforest.Build(&trainset)
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

	fmt.Println(tag, "Test set:", &testset)
	fmt.Println(tag, "Sample test set:", testset.GetRow(0))

	predicts, cm, probs := crforest.ClassifySetByWeight(&testset, nil)

	fmt.Println("[crf] Test set CM:", cm)

	crforest.Performance(&testset, predicts, probs)

	e = crforest.WritePerformance()
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
	initFlags()

	fmt.Println(tag, "Training config:", trainCfg)
	fmt.Println(tag, "Test config:", testCfg)

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
