// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"fmt"
)

func ExampleCutUntilToken() {
	line := []byte(`abc \def ghi`)

	cut, p, found := CutUntilToken(line, []byte("def"), 0, false)
	fmt.Printf("'%s' %d %t\n", cut, p, found)

	cut, p, found = CutUntilToken(line, []byte("def"), 0, true)
	fmt.Printf("'%s' %d %t\n", cut, p, found)

	cut, p, found = CutUntilToken(line, []byte("ef"), 0, true)
	fmt.Printf("'%s' %d %t\n", cut, p, found)

	cut, p, found = CutUntilToken(line, []byte("hi"), 0, true)
	fmt.Printf("'%s' %d %t\n", cut, p, found)

	// Output:
	// 'abc \' 8 true
	// 'abc def ghi' 12 false
	// 'abc \d' 8 true
	// 'abc \def g' 12 true
}

func ExampleEncloseRemove() {
	line := []byte(`[[ ABC ]] DEF`)
	leftcap := []byte(`[[`)
	rightcap := []byte(`]]`)

	got, changed := EncloseRemove(line, leftcap, rightcap)

	fmt.Printf("'%s' %t\n", got, changed)
	// Output: ' DEF' true
}

func ExampleEncloseToken() {
	line := []byte(`// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`)
	token := []byte(`"`)
	leftcap := []byte(`\`)
	rightcap := []byte(`_`)

	got, changed := EncloseToken(line, token, leftcap, rightcap)

	fmt.Printf("'%s' %t\n", got, changed)
	// Output:
	// '// Copyright 2016-2018 \"_Shulhan <ms@kilabit.info>\"_. All rights reserved.' true
}

func ExampleIndexes() {
	s := []byte("moo moomoo")
	token := []byte("moo")

	idxs := Indexes(s, token)
	fmt.Println(idxs)
	// Output:
	// [0 4 7]
}

func ExampleIsTokenAt() {
	line := []byte("Hello, world")
	token := []byte("world")
	token2 := []byte("worlds")
	tokenEmpty := []byte{}

	fmt.Printf("%t\n", IsTokenAt(line, tokenEmpty, 6))
	fmt.Printf("%t\n", IsTokenAt(line, token, 6))
	fmt.Printf("%t\n", IsTokenAt(line, token, 7))
	fmt.Printf("%t\n", IsTokenAt(line, token, 8))
	fmt.Printf("%t\n", IsTokenAt(line, token2, 8))
	// Output:
	// false
	// false
	// true
	// false
	// false
}

func ExampleSkipAfterToken() {
	line := []byte(`abc \def ghi`)

	p, found := SkipAfterToken(line, []byte("def"), 0, false)
	fmt.Printf("%d %t\n", p, found)

	p, found = SkipAfterToken(line, []byte("def"), 0, true)
	fmt.Printf("%d %t\n", p, found)

	p, found = SkipAfterToken(line, []byte("ef"), 0, true)
	fmt.Printf("%d %t\n", p, found)

	p, found = SkipAfterToken(line, []byte("hi"), 0, true)
	fmt.Printf("%d %t\n", p, found)

	// Output:
	// 8 true
	// 12 false
	// 8 true
	// 12 true
}

func ExampleSnippetByIndexes() {
	s := []byte("// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.")
	indexes := []int{3, 20, len(s) - 4}

	snippets := SnippetByIndexes(s, indexes, 5)
	for _, snip := range snippets {
		fmt.Printf("%s\n", snip)
	}
	// Output:
	// // Copyr
	// 18, Shulha
	// reserved.
}

func ExampleTokenFind() {
	line := []byte("// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.")
	token := []byte("right")

	at := TokenFind(line, token, 0)

	fmt.Printf("%d\n", at)
	// Output: 7
}
