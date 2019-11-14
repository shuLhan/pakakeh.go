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

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/dsv"
	"github.com/shuLhan/share/lib/mining/resampling/smote"
	"github.com/shuLhan/share/lib/tabula"
)

type options struct {
	// percentOver contain percentage of over sampling.
	percentOver int
	// knn contain number of nearest neighbours considered when
	// oversampling.
	knn int
	// synFile flag for synthetic file output.
	synFile string
	// merge flag, if its true the original and synthetic will be merged
	// into `synFile`.
	merge bool
}

func usage() {
	cmd := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage of %s:\n"+
		"[-percentover number] "+
		"[-knn number] "+
		"[-syntheticfile string] "+
		"[-merge bool] "+
		"[config.dsv]\n", cmd)
	flag.PrintDefaults()
}

func initFlags() (o options) {
	flagUsage := []string{
		"Percentage of oversampling (default 100)",
		"Number of nearest neighbours (default 5)",
		"File where synthetic samples will be written (default '')",
		"If true then original and synthetic will be merged when" +
			" written to file (default false)",
	}

	flag.IntVar(&o.percentOver, "percentover", -1, flagUsage[0])
	flag.IntVar(&o.knn, "knn", -1, flagUsage[1])
	flag.StringVar(&o.synFile, "syntheticfile", "", flagUsage[2])
	flag.BoolVar(&o.merge, "merge", false, flagUsage[3])

	flag.Parse()

	return o
}

func trace(s string) (string, time.Time) {
	fmt.Println("[START]", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	fmt.Println("[END]", s, "with elapsed time",
		endTime.Sub(startTime))
}

//
// createSmote will create and initialize SMOTE object from config file and
// from command parameter.
//
func createSmote(fcfg string, o *options) (smoteRun *smote.Runtime, e error) {
	smoteRun = &smote.Runtime{}

	config, e := ioutil.ReadFile(fcfg)
	if e != nil {
		return nil, e
	}

	e = json.Unmarshal(config, smoteRun)
	if e != nil {
		return nil, e
	}

	// Use option value from command parameter.
	if o.percentOver > 0 {
		smoteRun.PercentOver = o.percentOver
	}
	if o.knn > 0 {
		smoteRun.K = o.knn
	}

	if debug.Value >= 1 {
		fmt.Println("[smote]", smoteRun)
	}

	return
}

//
// runSmote will select minority class from dataset and run oversampling.
//
func runSmote(smote *smote.Runtime, dataset *tabula.Claset) (e error) {
	minorset := dataset.GetMinorityRows()

	if debug.Value >= 1 {
		fmt.Println("[smote] # minority samples:", minorset.Len())
	}

	e = smote.Resampling(*minorset)
	if e != nil {
		return
	}

	if debug.Value >= 1 {
		fmt.Println("[smote] # synthetics:", smote.Synthetics.Len())
	}

	return
}

// runMerge will append original dataset to synthetic file.
func runMerge(smote *smote.Runtime, dataset *tabula.Claset) (e error) {
	writer, e := dsv.NewWriter("")
	if e != nil {
		return
	}

	e = writer.ReopenOutput(smote.SyntheticFile)
	if e != nil {
		return
	}

	sep := dsv.DefSeparator
	n, e := writer.WriteRawDataset(dataset, &sep)
	if e != nil {
		return
	}

	if debug.Value >= 1 {
		fmt.Println("[smote] # appended:", n)
	}

	return writer.Close()
}

func main() {
	defer un(trace("smote"))

	o := initFlags()

	if len(flag.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	fcfg := flag.Arg(0)

	// Parsing config file and parameter.
	smote, e := createSmote(fcfg, &o)
	if e != nil {
		panic(e)
	}

	// Get dataset.
	dataset := tabula.Claset{}
	_, e = dsv.SimpleRead(fcfg, &dataset)
	if e != nil {
		panic(e)
	}

	fmt.Println("[smote] Dataset:", &dataset)

	row := dataset.GetRow(0)
	fmt.Println("[smote] sample:", row)

	e = runSmote(smote, &dataset)
	if e != nil {
		panic(e)
	}

	if !o.merge {
		return
	}

	e = runMerge(smote, &dataset)
	if e != nil {
		panic(e)
	}
}
