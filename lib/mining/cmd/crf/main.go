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

type options struct {
	// nStage number of stage.
	nStage int
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

func initFlags() (o options) {
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

	flag.IntVar(&o.nStage, "nstage", -1, flagUsage[0])
	flag.IntVar(&o.nTree, "ntree", -1, flagUsage[1])
	flag.IntVar(&o.nRandomFeature, "nrandomfeature", -1, flagUsage[2])
	flag.IntVar(&o.percentBoot, "percentboot", -1, flagUsage[3])

	flag.StringVar(&o.oobStatsFile, "oobstatsfile", "", flagUsage[4])
	flag.StringVar(&o.perfFile, "perffile", "", flagUsage[5])

	flag.StringVar(&o.trainCfg, "train", "", flagUsage[6])
	flag.StringVar(&o.testCfg, "test", "", flagUsage[7])

	flag.Parse()

	return o
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
func createCRF(o *options) (crforest *crf.Runtime, e error) {
	// (1)
	config, e := ioutil.ReadFile(o.trainCfg)
	if e != nil {
		return nil, e
	}

	crforest = &crf.Runtime{}

	e = json.Unmarshal(config, &crforest)
	if e != nil {
		return nil, e
	}

	// (2)
	if o.nStage > 0 {
		crforest.NStage = o.nStage
	}
	if o.nTree > 0 {
		crforest.NTree = o.nTree
	}
	if o.nRandomFeature > 0 {
		crforest.NRandomFeature = o.nRandomFeature
	}
	if o.percentBoot > 0 {
		crforest.PercentBoot = o.percentBoot
	}
	if o.oobStatsFile != "" {
		crforest.OOBStatsFile = o.oobStatsFile
	}
	if o.perfFile != "" {
		crforest.PerfFile = o.perfFile
	}

	crforest.RunOOB = true

	return crforest, nil
}

func train(o *options) (crforest *crf.Runtime) {
	crforest, e := createCRF(o)
	if e != nil {
		panic(e)
	}

	trainset := tabula.Claset{}

	_, e = dsv.SimpleRead(o.trainCfg, &trainset)
	if e != nil {
		panic(e)
	}

	e = crforest.Build(&trainset)
	if e != nil {
		panic(e)
	}

	return crforest
}

func test(crforest *crf.Runtime, o *options) {
	testset := tabula.Claset{}
	_, e := dsv.SimpleRead(o.testCfg, &testset)
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
	var crforest *crf.Runtime

	defer un(trace())

	// (0)
	o := initFlags()

	fmt.Println(tag, "Training config:", o.trainCfg)
	fmt.Println(tag, "Test config:", o.testCfg)

	// (1)
	if o.trainCfg != "" {
		// (1.1)
		crforest = train(&o)
	} else {
		// (1.2)
		if len(flag.Args()) == 0 {
			usage()
			os.Exit(1)
		}
	}

	// (2)
	if o.testCfg != "" {
		test(crforest, &o)
	}
}
