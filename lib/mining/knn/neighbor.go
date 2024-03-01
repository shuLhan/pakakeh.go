// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package knn

import (
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

// Neighbors is a mapping between sample and their distance.
// This type implement the sort interface.
type Neighbors struct {
	// rows contain pointer to rows.
	rows []*tabula.Row
	// Distance value
	distances []float64
}

// Rows return all rows.
func (neighbors *Neighbors) Rows() *[]*tabula.Row {
	return &neighbors.rows
}

// Row return pointer to row at index `idx`.
func (neighbors *Neighbors) Row(idx int) *tabula.Row {
	return neighbors.rows[idx]
}

// Distances return slice of distance of each neighbours.
func (neighbors *Neighbors) Distances() *[]float64 {
	return &neighbors.distances
}

// Distance return distance value at index `idx`.
func (neighbors *Neighbors) Distance(idx int) float64 {
	return neighbors.distances[idx]
}

// Len return the number of neighbors.
// This is for sort interface.
func (neighbors *Neighbors) Len() int {
	return len(neighbors.distances)
}

// Less return true if i < j.
// This is for sort interface.
func (neighbors *Neighbors) Less(i, j int) bool {
	return neighbors.distances[i] < neighbors.distances[j]
}

// Swap content of object in index i with index j.
// This is for sort interface.
func (neighbors *Neighbors) Swap(i, j int) {
	row := neighbors.rows[i]
	distance := neighbors.distances[i]

	neighbors.rows[i] = neighbors.rows[j]
	neighbors.distances[i] = neighbors.distances[j]

	neighbors.rows[j] = row
	neighbors.distances[j] = distance
}

// Add new neighbor.
func (neighbors *Neighbors) Add(row *tabula.Row, distance float64) {
	neighbors.rows = append(neighbors.rows, row)
	neighbors.distances = append(neighbors.distances, distance)
}

// SelectRange select all neighbors from index `start` to `end`.
// Return an empty set if start or end is out of range.
func (neighbors *Neighbors) SelectRange(start, end int) (newn Neighbors) {
	if start < 0 {
		return
	}

	if end > neighbors.Len() {
		return
	}

	for x := start; x < end; x++ {
		row := neighbors.rows[x]
		newn.Add(row, neighbors.distances[x])
	}
	return
}

// SelectWhere return all neighbors where row value at index `idx` is equal
// to string `val`.
func (neighbors *Neighbors) SelectWhere(idx int, val string) (newn Neighbors) {
	for x, row := range neighbors.rows {
		colval := (*row)[idx].String()

		if colval == val {
			newn.Add(row, neighbors.Distance(x))
		}
	}
	return
}

// Contain return true if `row` is in neighbors and their index, otherwise
// return false and -1.
func (neighbors *Neighbors) Contain(row *tabula.Row) (bool, int) {
	for x, xrow := range neighbors.rows {
		if xrow.IsEqual(row) {
			return true, x
		}
	}
	return false, -1
}

// Replace neighbor at index `idx` with new row and distance value.
func (neighbors *Neighbors) Replace(idx int, row *tabula.Row, distance float64) {
	if idx > len(neighbors.rows) {
		return
	}

	neighbors.rows[idx] = row
	neighbors.distances[idx] = distance
}
