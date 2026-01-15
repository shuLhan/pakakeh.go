// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2015 Mhd Sulhan <ms@kilabit.info>

package math

import (
	"testing"
)

func TestFactorial(t *testing.T) {
	in := []int{-3, -2, -1, 0, 1, 2, 3}
	exp := []int{-6, -2, -1, 1, 1, 2, 6}

	for i := range in {
		res := Factorial(in[i])

		if res != exp[i] {
			t.Fatal("Expecting ", exp[i], ", got ", res)
		}
	}
}

func TestBinomialCoefficient(t *testing.T) {
	in := [][]int{{1, 2}, {1, 1}, {3, 2}, {5, 3}}
	exp := []int{0, 1, 3, 10}

	for i := range in {
		res := BinomialCoefficient(in[i][0], in[i][1])

		if res != exp[i] {
			t.Fatal("Expecting ", exp[i], ", got ", res)
		}
	}
}

func TestStirlingS2(t *testing.T) {
	in := [][]int{{3, 1}, {3, 2}, {3, 3}, {5, 3}}
	exp := []int{1, 3, 1, 25}

	for i := range in {
		res := StirlingS2(in[i][0], in[i][1])

		if res != exp[i] {
			t.Fatal("Expecting ", exp[i], ", got ", res)
		}
	}
}
