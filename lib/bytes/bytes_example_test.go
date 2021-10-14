// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"fmt"
	"math"
)

func ExampleAppendInt16() {
	for _, v := range []int16{math.MinInt16, 0xab, 0xabc, math.MaxInt16} {
		out := AppendInt16([]byte{}, v)
		fmt.Printf("%6d => %#04x => %#02v\n", v, v, out)
	}
	// Output:
	// -32768 => -0x8000 => []byte{0x80, 0x00}
	//    171 => 0x00ab => []byte{0x00, 0xab}
	//   2748 => 0x0abc => []byte{0x0a, 0xbc}
	//  32767 => 0x7fff => []byte{0x7f, 0xff}
}

func ExampleAppendInt32() {
	for _, v := range []int32{math.MinInt32, 0xab, 0xabc, math.MaxInt32} {
		out := AppendInt32([]byte{}, v)
		fmt.Printf("%11d => %#x => %#v\n", v, v, out)
	}
	// Output:
	// -2147483648 => -0x80000000 => []byte{0x80, 0x0, 0x0, 0x0}
	//         171 => 0xab => []byte{0x0, 0x0, 0x0, 0xab}
	//        2748 => 0xabc => []byte{0x0, 0x0, 0xa, 0xbc}
	//  2147483647 => 0x7fffffff => []byte{0x7f, 0xff, 0xff, 0xff}
}

func ExampleAppendUint16() {
	inputs := []uint16{0, 0xab, 0xabc, math.MaxInt16, math.MaxUint16}
	for _, v := range inputs {
		out := AppendUint16([]byte{}, v)
		fmt.Printf("%5d => %#04x => %#02v\n", v, v, out)
	}

	v := inputs[4] + 1 // MaxUint16 + 1
	out := AppendUint16([]byte{}, v)
	fmt.Printf("%5d => %#04x => %#02v\n", v, v, out)

	// Output:
	// 0 => 0x0000 => []byte{0x00, 0x00}
	//   171 => 0x00ab => []byte{0x00, 0xab}
	//  2748 => 0x0abc => []byte{0x0a, 0xbc}
	// 32767 => 0x7fff => []byte{0x7f, 0xff}
	// 65535 => 0xffff => []byte{0xff, 0xff}
	//     0 => 0x0000 => []byte{0x00, 0x00}
}

func ExampleAppendUint32() {
	inputs := []uint32{0, 0xab, 0xabc, math.MaxInt32, math.MaxUint32}
	for _, v := range inputs {
		out := AppendUint32([]byte{}, v)
		fmt.Printf("%11d => %#x => %#v\n", v, v, out)
	}

	v := inputs[4] + 2 // MaxUint32 + 2
	out := AppendUint32([]byte{}, v)
	fmt.Printf("%11d => %#x => %#v\n", v, v, out)

	// Output:
	// 0 => 0x0 => []byte{0x0, 0x0, 0x0, 0x0}
	//         171 => 0xab => []byte{0x0, 0x0, 0x0, 0xab}
	//        2748 => 0xabc => []byte{0x0, 0x0, 0xa, 0xbc}
	//  2147483647 => 0x7fffffff => []byte{0x7f, 0xff, 0xff, 0xff}
	//  4294967295 => 0xffffffff => []byte{0xff, 0xff, 0xff, 0xff}
	//           1 => 0x1 => []byte{0x0, 0x0, 0x0, 0x1}
}

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

func ExampleWordIndexes() {
	s := []byte("moo moomoo moo")
	token := []byte("moo")

	idxs := WordIndexes(s, token)
	fmt.Println(idxs)
	// Output:
	// [0 11]
}
