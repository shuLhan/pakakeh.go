// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package slices_test

import (
	"sort"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/slices"
)

func BenchmarkSort_int(b *testing.B) {
	const n = 10_000

	var data = make([]int, n)
	generateRandomInts(data, n)

	var dataIndirect = make([]int, n)
	copy(dataIndirect, data)

	var dataInplaceMergesort = make([]int, n)
	var inplaceIdx = make([]int, n)
	copy(dataInplaceMergesort, data)

	var dataSortInts = make([]int, n)
	copy(dataSortInts, data)
	b.ResetTimer()

	b.Run(`sort.Ints`, func(b *testing.B) {
		for x := 0; x < b.N; x++ {
			sort.Ints(dataSortInts)
		}
	})

	b.Run(`IndirectSort`, func(b *testing.B) {
		for x := 0; x < b.N; x++ {
			slices.IndirectSort(dataIndirect, true)
		}
	})

	b.Run(`InplaceMergesort`, func(b *testing.B) {
		for x := 0; x < b.N; x++ {
			slices.InplaceMergesort(dataInplaceMergesort,
				inplaceIdx, 0, n, true)
		}
	})

}

func BenchmarkInplaceMergesort_float64(b *testing.B) {
	var slice = []float64{
		5.1, 5, 5.6, 5.5, 5.5, 5.8, 5.5, 5.5, 5.8, 5.6,
		5.7, 5, 5.6, 5.9, 6.2, 6, 4.9, 6.3, 6.1, 5.6,
		5.8, 6.7, 6.1, 5.9, 6, 4.9, 5.6, 5.2, 6.1, 6.4,
		7, 5.7, 6.5, 6.9, 5.7, 6.4, 6.2, 6.6, 6.3, 6.2,
		5.4, 6.7, 6.1, 5.7, 5.5, 6, 3, 6.6, 5.7, 6,
		6.8, 6, 6.1, 6.3, 5.8, 5.8, 5.6, 5.7, 6, 6.9,
		6.9, 6.4, 6.3, 6.3, 6.7, 6.5, 5.8, 6.3, 6.4, 6.7,
		5.9, 7.2, 6.3, 6.3, 6.5, 7.1, 6.7, 7.6, 7.3, 6.4,
		6.7, 7.4, 6, 6.8, 6.5, 6.4, 6.7, 6.4, 6.5, 6.9,
		7.7, 6.7, 7.2, 7.7, 7.2, 7.7, 6.1, 7.9, 7.7, 6.8,
		6.2,
	}
	var size = len(slice)
	var ids = make([]int, size)

	for i := 0; i < b.N; i++ {
		slices.InplaceMergesort(slice, ids, 0, size, true)
	}
}
