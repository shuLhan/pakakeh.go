// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ints

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

const n = 10000

func generateRandomInts(data []int) {
	rand.Seed(time.Now().Unix())
	for x := 0; x < n; x++ {
		data[x] = rand.Intn(n)
	}
}

func BenchmarkIndirectSort(b *testing.B) {
	data := make([]int, n)
	generateRandomInts(data)
	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		IndirectSort(data, true)

		b.StopTimer()
		generateRandomInts(data)
		b.StartTimer()
	}
}

func BenchmarkStdSortInts(b *testing.B) {
	data := make([]int, n)
	generateRandomInts(data)
	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		sort.Ints(data)

		b.StopTimer()
		generateRandomInts(data)
		b.StartTimer()
	}
}
