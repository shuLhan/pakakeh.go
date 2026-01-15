// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

import (
	"testing"
)

func BenchmarkPushRow(b *testing.B) {
	dataset := NewDataset(DatasetModeRows, nil, nil)

	for i := 0; i < b.N; i++ {
		e := populateWithRows(dataset)
		if e != nil {
			b.Fatal(e)
		}
	}
}
