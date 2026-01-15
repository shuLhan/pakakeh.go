// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2016 Mhd Sulhan <ms@kilabit.info>

package rf

import (
	"testing"
)

func BenchmarkPhoneme(b *testing.B) {
	SampleDsvFile = "../../testdata/phoneme/phoneme.dsv"
	OOBStatsFile = "phoneme.oob"
	PerfFile = "phoneme.perf"

	MinFeature = 3
	MaxFeature = 4

	for x := 0; x < b.N; x++ {
		runRandomForest()
	}
}
