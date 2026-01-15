// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package debug

import (
	"runtime"
)

// MemHeap store the difference between heap allocation.
type MemHeap struct {
	// RelHeapAlloc difference between heap allocation, relative to the
	// first time this object initialized.
	RelHeapAlloc int64
	// RelHeapObjects difference between heap objects, relative to the
	// first time this object initialized.
	RelHeapObjects int64

	// DiffHeapObject different between last heap objects and with current
	// relative heap objects.
	// This value is equal to MemStats.Mallocs - MemStats.Frees.
	// If its positive its means the number of objects allocated,
	// otherwise its represent the number of objects freed.
	DiffHeapObjects int64

	memStats         runtime.MemStats
	firstHeapAlloc   int64
	firstHeapObjects int64
	lastHeapObjects  int64
}

// NewMemHeap create and initialize MemStatsDiff for the first time.
func NewMemHeap() (memHeap *MemHeap) {
	firstStat := runtime.MemStats{}
	runtime.ReadMemStats(&firstStat)

	memHeap = &MemHeap{
		firstHeapAlloc:   int64(firstStat.HeapAlloc),
		firstHeapObjects: int64(firstStat.HeapObjects),
	}

	return
}

// Collect and compute the difference on the current heap allocation (in
// bytes) and heap objects.
func (msd *MemHeap) Collect() {
	runtime.ReadMemStats(&msd.memStats)

	msd.lastHeapObjects = msd.RelHeapObjects
	msd.RelHeapAlloc = int64(msd.memStats.HeapAlloc) - msd.firstHeapAlloc
	msd.RelHeapObjects = int64(msd.memStats.HeapObjects) - msd.firstHeapObjects
	msd.DiffHeapObjects = msd.RelHeapObjects - msd.lastHeapObjects
}
