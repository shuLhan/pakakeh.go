// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package strings

// Table is for working with set of row.
//
// Each element in table is in the form of
//
//	[
//		[["a"],["b","c"],...], // Row
//		[["x"],["y",z"],...]   // Row
//	]
type Table []Row

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

	// generate child list
	genTable := Partition(seed, k)

	// join elemen with generated set
	table = genTable.JoinCombination(el)

	genTable = Partition(seed, k-1)

	for _, row := range genTable {
		list := make(Row, len(row))
		copy(list, row)
		list = append(list, []string{el})
		table = append(table, list)
	}

	return table
}

// SinglePartition create a table from a slice of string, where each element
// in slice become a single record.
func SinglePartition(ss []string) Table {
	table := make(Table, 0)
	row := make(Row, len(ss))

	for x := range len(ss) {
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

	for x := range len(table) {
		for y := range len(other) {
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
		for y := range len(row) {
			newRow := make(Row, len(row))
			copy(newRow, row)
			newRow[y] = append(newRow[y], s)
			tout = append(tout, newRow)
		}
	}
	return
}
