// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
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
