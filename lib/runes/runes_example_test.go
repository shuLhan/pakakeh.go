// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package runes

import (
	"fmt"
)

func ExampleContain() {
	line := []rune(`a b c`)
	found, idx := Contain(line, 'a')
	fmt.Printf("%t %d\n", found, idx)
	found, idx = Contain(line, 'x')
	fmt.Printf("%t %d\n", found, idx)
	// Output:
	// true 0
	// false -1
}

func ExampleDiff() {
	l := []rune{'a', 'b', 'c', 'd'}
	r := []rune{'b', 'c'}
	fmt.Printf("%c\n", Diff(l, r))
	// Output: [a d]
}

func ExampleEncloseRemove() {
	line := []rune(`[[ ABC ]] DEF`)
	leftcap := []rune(`[[`)
	rightcap := []rune(`]]`)

	got, changed := EncloseRemove(line, leftcap, rightcap)

	fmt.Printf("'%s' %t\n", string(got), changed)
	// Output: ' DEF' true
}

func ExampleFindSpace() {
	line := []rune(`Find a space`)
	fmt.Printf("%d\n", FindSpace(line, 0))
	fmt.Printf("%d\n", FindSpace(line, 5))
	// Output:
	// 4
	// 6
}

func ExampleInverse() {
	fmt.Printf("%s\n", string(Inverse([]rune(``))))
	fmt.Printf("%s\n", string(Inverse([]rune(`a`))))
	fmt.Printf("%s\n", string(Inverse([]rune(`ab`))))
	fmt.Printf("%s\n", string(Inverse([]rune(`abc`))))
	fmt.Printf("%s\n", string(Inverse([]rune(`abcd`))))
	fmt.Printf("%s\n", string(Inverse([]rune(`abcde`))))
	// Output:
	//
	// a
	// ba
	// cba
	// dcba
	// edcba
}

func ExampleTokenFind() {
	line := []rune("// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.")
	token := []rune("right")

	at := TokenFind(line, token, 0)

	fmt.Printf("%d\n", at)
	// Output: 7
}
