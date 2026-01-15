// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package html

import (
	"fmt"
	"log"
	"strings"
)

func ExampleParse() {
	rawHTML := `
<ul>
	<li>
		<b>item</b>
		<span>one</span>
	</li>
</ul>
`

	r := strings.NewReader(rawHTML)

	iter, err := Parse(r)
	if err != nil {
		log.Fatal(err)
	}

	for node := iter.Next(); node != nil; node = iter.Next() {
		if node.IsElement() {
			fmt.Printf("%s\n", node.Data)
		} else {
			fmt.Printf("\t%s\n", node.Data)
		}
	}

	// Output:
	// html
	// head
	// body
	// ul
	// li
	// b
	// 	item
	// b
	// span
	// 	one
	// span
	// li
	// ul
	// body
	// html
}

func ExampleNodeIterator_SetNext() {
	rawHTML := `
<ul>
	<li>
		<b>item</b>
		<span>one</span>
	</li>
</ul>
<h2>Jump here</h2>
`

	r := strings.NewReader(rawHTML)

	iter, err := Parse(r)
	if err != nil {
		log.Fatal(err)
	}

	for node := iter.Next(); node != nil; node = iter.Next() {
		if node.IsElement() {
			if node.Data == "ul" {
				// Skip iterating the "ul" element.
				iter.SetNext(node.GetNextSibling())
				continue
			}
			fmt.Printf("%s\n", node.Data)
		} else {
			fmt.Printf("\t%s\n", node.Data)
		}
	}

	// Output:
	// html
	// head
	// body
	// h2
	// 	Jump here
	// h2
	// body
	// html
}
