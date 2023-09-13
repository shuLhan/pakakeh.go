// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ints

import (
	"crypto/rand"
	"log"
	"math"
	"math/big"
	"sort"
	"testing"
)

const n = 10000

func generateRandomInts(data []int) {
	var (
		max   = big.NewInt(math.MaxInt)
		randv *big.Int
		err   error
	)
	for x := 0; x < n; x++ {
		randv, err = rand.Int(rand.Reader, max)
		if err != nil {
			log.Fatalf(`generateRandomInts: %s`, err)
		}
		data[x] = int(randv.Int64())
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
