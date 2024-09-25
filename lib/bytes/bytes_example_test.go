// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bytes_test

import (
	"bytes"
	"fmt"
	"log"
	"math"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

func ExampleAppendInt16() {
	for _, v := range []int16{math.MinInt16, 0xab, 0xabc, math.MaxInt16} {
		out := libbytes.AppendInt16([]byte{}, v)
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
		out := libbytes.AppendInt32([]byte{}, v)
		fmt.Printf("%11d => %#x => %#v\n", v, v, out)
	}
	// Output:
	// -2147483648 => -0x80000000 => []byte{0x80, 0x0, 0x0, 0x0}
	//         171 => 0xab => []byte{0x0, 0x0, 0x0, 0xab}
	//        2748 => 0xabc => []byte{0x0, 0x0, 0xa, 0xbc}
	//  2147483647 => 0x7fffffff => []byte{0x7f, 0xff, 0xff, 0xff}
}

func ExampleAppendInt64() {
	var listInt64 = []int64{
		math.MaxInt64,
		1000,
		math.MinInt64,
	}

	var (
		v   int64
		out []byte
	)
	for _, v = range listInt64 {
		out = libbytes.AppendInt64(nil, v)
		fmt.Printf("%d\n=> %#x\n=> %#v\n", v, v, out)
	}
	// Output:
	// 9223372036854775807
	// => 0x7fffffffffffffff
	// => []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	// 1000
	// => 0x3e8
	// => []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0xe8}
	// -9223372036854775808
	// => -0x8000000000000000
	// => []byte{0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
}

func ExampleAppendUint16() {
	inputs := []uint16{0, 0xab, 0xabc, math.MaxInt16, math.MaxUint16}
	for _, v := range inputs {
		out := libbytes.AppendUint16([]byte{}, v)
		fmt.Printf("%5d => %#04x => %#02v\n", v, v, out)
	}

	v := inputs[4] + 1 // MaxUint16 + 1
	out := libbytes.AppendUint16([]byte{}, v)
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
		out := libbytes.AppendUint32([]byte{}, v)
		fmt.Printf("%11d => %#x => %#v\n", v, v, out)
	}

	v := inputs[4] + 2 // MaxUint32 + 2
	out := libbytes.AppendUint32([]byte{}, v)
	fmt.Printf("%11d => %#x => %#v\n", v, v, out)

	// Output:
	// 0 => 0x0 => []byte{0x0, 0x0, 0x0, 0x0}
	//         171 => 0xab => []byte{0x0, 0x0, 0x0, 0xab}
	//        2748 => 0xabc => []byte{0x0, 0x0, 0xa, 0xbc}
	//  2147483647 => 0x7fffffff => []byte{0x7f, 0xff, 0xff, 0xff}
	//  4294967295 => 0xffffffff => []byte{0xff, 0xff, 0xff, 0xff}
	//           1 => 0x1 => []byte{0x0, 0x0, 0x0, 0x1}
}

func ExampleAppendUint64() {
	var listUint64 = []uint64{
		math.MaxUint64,
		math.MaxInt64,
		1000,
	}

	var (
		v   uint64
		out []byte
	)
	for _, v = range listUint64 {
		out = libbytes.AppendUint64(nil, v)
		fmt.Printf("%d\n=> %#x\n=> %#v\n", v, v, out)
	}
	// Output:
	// 18446744073709551615
	// => 0xffffffffffffffff
	// => []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	// 9223372036854775807
	// => 0x7fffffffffffffff
	// => []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	// 1000
	// => 0x3e8
	// => []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0xe8}
}

func ExampleConcat() {
	fmt.Printf("%s\n", libbytes.Concat())
	fmt.Printf("%s\n", libbytes.Concat([]byte{}))
	fmt.Printf("%s\n", libbytes.Concat([]byte{}, []byte("B")))
	fmt.Printf("%s\n", libbytes.Concat("with []int:", []int{1, 2}))
	fmt.Printf("%s\n", libbytes.Concat([]byte("bytes"), " and ", []byte("string")))
	fmt.Printf("%s\n", libbytes.Concat([]byte("A"), 1, []int{2}, []byte{}, []byte("C")))
	// Output:
	//
	//
	// B
	// with []int:
	// bytes and string
	// AC
}

func ExampleCopy() {
	// Copying empty slice.
	org := []byte{}
	cp := libbytes.Copy(org)
	fmt.Printf("%d %q\n", len(cp), cp)

	org = []byte("slice of life")
	tmp := org
	cp = libbytes.Copy(org)
	fmt.Printf("%d %q\n", len(cp), cp)
	fmt.Printf("Original address == tmp address: %v\n", &org[0] == &tmp[0])
	fmt.Printf("Original address == copy address: %v\n", &org[0] == &cp[0])

	// Output:
	// 0 ""
	// 13 "slice of life"
	// Original address == tmp address: true
	// Original address == copy address: false
}

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

func ExampleParseHexDump() {
	var (
		in = []byte("0000000 6548 6c6c 2c6f 7720 726f 646c 0021")

		out []byte
		err error
	)

	out, err = libbytes.ParseHexDump(in, false)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, libbytes.TrimNull(out))
	// Output:
	// Hello, world!
}

func ExamplePrintHex() {
	title := "PrintHex"
	data := []byte("Hello, world !")
	libbytes.PrintHex(title, data, 5)
	// Output:
	// PrintHex
	//    0 - 48 65 6C 6C 6F || H e l l o
	//    5 - 2C 20 77 6F 72 || , . w o r
	//   10 - 6C 64 20 21    || l d . !
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

func ExampleDumpPrettyTable() {
	var (
		data = []byte{1, 2, 3, 'H', 'e', 'l', 'l', 'o', 254, 255}
		bb   bytes.Buffer
	)

	libbytes.DumpPrettyTable(&bb, `DumpPrettyTable`, data)
	fmt.Println(bb.String())
	// Output:
	// DumpPrettyTable
	//           |  0  1  2  3  4  5  6  7 | 01234567 |   0   1   2   3   4   5   6   7 |
	//           |  8  9  A  B  C  D  E  F | 89ABCDEF |   8   9   A   B   C   D   E   F |
	// 0x00000000| 01 02 03 48 65 6c 6c 6f | ...Hello |   1   2   3  72 101 108 108 111 |0
	// 0x00000008| fe ff                   | ..       | 254 255                         |8
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
