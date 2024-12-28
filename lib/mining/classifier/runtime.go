// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package classifier

import (
	"fmt"
	"math"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/floats64"
	"git.sr.ht/~shulhan/pakakeh.go/lib/numbers"
	"git.sr.ht/~shulhan/pakakeh.go/lib/slices"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

const (
	tag = "[classifier.runtime]"
)

// Runtime define a generic type which provide common fields that can be
// embedded by the real classifier (e.g. RandomForest).
type Runtime struct {
	// oobWriter contain file writer for statistic.
	oobWriter *dsv.Writer

	// OOBStatsFile is the file where OOB statistic will be written.
	OOBStatsFile string `json:"OOBStatsFile"`

	// PerfFile is the file where statistic of performance will be written.
	PerfFile string `json:"PerfFile"`

	// StatFile is the file where statistic of classifying samples will be
	// written.
	StatFile string `json:"StatFile"`

	// oobStats contain statistic of classifier for each OOB in iteration.
	oobStats Stats

	// perfs contain performance statistic per sample, after classifying
	// sample on classifier.
	perfs Stats

	// oobCms contain confusion matrix value for each OOB in iteration.
	oobCms []CM

	// oobStatTotal contain total OOB statistic values.
	oobStatTotal Stat

	// RunOOB if its true the OOB will be computed, default is false.
	RunOOB bool `json:"RunOOB"`
}

// Initialize will start the runtime for processing by saving start time and
// opening stats file.
func (rt *Runtime) Initialize() error {
	rt.oobStatTotal.Start()

	return rt.OpenOOBStatsFile()
}

// Finalize finish the runtime, compute total statistic, write it to file, and
// close the file.
func (rt *Runtime) Finalize() (e error) {
	st := &rt.oobStatTotal

	st.End()
	st.ID = int64(len(rt.oobStats))

	e = rt.WriteOOBStat(st)
	if e != nil {
		return e
	}

	return rt.CloseOOBStatsFile()
}

// OOBStats return all statistic objects.
func (rt *Runtime) OOBStats() *Stats {
	return &rt.oobStats
}

// StatTotal return total statistic.
func (rt *Runtime) StatTotal() *Stat {
	return &rt.oobStatTotal
}

// AddOOBCM will append new confusion matrix.
func (rt *Runtime) AddOOBCM(cm *CM) {
	rt.oobCms = append(rt.oobCms, *cm)
}

// AddStat will append new classifier statistic data.
func (rt *Runtime) AddStat(stat *Stat) {
	rt.oobStats = append(rt.oobStats, stat)
}

// ComputeCM will compute confusion matrix of sample using value space, actual
// and prediction values.
func (rt *Runtime) ComputeCM(sampleListID []int,
	vs, actuals, predicts []string,
) (
	cm *CM,
) {
	cm = &CM{}

	cm.ComputeStrings(vs, actuals, predicts)
	cm.GroupIndexPredictionsStrings(sampleListID, actuals, predicts)

	return cm
}

// ComputeStatFromCM will compute statistic using confusion matrix.
func (rt *Runtime) ComputeStatFromCM(stat *Stat, cm *CM) {
	stat.OobError = cm.GetFalseRate()

	stat.OobErrorMean = rt.oobStatTotal.OobError /
		float64(len(rt.oobStats)+1)

	stat.TP = int64(cm.TP())
	stat.FP = int64(cm.FP())
	stat.TN = int64(cm.TN())
	stat.FN = int64(cm.FN())

	t := float64(stat.TP + stat.FN)
	if t == 0 {
		stat.TPRate = 0
	} else {
		stat.TPRate = float64(stat.TP) / t
	}

	t = float64(stat.FP + stat.TN)
	if t == 0 {
		stat.FPRate = 0
	} else {
		stat.FPRate = float64(stat.FP) / t
	}

	t = float64(stat.FP + stat.TN)
	if t == 0 {
		stat.TNRate = 0
	} else {
		stat.TNRate = float64(stat.TN) / t
	}

	t = float64(stat.TP + stat.FP)
	if t == 0 {
		stat.Precision = 0
	} else {
		stat.Precision = float64(stat.TP) / t
	}

	t = (1 / stat.Precision) + (1 / stat.TPRate)
	if t == 0 {
		stat.FMeasure = 0
	} else {
		stat.FMeasure = 2 / t
	}

	t = float64(stat.TP + stat.TN + stat.FP + stat.FN)
	if t == 0 {
		stat.Accuracy = 0
	} else {
		stat.Accuracy = float64(stat.TP+stat.TN) / t
	}
}

// ComputeStatTotal compute total statistic.
func (rt *Runtime) ComputeStatTotal(stat *Stat) {
	if stat == nil {
		return
	}

	nstat := len(rt.oobStats)
	if nstat == 0 {
		return
	}

	t := &rt.oobStatTotal

	t.OobError += stat.OobError
	t.OobErrorMean = t.OobError / float64(nstat)
	t.TP += stat.TP
	t.FP += stat.FP
	t.TN += stat.TN
	t.FN += stat.FN

	total := float64(t.TP + t.FN)
	if total == 0 {
		t.TPRate = 0
	} else {
		t.TPRate = float64(t.TP) / total
	}

	total = float64(t.FP + t.TN)
	if total == 0 {
		t.FPRate = 0
	} else {
		t.FPRate = float64(t.FP) / total
	}

	total = float64(t.FP + t.TN)
	if total == 0 {
		t.TNRate = 0
	} else {
		t.TNRate = float64(t.TN) / total
	}

	total = float64(t.TP + t.FP)
	if total == 0 {
		t.Precision = 0
	} else {
		t.Precision = float64(t.TP) / total
	}

	total = (1 / t.Precision) + (1 / t.TPRate)
	if total == 0 {
		t.FMeasure = 0
	} else {
		t.FMeasure = 2 / total
	}

	total = float64(t.TP + t.TN + t.FP + t.FN)
	if total == 0 {
		t.Accuracy = 0
	} else {
		t.Accuracy = float64(t.TP+t.TN) / total
	}
}

// OpenOOBStatsFile will open statistic file for output.
func (rt *Runtime) OpenOOBStatsFile() error {
	if rt.oobWriter != nil {
		_ = rt.CloseOOBStatsFile()
	}
	rt.oobWriter = &dsv.Writer{}
	return rt.oobWriter.OpenOutput(rt.OOBStatsFile)
}

// WriteOOBStat will write statistic of process to file.
func (rt *Runtime) WriteOOBStat(stat *Stat) error {
	if rt.oobWriter == nil {
		return nil
	}
	if stat == nil {
		return nil
	}
	return rt.oobWriter.WriteRawRow(stat.ToRow(), nil, nil)
}

// CloseOOBStatsFile will close statistics file for writing.
func (rt *Runtime) CloseOOBStatsFile() (e error) {
	if rt.oobWriter == nil {
		return
	}

	e = rt.oobWriter.Close()
	rt.oobWriter = nil

	return
}

// PrintOobStat will print the out-of-bag statistic to standard output.
func (rt *Runtime) PrintOobStat(stat *Stat, cm *CM) {
	fmt.Printf("%s OOB error rate: %.4f,"+
		" total: %.4f, mean %.4f, true rate: %.4f\n", tag,
		stat.OobError, rt.oobStatTotal.OobError,
		stat.OobErrorMean, cm.GetTrueRate())
}

// PrintStat will print statistic value to standard output.
func (rt *Runtime) PrintStat(stat *Stat) {
	if stat == nil {
		statslen := len(rt.oobStats)
		if statslen <= 0 {
			return
		}
		stat = rt.oobStats[statslen-1]
	}

	fmt.Printf("%s TPRate: %.4f, FPRate: %.4f,"+
		" TNRate: %.4f, precision: %.4f, f-measure: %.4f,"+
		" accuracy: %.4f\n", tag, stat.TPRate, stat.FPRate,
		stat.TNRate, stat.Precision, stat.FMeasure, stat.Accuracy)
}

// PrintStatTotal will print total statistic to standard output.
func (rt *Runtime) PrintStatTotal(st *Stat) {
	if st == nil {
		st = &rt.oobStatTotal
	}
	rt.PrintStat(st)
}

// Performance given an actuals class label and their probabilities, compute
// the performance statistic of classifier.
//
// Algorithm,
// (1) Sort the probabilities in descending order.
// (2) Sort the actuals and predicts using sorted index from probs
// (3) Compute tpr, fpr, precision
// (4) Write performance to file.
func (rt *Runtime) Performance(samples tabula.ClasetInterface,
	predicts []string, probs []float64,
) (
	perfs Stats,
) {
	// (1)
	actuals := samples.GetClassAsStrings()
	sortedListID := numbers.IntCreateSeq(0, len(probs)-1)
	floats64.InplaceMergesort(probs, sortedListID, 0, len(probs), false)

	// (2)
	libstrings.SortByIndex(&actuals, sortedListID)
	libstrings.SortByIndex(&predicts, sortedListID)

	// (3)
	rt.computePerfByProbs(samples, actuals, probs)

	return rt.perfs
}

func trapezoidArea(fp, fpprev, tp, tpprev int64) float64 {
	base := math.Abs(float64(fp - fpprev))
	heightAvg := float64(tp+tpprev) / float64(2.0)
	return base * heightAvg
}

// computePerfByProbs will compute classifier performance using probabilities
// or score `probs`.
//
// This currently only work for two class problem.
func (rt *Runtime) computePerfByProbs(samples tabula.ClasetInterface,
	actuals []string, probs []float64,
) {
	vs := samples.GetClassValueSpace()
	nactuals := slices.ToInt64(samples.Counts())
	nclass := libstrings.CountTokens(actuals, vs, false)

	pprev := math.Inf(-1)
	tp := int64(0)
	fp := int64(0)
	tpprev := int64(0)
	fpprev := int64(0)

	auc := float64(0)

	for x, p := range probs {
		if p != pprev {
			stat := Stat{}
			stat.SetTPRate(tp, nactuals[0])
			stat.SetFPRate(fp, nactuals[1])
			stat.SetPrecisionFromRate(nactuals[0], nactuals[1])

			auc += trapezoidArea(fp, fpprev, tp, tpprev)
			stat.SetAUC(auc)

			rt.perfs = append(rt.perfs, &stat)

			pprev = p
			tpprev = tp
			fpprev = fp
		}

		if actuals[x] == vs[0] {
			tp++
		} else {
			fp++
		}
	}

	stat := Stat{}
	stat.SetTPRate(tp, nactuals[0])
	stat.SetFPRate(fp, nactuals[1])
	stat.SetPrecisionFromRate(nactuals[0], nactuals[1])

	auc += trapezoidArea(fp, fpprev, tp, tpprev)
	auc /= float64(nclass[0] * nclass[1])
	stat.SetAUC(auc)

	rt.perfs = append(rt.perfs, &stat)

	if len(rt.perfs) >= 2 {
		// Replace the first stat with second stat, because of NaN
		// value on the first precision.
		rt.perfs[0] = rt.perfs[1]
	}
}

// WritePerformance will write performance data to file.
func (rt *Runtime) WritePerformance() error {
	return rt.perfs.Write(rt.PerfFile)
}
