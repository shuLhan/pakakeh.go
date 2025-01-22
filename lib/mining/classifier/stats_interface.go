// SPDX-FileCopyrightText: 2016 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package classifier

// ComputeFMeasures given array of precisions and recalls, compute F-measure
// of each instance and return it.
func ComputeFMeasures(precisions, recalls []float64) (fmeasures []float64) {
	// Get the minimum length of precision and recall.
	// This is to make sure that we are not looping out of range.
	minlen := len(precisions)
	recallslen := len(recalls)
	if recallslen < minlen {
		minlen = recallslen
	}

	for x := range minlen {
		f := 2 / ((1 / precisions[x]) + (1 / recalls[x]))
		fmeasures = append(fmeasures, f)
	}
	return
}

// ComputeAccuracies will compute and return accuracy from array of
// true-positive, false-positive, true-negative, and false-negative; using
// formula,
//
//	(tp + tn) / (tp + tn + tn + fn)
func ComputeAccuracies(tp, fp, tn, fn []int64) (accuracies []float64) {
	// Get minimum length of input, just to make sure we are not looping
	// out of range.
	minlen := len(tp)
	if len(fp) < len(tn) {
		minlen = len(fp)
	}
	if len(fn) < minlen {
		minlen = len(fn)
	}

	for x := range minlen {
		acc := float64(tp[x]+tn[x]) /
			float64(tp[x]+fp[x]+tn[x]+fn[x])
		accuracies = append(accuracies, acc)
	}
	return
}

// ComputeElapsedTimes will compute and return elapsed time between `start`
// and `end` timestamps.
func ComputeElapsedTimes(start, end []int64) (elaps []int64) {
	// Get minimum length.
	minlen := len(start)
	if len(end) < minlen {
		minlen = len(end)
	}

	for x := range minlen {
		elaps = append(elaps, end[x]-start[x])
	}
	return
}
