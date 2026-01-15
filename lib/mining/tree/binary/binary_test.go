// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2015 Mhd Sulhan <ms@kilabit.info>

package binary

import (
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {
	exp := `1
	12
		24
			34
			33
		23
	11
		22
			32
			31
		21
`

	btree := NewTree()

	root := NewBTNode(1,
		NewBTNode(11,
			NewBTNode(21, nil, nil),
			NewBTNode(22,
				NewBTNode(31, nil, nil),
				NewBTNode(32, nil, nil))),
		NewBTNode(12,
			NewBTNode(23, nil, nil),
			NewBTNode(24,
				NewBTNode(33, nil, nil),
				NewBTNode(34, nil, nil))))

	btree.Root = root

	res := fmt.Sprint(btree)

	if res != exp {
		t.Fatal("error, expecting:\n", exp, "\n got:\n", res)
	}
}
