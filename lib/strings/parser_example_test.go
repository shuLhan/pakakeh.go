// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package strings

import (
	"fmt"
	"strings"
)

func ExampleNewParser() {
	content := "[test]\nkey = value"
	p := NewParser(content, `=[]`)

	for {
		token, del := p.Read()
		token = strings.TrimSpace(token)
		fmt.Printf("%q %q\n", token, del)
		if del == 0 {
			break
		}
	}
	// Output:
	// "" '['
	// "test" ']'
	// "key" '='
	// "value" '\x00'
}

func ExampleParser_ReadNoSpace() {
	var (
		content = " 1 , \r\t\f, 2 , 3 , 4 , "
		p       = NewParser(content, `,`)

		tok string
		r   rune
	)
	for {
		tok, r = p.ReadNoSpace()
		fmt.Printf("%q\n", tok)
		if r == 0 {
			break
		}
	}
	// Output:
	// "1"
	// ""
	// "2"
	// "3"
	// "4"
	// ""
}
