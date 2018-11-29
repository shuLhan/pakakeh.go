// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package classifier

import (
	"fmt"
	"strconv"

	"github.com/shuLhan/share/lib/tabula"
)

//
// CM represent the matrix of classification.
//
type CM struct {
	tabula.Dataset
	// rowNames contain name in each row.
	rowNames []string
	// nSamples contain number of class.
	nSamples int64
	// nTrue contain number of true positive and negative.
	nTrue int64
	// nFalse contain number of false positive and negative.
	nFalse int64

	// tpIds contain index of true-positive samples.
	tpIds []int
	// fpIds contain index of false-positive samples.
	fpIds []int
	// tnIds contain index of true-negative samples.
	tnIds []int
	// fnIds contain index of false-negative samples.
	fnIds []int
}

//
// initByNumeric will initialize confusion matrix using numeric value space.
//
func (cm *CM) initByNumeric(vs []int64) {
	colTypes := make([]int, 0, len(vs))
	colNames := make([]string, 0, len(vs))

	for _, v := range vs {
		vstr := strconv.FormatInt(v, 10)
		colTypes = append(colTypes, tabula.TInteger)
		colNames = append(colNames, vstr)
		cm.rowNames = append(cm.rowNames, vstr)
	}

	// class error column
	colTypes = append(colTypes, tabula.TReal)
	colNames = append(colNames, "class_error")

	cm.Dataset.Init(tabula.DatasetModeMatrix, colTypes, colNames)
}

//
// ComputeStrings will calculate confusion matrix using targets and predictions
// class values.
//
func (cm *CM) ComputeStrings(valueSpace, targets, predictions []string) {
	cm.init(valueSpace)

	for x, target := range valueSpace {
		col := cm.GetColumn(x)

		for _, predict := range valueSpace {
			cnt := cm.countTargetPrediction(target, predict,
				targets, predictions)

			col.PushBack(tabula.NewRecordInt(cnt))
		}

		cm.PushColumnToRows(*col)
	}

	cm.computeClassError()
}

//
// ComputeNumeric will calculate confusion matrix using targets and predictions
// values.
//
func (cm *CM) ComputeNumeric(vs, actuals, predictions []int64) {
	cm.initByNumeric(vs)

	for x, act := range vs {
		col := cm.GetColumn(x)

		for _, pred := range vs {
			cnt := cm.countNumeric(act, pred, actuals, predictions)

			rec := tabula.NewRecordInt(cnt)
			col.PushBack(rec)
		}

		cm.PushColumnToRows(*col)
	}

	cm.computeClassError()
}

//
// create will initialize confusion matrix using value space.
//
func (cm *CM) init(valueSpace []string) {
	colTypes := make([]int, 0, len(valueSpace))
	colNames := make([]string, 0, len(valueSpace))

	for _, v := range valueSpace {
		colTypes = append(colTypes, tabula.TInteger)
		colNames = append(colNames, v)
		cm.rowNames = append(cm.rowNames, v)
	}

	// class error column
	colTypes = append(colTypes, tabula.TReal)
	colNames = append(colNames, "class_error")

	cm.Dataset.Init(tabula.DatasetModeMatrix, colTypes, colNames)
}

//
// countTargetPrediction will count and return number of true positive or false
// positive in predictions using targets values.
//
func (cm *CM) countTargetPrediction(target, predict string,
	targets, predictions []string,
) (
	cnt int64,
) {
	predictslen := len(predictions)

	for x, v := range targets {
		// In case out of range, where predictions length less than
		// targets length.
		if x >= predictslen {
			break
		}
		if v != target {
			continue
		}
		if predict == predictions[x] {
			cnt++
		}
	}
	return
}

//
// countNumeric will count and return number of `pred` in predictions where
// actuals value is `act`.
//
func (cm *CM) countNumeric(act, pred int64, actuals, predictions []int64) (
	cnt int64,
) {
	// Find minimum length to mitigate out-of-range loop.
	minlen := len(actuals)
	if len(predictions) < minlen {
		minlen = len(predictions)
	}

	for x := 0; x < minlen; x++ {
		if actuals[x] != act {
			continue
		}
		if predictions[x] != pred {
			continue
		}
		cnt++
	}
	return cnt
}

//
// computeClassError will compute the classification error in matrix.
//
func (cm *CM) computeClassError() {
	var tp, fp int64

	cm.nSamples = 0
	cm.nFalse = 0

	classcol := cm.GetNColumn() - 1
	col := cm.GetColumnClassError()
	rows := cm.GetDataAsRows()
	for x, row := range *rows {
		for y, cell := range *row {
			if y == classcol {
				break
			}
			if x == y {
				tp = cell.Integer()
			} else {
				fp += cell.Integer()
			}
		}

		nSamplePerRow := tp + fp
		errv := float64(fp) / float64(nSamplePerRow)
		col.PushBack(tabula.NewRecordReal(errv))

		cm.nSamples += nSamplePerRow
		cm.nTrue += tp
		cm.nFalse += fp
	}

	cm.PushColumnToRows(*col)
}

