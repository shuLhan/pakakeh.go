// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bytes_test

import (
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

func ExampleCutUntilToken() {
	text := []byte(`\\abc \def \deg`)

	cut, pos, found := libbytes.CutUntilToken(text, nil, 0, false)
	fmt.Printf("'%s' %d %t\n", cut, pos, found)

	cut, pos, found = libbytes.CutUntilToken(text, []byte("def"), 0, false)
	fmt.Printf("'%s' %d %t\n", cut, pos, found)

	cut, pos, found = libbytes.CutUntilToken(text, []byte("def"), 0, true)
	fmt.Printf("'%s' %d %t\n", cut, pos, found)

	cut, pos, found = libbytes.CutUntilToken(text, []byte("ef"), -1, true)
	fmt.Printf("'%s' %d %t\n", cut, pos, found)

	cut, pos, found = libbytes.CutUntilToken(text, []byte("hi"), 0, true)
	fmt.Printf("'%s' %d %t\n", cut, pos, found)

	// Output:
	// '\\abc \def \deg' -1 false
	// '\\abc \' 10 true
	// '\abc def \deg' 15 false
	// '\abc \d' 10 true
	// '\abc \def \deg' 15 false
}

func ExampleEncloseRemove() {
	text := []byte(`[[ A ]]-[[ B ]] C`)

	got, isCut := libbytes.EncloseRemove(text, []byte("[["), []byte("]]"))
	fmt.Printf("'%s' %t\n", got, isCut)

	got, isCut = libbytes.EncloseRemove(text, []byte("[["), []byte("}}"))
	fmt.Printf("'%s' %t\n", got, isCut)

	text = []byte(`// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`)

	got, isCut = libbytes.EncloseRemove(text, []byte("<"), []byte(">"))
	fmt.Printf("'%s' %t\n", got, isCut)

	got, isCut = libbytes.EncloseRemove(text, []byte(`"`), []byte(`"`))
	fmt.Printf("'%s' %t\n", got, isCut)

	got, isCut = libbytes.EncloseRemove(text, []byte(`/`), []byte(`/`))
	fmt.Printf("'%s' %t\n", got, isCut)

	text = []byte(`/* TEST */`)

	got, isCut = libbytes.EncloseRemove(text, []byte(`/*`), []byte(`*/`))
	fmt.Printf("'%s' %t\n", got, isCut)

	// Output:
	// '- C' true
	// '[[ A ]]-[[ B ]] C' false
	// '// Copyright 2016-2018 "Shulhan ". All rights reserved.' true
	// '// Copyright 2016-2018 . All rights reserved.' true
	// ' Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.' true
	// '' true
}

func ExampleEncloseToken() {
	text := []byte(`// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`)

	got, isChanged := libbytes.EncloseToken(text, []byte(`"`), []byte(`\`), []byte(`_`))
	fmt.Printf("%t '%s'\n", isChanged, got)

	got, isChanged = libbytes.EncloseToken(text, []byte(`_`), []byte(`-`), []byte(`-`))
	fmt.Printf("%t '%s'\n", isChanged, got)

	got, isChanged = libbytes.EncloseToken(text, []byte(`/`), []byte(`\`), nil)
	fmt.Printf("%t '%s'\n", isChanged, got)

	got, isChanged = libbytes.EncloseToken(text, []byte(`<`), []byte(`<`), []byte(` `))
	fmt.Printf("%t '%s'\n", isChanged, got)

	// Output:
	// true '// Copyright 2016-2018 \"_Shulhan <ms@kilabit.info>\"_. All rights reserved.'
	// false '// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.'
	// true '\/\/ Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.'
	// true '// Copyright 2016-2018 "Shulhan << ms@kilabit.info>". All rights reserved.'
}

func ExampleInReplace() {
	text := libbytes.InReplace([]byte{}, []byte(ascii.LettersNumber), '_')
	fmt.Printf("%q\n", text)

	text = libbytes.InReplace([]byte("/a/b/c"), []byte(ascii.LettersNumber), '_')
	fmt.Printf("%q\n", text)

	_ = libbytes.InReplace(text, []byte(ascii.LettersNumber), '/')
	fmt.Printf("%q\n", text)

	// Output:
	// ""
	// "_a_b_c"
	// "/a/b/c"
}

func ExampleIndexes() {
	fmt.Println(libbytes.Indexes([]byte(""), []byte("moo")))
	fmt.Println(libbytes.Indexes([]byte("moo moomoo"), []byte{}))
	fmt.Println(libbytes.Indexes([]byte("moo moomoo"), []byte("moo")))
	// Output:
	// []
	// []
	// [0 4 7]
}

func ExampleIsTokenAt() {
	text := []byte("Hello, world")
	tokenWorld := []byte("world")
	tokenWorlds := []byte("worlds")
	tokenEmpty := []byte{}

	fmt.Printf("%t\n", libbytes.IsTokenAt(text, tokenEmpty, 6))
	fmt.Printf("%t\n", libbytes.IsTokenAt(text, tokenWorld, -1))
	fmt.Printf("%t\n", libbytes.IsTokenAt(text, tokenWorld, 6))
	fmt.Printf("%t\n", libbytes.IsTokenAt(text, tokenWorld, 7))
	fmt.Printf("%t\n", libbytes.IsTokenAt(text, tokenWorld, 8))
	fmt.Printf("%t\n", libbytes.IsTokenAt(text, tokenWorlds, 8))
	// Output:
	// false
	// false
	// false
	// true
	// false
	// false
}

func ExampleMergeSpaces() {
	fmt.Printf("%s\n", libbytes.MergeSpaces([]byte("")))
	fmt.Printf("%s\n", libbytes.MergeSpaces([]byte(" \t\v\r\n\r\n\fa \t\v\r\n\r\n\f")))
	// Output:
	//
	//  a
}

func ExampleReadHexByte() {
	fmt.Println(libbytes.ReadHexByte([]byte{}, 0))
	fmt.Println(libbytes.ReadHexByte([]byte("x0"), 0))
	fmt.Println(libbytes.ReadHexByte([]byte("00"), 0))
	fmt.Println(libbytes.ReadHexByte([]byte("01"), 0))
	fmt.Println(libbytes.ReadHexByte([]byte("10"), 0))
	fmt.Println(libbytes.ReadHexByte([]byte("1A"), 0))
	fmt.Println(libbytes.ReadHexByte([]byte("1a"), 0))
	fmt.Println(libbytes.ReadHexByte([]byte("1a"), -1))
	// Output:
	// 0 false
	// 0 false
	// 0 true
	// 1 true
	// 16 true
	// 26 true
	// 26 true
	// 0 false
}

func ExampleReadInt16() {
	fmt.Println(libbytes.ReadInt16([]byte{0x01, 0x02, 0x03, 0x04}, 3)) // x is out of range.
	fmt.Println(libbytes.ReadInt16([]byte{0x01, 0x02, 0x03, 0x04}, 0)) // 0x0102
	fmt.Println(libbytes.ReadInt16([]byte{0x01, 0x02, 0xf0, 0x04}, 2)) // 0xf004
	// Output:
	// 0
	// 258
	// -4092
}

func ExampleReadInt32() {
	fmt.Println(libbytes.ReadInt32([]byte{0x01, 0x02, 0x03, 0x04}, 1)) // x is out of range.
	fmt.Println(libbytes.ReadInt32([]byte{0x01, 0x02, 0x03, 0x04}, 0)) // 0x01020304
	fmt.Println(libbytes.ReadInt32([]byte{0xf1, 0x02, 0x03, 0x04}, 0)) // 0xf1020304
	// Output:
	// 0
	// 16909060
	// -251526396
}

func ExampleReadUint16() {
	fmt.Println(libbytes.ReadUint16([]byte{0x01, 0x02, 0xf0, 0x04}, 3)) // x is out of range.
	fmt.Println(libbytes.ReadUint16([]byte{0x01, 0x02, 0xf0, 0x04}, 0)) // 0x0102
	fmt.Println(libbytes.ReadUint16([]byte{0x01, 0x02, 0xf0, 0x04}, 2)) // 0xf004
	// Output:
	// 0
	// 258
	// 61444
}

func ExampleReadUint32() {
	fmt.Println(libbytes.ReadUint32([]byte{0x01, 0x02, 0x03, 0x04}, 1)) // x is out of range.
	fmt.Println(libbytes.ReadUint32([]byte{0x01, 0x02, 0x03, 0x04}, 0)) // 0x01020304
	fmt.Println(libbytes.ReadUint32([]byte{0xf1, 0x02, 0x03, 0x04}, 0)) // 0xf1020304
	// Output:
	// 0
	// 16909060
	// 4043440900
}

func ExampleRemoveSpaces() {
	var (
		in  = []byte(" a\nb\tc d\r")
		out = libbytes.RemoveSpaces(in)
	)

	fmt.Printf("%s\n", out)

	// Output:
	// abcd
}

func ExampleSkipAfterToken() {
	text := []byte(`abc \def ghi`)

	fmt.Println(libbytes.SkipAfterToken(text, []byte("def"), -1, false))
	fmt.Println(libbytes.SkipAfterToken(text, []byte("def"), 0, true))
	fmt.Println(libbytes.SkipAfterToken(text, []byte("deg"), 0, false))
	fmt.Println(libbytes.SkipAfterToken(text, []byte("deg"), 0, true))
	fmt.Println(libbytes.SkipAfterToken(text, []byte("ef"), 0, true))
	fmt.Println(libbytes.SkipAfterToken(text, []byte("hi"), 0, true))
	// Output:
	// 8 true
	// -1 false
	// -1 false
	// -1 false
	// 8 true
	// 12 true
}

func ExampleSnippetByIndexes() {
	s := []byte("// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.")
	indexes := []int{3, 20, len(s) - 4}

	snippets := libbytes.SnippetByIndexes(s, indexes, 5)
	for _, snip := range snippets {
		fmt.Printf("%s\n", snip)
	}
	// Output:
	// // Copyr
	// 18, Shulha
	// reserved.
}

func ExampleSplitEach() {
	var data = []byte(`Hello`)

	fmt.Printf("%s\n", libbytes.SplitEach(data, 0))
	fmt.Printf("%s\n", libbytes.SplitEach(data, 1))
	fmt.Printf("%s\n", libbytes.SplitEach(data, 2))
	fmt.Printf("%s\n", libbytes.SplitEach(data, 5))
	fmt.Printf("%s\n", libbytes.SplitEach(data, 10))
	// Output:
	// [Hello]
	// [H e l l o]
	// [He ll o]
	// [Hello]
	// [Hello]
}

func ExampleTokenFind() {
	text := []byte("// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.")

	fmt.Println(libbytes.TokenFind(text, []byte{}, 0))
	fmt.Println(libbytes.TokenFind(text, []byte("right"), -1))
	fmt.Println(libbytes.TokenFind(text, []byte("."), 0))
	fmt.Println(libbytes.TokenFind(text, []byte("."), 42))
	fmt.Println(libbytes.TokenFind(text, []byte("."), 48))
	fmt.Println(libbytes.TokenFind(text, []byte("d."), 0))
	// Output:
	// -1
	// 7
	// 38
	// 44
	// 65
	// 64
}

func ExampleTrimNull() {
	var in = []byte{0, 'H', 'e', 'l', 'l', 'o', 0, 0}

	in = libbytes.TrimNull(in)
	fmt.Printf(`%s`, in)
	// Output: Hello
}

func ExampleWordIndexes() {
	text := []byte("moo moomoo moo")

	fmt.Println(libbytes.WordIndexes(text, []byte("mo")))
	fmt.Println(libbytes.WordIndexes(text, []byte("moo")))
	fmt.Println(libbytes.WordIndexes(text, []byte("mooo")))
	// Output:
	// []
	// [0 11]
	// []
}

func ExampleWriteUint16() {
	data := []byte("Hello, world!")

	var v uint16

	v = 'h' << 8
	v |= 'E'

	libbytes.WriteUint16(data, uint(len(data)-1), v) // Index out of range
	fmt.Println(string(data))

	libbytes.WriteUint16(data, 0, v)
	fmt.Println(string(data))
	// Output:
	// Hello, world!
	// hEllo, world!
}

func ExampleWriteUint32() {
	data := []byte("Hello, world!")

	var v uint32

	v = 'h' << 24
	v |= 'E' << 16
	v |= 'L' << 8
	v |= 'L'

	libbytes.WriteUint32(data, uint(len(data)-1), v) // Index out of range
	fmt.Println(string(data))

	libbytes.WriteUint32(data, 0, v)
	fmt.Println(string(data))

	libbytes.WriteUint32(data, 7, v)
	fmt.Println(string(data))
	// Output:
	// Hello, world!
	// hELLo, world!
	// hELLo, hELLd!
}
