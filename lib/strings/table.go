// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"

	"github.com/shuLhan/share/lib/debug"
)

// Table is for working with set of row.
//
// Each element in table is in the form of
//
//	[
//		[["a"],["b","c"],...], // Row
//		[["x"],["y",z"],...]   // Row
//	]
type Table []Row

// createIndent create n space indentation and return it.
func createIndent(n int) (s string) {
	for i := 0; i < n; i++ {
		s += " "
	}
	return
}

// Partition group the each element of slice "ss" into non-empty
// record, in such a way that every element is included in one and only of the
// record.
//
// Given a list of element in "ss", and number of partition "k", return
// the set of all group of all elements without duplication.
//
// Number of possible list can be computed using Stirling number of second kind.
//
// For more information see,
//
//   - https://en.wikipedia.org/wiki/Partition_of_a_set
func Partition(ss []string, k int) (table Table) {
	n := len(ss)
	seed := make([]string, n)
	copy(seed, ss)

	if debug.Value >= 1 {
		fmt.Printf("lib/strings: %s Partition(%v,%v)\n", createIndent(n), n, k)
	}

	// if only one split return the set contain only seed as list.
	// input: {a,b,c},  output: {{a,b,c}}
	if k == 1 {
		list := make(Row, 1)
		list[0] = seed

		table = append(table, list)
		return table
	}

	// if number of element in set equal with number split, return the set
	// that contain each element in list.
	// input: {a,b,c},  output:= {{a},{b},{c}}
	if n == k {
		return SinglePartition(seed)
	}

	// take the first element
	el := seed[0]

	// remove the first element from set
	seed = append(seed[:0], seed[1:]...)

	if debug.Value >= 1 {
		fmt.Printf("[tekstus] %s el: %s, seed: %s", createIndent(n), el, seed)
	}

	// generate child list
	genTable := Partition(seed, k)

	if debug.Value >= 1 {
		fmt.Printf("[tekstus] %s genTable join: %v", createIndent(n), genTable)
	}

	// join elemen with generated set
	table = genTable.JoinCombination(el)

	if debug.Value >= 1 {
		fmt.Printf("[tekstus] %s join %s      : %v\n", createIndent(n), el,
			table)
	}

	genTable = Partition(seed, k-1)

	if debug.Value >= 1 {
		fmt.Printf("[tesktus] %s genTable append: %s", createIndent(n), genTable)
	}

	for _, row := range genTable {
		list := make(Row, len(row))
		copy(list, row)
		list = append(list, []string{el})
		table = append(table, list)
	}

	if debug.Value >= 1 {
		fmt.Printf("[tesktus] %s append %v      : %v\n", createIndent(n), el,
			table)
	}

	return table
}

// SinglePartition create a table from a slice of string, where each element
// in slice become a single record.
func SinglePartition(ss []string) Table {
	table := make(Table, 0)
	row := make(Row, len(ss))

	for x := 0; x < len(ss); x++ {
		row[x] = []string{ss[x]}
	}

	table = append(table, row)

	return table
}

// IsEqual compare two table of string without regard to their order.
//
// Return true if both set is contains the same list, false otherwise.
func (table Table) IsEqual(other Table) bool {
	if len(table) != len(other) {
		return false
	}

	check := make([]bool, len(table))

	for x := 0; x < len(table); x++ {
		for y := 0; y < len(other); y++ {
			if table[x].IsEqual(other[y]) {
				check[x] = true
				break
			}
		}
	}

	for _, v := range check {
		if !v {
			return false
		}
	}
	return true
}

// JoinCombination for each row in table, generate new row and insert "s" into
// different record in different new row.
func (table Table) JoinCombination(s string) (tout Table) {
	for _, row := range table {
		for y := 0; y < len(row); y++ {
			newRow := make(Row, len(row))
			copy(newRow, row)
			newRow[y] = append(newRow[y], s)
			tout = append(tout, newRow)
		}
	}
	return
}
