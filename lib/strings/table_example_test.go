// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"fmt"
)

func ExamplePartition() {
	ss := []string{"a", "b", "c"}

	fmt.Println("Partition k=1:", Partition(ss, 1))
	fmt.Println("Partition k=2:", Partition(ss, 2))
	fmt.Println("Partition k=3:", Partition(ss, 3))

	// Output:
	// Partition k=1: [[[a b c]]]
	// Partition k=2: [[[b a] [c]] [[b] [c a]] [[b c] [a]]]
	// Partition k=3: [[[a] [b] [c]]]
}

func ExampleSinglePartition() {
	ss := []string{"a", "b", "c"}
	fmt.Println(SinglePartition(ss))
	// Output:
	// [[[a] [b] [c]]]
}

func ExampleTable_IsEqual() {
	table := Table{
		{{"a"}, {"b", "c"}},
		{{"b"}, {"a", "c"}},
		{{"c"}, {"a", "b"}},
	}
	fmt.Println(table.IsEqual(table))

	other := Table{
		{{"c"}, {"a", "b"}},
		{{"a"}, {"b", "c"}},
		{{"b"}, {"a", "c"}},
	}
	fmt.Println(table.IsEqual(other))

	other = Table{
		{{"a"}, {"b", "c"}},
		{{"b"}, {"a", "c"}},
	}
	fmt.Println(table.IsEqual(other))

	// Output:
	// true
	// true
	// false
}

func ExampleTable_JoinCombination() {
	table := Table{
		{{"a"}, {"b"}, {"c"}},
	}
	s := "X"

	fmt.Println(table.JoinCombination(s))
	// Output:
	// [[[a X] [b] [c]] [[a] [b X] [c]] [[a] [b] [c X]]]

}
