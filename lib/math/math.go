// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2015 Mhd Sulhan <ms@kilabit.info>

// Package math provide generic functions working with math.
package math

import (
	"math"
)

// Factorial compute the factorial of n.
func Factorial(n int) (f int) {
	if n >= 0 {
		f = 1
	} else {
		f = -1
		n *= -1
	}

	for ; n > 0; n-- {
		f *= n
	}

	return f
}

// BinomialCoefficient or combination, compute number of picking k from
// n possibilities.
//
// Result is n! / ((n - k)! * k!)
func BinomialCoefficient(n int, k int) int {
	if k > n {
		return 0
	}
	return Factorial(n) / (Factorial(n-k) * Factorial(k))
}

// StirlingS2 The number of ways of partitioning a set of n elements into
// k nonempty sets (i.e., k set blocks), also called a Stirling set number.
//
// For example, the set {1,2,3} can be partitioned into three subsets in one way:
// {{1},{2},{3}}; into two subsets in three ways: {{1,2},{3}}, {{1,3},{2}},
// and {{1},{2,3}}; and into one subset in one way: {{1,2,3}}.
//
// Ref: http://mathworld.wolfram.com/StirlingNumberoftheSecondKind.html
func StirlingS2(n int, k int) int {
	if k == 1 || n == k {
		return 1
	}
	var sum int

	for i := 0; i <= k; i++ {
		sum += int(math.Pow(-1, float64(i))) *
			BinomialCoefficient(k, i) *
			int(math.Pow(float64(k-i), float64(n)))
	}

	return sum / Factorial(k)
}
