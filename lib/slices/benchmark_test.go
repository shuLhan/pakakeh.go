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
