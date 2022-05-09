// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/dsv"
	"github.com/shuLhan/share/lib/mining/resampling/lnsmote"
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

func initFlags() (opts *options) {
	opts = &options{}

	flagUsage := []string{
		"Percentage of oversampling (default 100)",
		"Number of nearest neighbours (default 5)",
		"File where synthetic samples will be written (default '')",
		"If true then original and synthetic will be merged when" +
			" written to file (default false)",
	}

	flag.IntVar(&opts.percentOver, "percentover", -1, flagUsage[0])
	flag.IntVar(&opts.knn, "knn", -1, flagUsage[1])
	flag.StringVar(&opts.synFile, "syntheticfile", "", flagUsage[2])
	flag.BoolVar(&opts.merge, "merge", false, flagUsage[3])

	flag.Parse()

	return opts
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

// createLnsmote will create and initialize SMOTE object from config file and
// from command parameter.
func createLnsmote(fcfg string, opts *options) (lnsmoteRun *lnsmote.Runtime, e error) {
	lnsmoteRun = &lnsmote.Runtime{}

	config, e := os.ReadFile(fcfg)
	if e != nil {
		return nil, e
	}

	e = json.Unmarshal(config, lnsmoteRun)
	if e != nil {
		return nil, e
	}

	// Use option value from command parameter.
	if opts.percentOver > 0 {
		lnsmoteRun.PercentOver = opts.percentOver
	}
	if opts.knn > 0 {
		lnsmoteRun.K = opts.knn
	}

	if debug.Value >= 1 {
		fmt.Println("[lnsmote]", lnsmoteRun)
	}

	return
}

// runLnsmote will select minority class from dataset and run oversampling.
func runLnsmote(lnsmoteRun *lnsmote.Runtime, dataset *tabula.Claset) (e error) {
	e = lnsmoteRun.Resampling(dataset)
	if e != nil {
		return
	}

	if debug.Value >= 1 {
		fmt.Println("[lnsmote] # synthetics:",
			lnsmoteRun.GetSynthetics().Len())
	}

	return
}

// runMerge will append original dataset to synthetic file.
func runMerge(lnsmoteRun *lnsmote.Runtime, dataset *tabula.Claset) (e error) {
	writer, e := dsv.NewWriter("")
	if e != nil {
		return
	}

	e = writer.ReopenOutput(lnsmoteRun.SyntheticFile)
	if e != nil {
		return
	}

	sep := dsv.DefSeparator
	n, e := writer.WriteRawDataset(dataset, &sep)
	if e != nil {
		return
	}

	if debug.Value >= 1 {
		fmt.Println("[lnsmote] # appended:", n)
	}

	return writer.Close()
}

func main() {
	defer un(trace("lnsmote"))

	opts := initFlags()

	if len(flag.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	fcfg := flag.Arg(0)

	// Parsing config file and parameter.
	lnsmoteRun, e := createLnsmote(fcfg, opts)
	if e != nil {
		panic(e)
	}

	// Get dataset.
	dataset := tabula.Claset{}
	_, e = dsv.SimpleRead(fcfg, &dataset)
	if e != nil {
		panic(e)
	}

	fmt.Println("[lnsmote] Dataset:", &dataset)

	row := dataset.GetRow(0)
	fmt.Println("[lnsmote] sample:", row)

	e = runLnsmote(lnsmoteRun, &dataset)
	if e != nil {
		panic(e)
	}

	if !opts.merge {
		return
	}

	e = runMerge(lnsmoteRun, &dataset)
	if e != nil {
		panic(e)
	}
}