//
// GroupIndexPredictions given index of samples, group the samples by their
// class of prediction. For example,
//
//	sampleIds:   [0, 1, 2, 3, 4, 5]
//	actuals:     [1, 1, 0, 0, 1, 0]
//	predictions: [1, 0, 1, 0, 1, 1]
//
// This function will group the index by true-positive, false-positive,
// true-negative, and false-negative, which result in,
//
//	true-positive indices:  [0, 4]
//	false-positive indices: [2, 5]
//	true-negative indices:  [3]
//      false-negative indices: [1]
//
// This function assume that positive value as "1" and negative value as "0".
//
func (cm *CM) GroupIndexPredictions(sampleIds []int,
	actuals, predictions []int64,
) {
	// Reset indices.
	cm.tpIds = nil
	cm.fpIds = nil
	cm.tnIds = nil
	cm.fnIds = nil

	// Make sure we are not out-of-range when looping, always pick the
	// minimum length between the three parameters.
	min := len(sampleIds)
	if len(actuals) < min {
		min = len(actuals)
	}
	if len(predictions) < min {
		min = len(predictions)
	}

	for x := 0; x < min; x++ {
		if actuals[x] == 1 {
			if predictions[x] == 1 {
				cm.tpIds = append(cm.tpIds, sampleIds[x])
			} else {
				cm.fnIds = append(cm.fnIds, sampleIds[x])
			}
		} else {
			if predictions[x] == 1 {
				cm.fpIds = append(cm.fpIds, sampleIds[x])
			} else {
				cm.tnIds = append(cm.tnIds, sampleIds[x])
			}
		}
	}
}

//
// GroupIndexPredictionsStrings is an alternative to GroupIndexPredictions
// which work with string class.
//
func (cm *CM) GroupIndexPredictionsStrings(sampleIds []int,
	actuals, predictions []string,
) {
	if len(sampleIds) == 0 {
		return
	}

	// Reset indices.
	cm.tpIds = nil
	cm.fpIds = nil
	cm.tnIds = nil
	cm.fnIds = nil

	// Make sure we are not out-of-range when looping, always pick the
	// minimum length between the three parameters.
	min := len(sampleIds)
	if len(actuals) < min {
		min = len(actuals)
	}
	if len(predictions) < min {
		min = len(predictions)
	}

	for x := 0; x < min; x++ {
		if actuals[x] == "1" {
			if predictions[x] == "1" {
				cm.tpIds = append(cm.tpIds, sampleIds[x])
			} else {
				cm.fnIds = append(cm.fnIds, sampleIds[x])
			}
		} else {
			if predictions[x] == "1" {
				cm.fpIds = append(cm.fpIds, sampleIds[x])
			} else {
				cm.tnIds = append(cm.tnIds, sampleIds[x])
			}
		}
	}
}

//
// GetColumnClassError return the last column which is the column that contain
// the error of classification.
//
func (cm *CM) GetColumnClassError() *tabula.Column {
	return cm.GetColumn(cm.GetNColumn() - 1)
}

//
// GetTrueRate return true-positive rate in term of
//
//	true-positive / (true-positive + false-positive)
//
func (cm *CM) GetTrueRate() float64 {
	return float64(cm.nTrue) / float64(cm.nTrue+cm.nFalse)
}

//
// GetFalseRate return false-positive rate in term of,
//
//	false-positive / (false-positive + true negative)
//
func (cm *CM) GetFalseRate() float64 {
	return float64(cm.nFalse) / float64(cm.nTrue+cm.nFalse)
}

//
// TP return number of true-positive in confusion matrix.
//
func (cm *CM) TP() int {
	row := cm.GetRow(0)
	if row == nil {
		return 0
	}

	v, _ := row.GetIntAt(0)
	return int(v)
}

//
// FP return number of false-positive in confusion matrix.
//
func (cm *CM) FP() int {
	row := cm.GetRow(0)
	if row == nil {
		return 0
	}

	v, _ := row.GetIntAt(1)
	return int(v)
}

//
// FN return number of false-negative.
//
func (cm *CM) FN() int {
	row := cm.GetRow(1)
	if row == nil {
		return 0
	}
	v, _ := row.GetIntAt(0)
	return int(v)
}

//
// TN return number of true-negative.
//
func (cm *CM) TN() int {
	row := cm.GetRow(1)
	if row == nil {
		return 0
	}
	v, _ := row.GetIntAt(1)
	return int(v)
}

//
// TPIndices return indices of all true-positive samples.
//
func (cm *CM) TPIndices() []int {
	return cm.tpIds
}

//
// FNIndices return indices of all false-negative samples.
//
func (cm *CM) FNIndices() []int {
	return cm.fnIds
}

//
// FPIndices return indices of all false-positive samples.
//
func (cm *CM) FPIndices() []int {
	return cm.fpIds
}

//
// TNIndices return indices of all true-negative samples.
//
func (cm *CM) TNIndices() []int {
	return cm.tnIds
}

//
// String will return the output of confusion matrix in table like format.
//
func (cm *CM) String() (s string) {
	s += "Confusion Matrix:\n"

	// Row header: column names.
	s += "\t"
	for _, col := range cm.GetColumnsName() {
		s += col + "\t"
	}
	s += "\n"

	rows := cm.GetDataAsRows()
	for x, row := range *rows {
		s += cm.rowNames[x] + "\t"

		for _, v := range *row {
			s += v.String() + "\t"
		}
		s += "\n"
	}

	s += fmt.Sprintf("TP-FP indices %d %d\n", len(cm.tpIds), len(cm.fpIds))
	s += fmt.Sprintf("FN-TN indices %d %d\n", len(cm.fnIds), len(cm.tnIds))

	return
}
