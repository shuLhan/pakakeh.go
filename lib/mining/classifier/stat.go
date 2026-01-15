// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// Copyright 2015-2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.

package classifier

import (
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

// Stat hold statistic value of classifier, including TP rate, FP rate, precision,
// and recall.
type Stat struct {
	// ID unique id for this statistic (e.g. number of tree).
	ID int64

	// StartTime contain the start time of classifier in unix timestamp.
	StartTime int64

	// EndTime contain the end time of classifier in unix timestamp.
	EndTime int64

	// ElapsedTime contain actual time, in seconds, between end and start
	// time.
	ElapsedTime int64

	// TP contain true-positive value.
	TP int64

	// FP contain false-positive value.
	FP int64

	// TN contain true-negative value.
	TN int64

	// FN contain false-negative value.
	FN int64

	// OobError contain out-of-bag error.
	OobError float64

	// OobErrorMean contain mean of out-of-bag error.
	OobErrorMean float64

	// TPRate contain true-positive rate (recall): tp/(tp+fn)
	TPRate float64

	// FPRate contain false-positive rate: fp/(fp+tn)
	FPRate float64

	// TNRate contain true-negative rate: tn/(tn+fp)
	TNRate float64

	// Precision contain: tp/(tp+fp)
	Precision float64

	// FMeasure contain value of F-measure or the harmonic mean of
	// precision and recall.
	FMeasure float64

	// Accuracy contain the degree of closeness of measurements of a
	// quantity to that quantity's true value.
	Accuracy float64

	// AUC contain the area under curve.
	AUC float64
}

// SetAUC will set the AUC value.
func (stat *Stat) SetAUC(v float64) {
	stat.AUC = v
}

// SetTPRate will set TP and TPRate using number of positive `p`.
func (stat *Stat) SetTPRate(tp, p int64) {
	stat.TP = tp
	stat.TPRate = float64(tp) / float64(p)
}

// SetFPRate will set FP and FPRate using number of negative `n`.
func (stat *Stat) SetFPRate(fp, n int64) {
	stat.FP = fp
	stat.FPRate = float64(fp) / float64(n)
}

// SetPrecisionFromRate will set Precision value using tprate and fprate.
// `p` and `n` is the number of positive and negative class in samples.
func (stat *Stat) SetPrecisionFromRate(p, n int64) {
	stat.Precision = (stat.TPRate * float64(p)) /
		((stat.TPRate * float64(p)) + (stat.FPRate * float64(n)))
}

// Recall return value of recall.
func (stat *Stat) Recall() float64 {
	return stat.TPRate
}

// Sum will add statistic from other stat object to current stat, not including
// the start and end time.
func (stat *Stat) Sum(other *Stat) {
	stat.OobError += other.OobError
	stat.OobErrorMean += other.OobErrorMean
	stat.TP += other.TP
	stat.FP += other.FP
	stat.TN += other.TN
	stat.FN += other.FN
	stat.TPRate += other.TPRate
	stat.FPRate += other.FPRate
	stat.TNRate += other.TNRate
	stat.Precision += other.Precision
	stat.FMeasure += other.FMeasure
	stat.Accuracy += other.Accuracy
}

// ToRow will convert the stat to tabula.row in the order of Stat field.
func (stat *Stat) ToRow() (row *tabula.Row) {
	row = &tabula.Row{}

	row.PushBack(tabula.NewRecordInt(stat.ID))
	row.PushBack(tabula.NewRecordInt(stat.StartTime))
	row.PushBack(tabula.NewRecordInt(stat.EndTime))
	row.PushBack(tabula.NewRecordInt(stat.ElapsedTime))
	row.PushBack(tabula.NewRecordReal(stat.OobError))
	row.PushBack(tabula.NewRecordReal(stat.OobErrorMean))
	row.PushBack(tabula.NewRecordInt(stat.TP))
	row.PushBack(tabula.NewRecordInt(stat.FP))
	row.PushBack(tabula.NewRecordInt(stat.TN))
	row.PushBack(tabula.NewRecordInt(stat.FN))
	row.PushBack(tabula.NewRecordReal(stat.TPRate))
	row.PushBack(tabula.NewRecordReal(stat.FPRate))
	row.PushBack(tabula.NewRecordReal(stat.TNRate))
	row.PushBack(tabula.NewRecordReal(stat.Precision))
	row.PushBack(tabula.NewRecordReal(stat.FMeasure))
	row.PushBack(tabula.NewRecordReal(stat.Accuracy))
	row.PushBack(tabula.NewRecordReal(stat.AUC))

	return
}

// Start will start the timer.
func (stat *Stat) Start() {
	stat.StartTime = time.Now().Unix()
}

// End will stop the timer and compute the elapsed time.
func (stat *Stat) End() {
	stat.EndTime = time.Now().Unix()
	stat.ElapsedTime = stat.EndTime - stat.StartTime
}

// Write will write the content of stat to `file`.
func (stat *Stat) Write(file string) (e error) {
	if file == "" {
		return
	}

	writer := &dsv.Writer{}
	e = writer.OpenOutput(file)
	if e != nil {
		return e
	}

	e = writer.WriteRawRow(stat.ToRow(), nil, nil)
	if e != nil {
		return e
	}

	return writer.Close()
}
