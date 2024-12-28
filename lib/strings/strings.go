// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package strings provide a library for working with string or slice of
// strings.
package strings

import (
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/slices"
)

// AppendUniq append case-insensitive strings to slice of input without
// duplicate.
func AppendUniq(in []string, vals ...string) []string {
	var found bool

	for x := 0; x < len(vals); x++ {
		found = false
		for y := 0; y < len(in); y++ {
			if vals[x] == in[y] {
				found = true
				break
			}
		}
		if found {
			continue
		}
		in = append(in, vals[x])
	}

	return in
}

// CountMissRate given two slice of string, count number of string that is
// not equal with each other, and return the miss rate as
//
//	number of not equal / number of data
//
// and count of missing, and length of input `src`.
func CountMissRate(src []string, target []string) (
	missrate float64,
	nmiss, length int,
) {
	length = len(src)
	targetlen := len(target)
	if length == 0 && targetlen == 0 {
		return
	}
	if targetlen < length {
		length = targetlen
	}

	for x := 0; x < length; x++ {
		if src[x] != target[x] {
			nmiss++
		}
	}

	return float64(nmiss) / float64(length), nmiss, length
}

// CountToken will return number of token occurrence in words.
func CountToken(words []string, token string, sensitive bool) int {
	if !sensitive {
		token = strings.ToLower(token)
	}

	var cnt int
	for _, v := range words {
		if !sensitive {
			v = strings.ToLower(v)
		}

		if v == token {
			cnt++
		}
	}
	return cnt
}

// CountTokens count number of occurrence of each `tokens` values in words.
// Return number of each tokens based on their index.
func CountTokens(words []string, tokens []string, sensitive bool) []int {
	tokenslen := len(tokens)
	if tokenslen <= 0 {
		return nil
	}

	counters := make([]int, tokenslen)

	for x := 0; x < len(tokens); x++ {
		counters[x] = CountToken(words, tokens[x], sensitive)
	}

	return counters
}

// Delete the first item that match with value while still preserving the
// order.
// It will return true if there is an item being deleted on slice, otherwise
// it will return false.
func Delete(in []string, value string) (out []string, ok bool) {
	x := 0
	for ; x < len(in); x++ {
		if in[x] == value {
			ok = true
			break
		}
	}
	if !ok {
		return in, false
	}

	copy(in[x:], in[x+1:])
	in[len(in)-1] = ""
	out = in[:len(in)-1]

	return out, true
}

// FrequencyOfToken return frequency of token in words using
//
//	count-of-token / total-words
func FrequencyOfToken(words []string, token string, sensitive bool) float64 {
	wordslen := float64(len(words))
	if wordslen <= 0 {
		return 0
	}

	cnt := CountToken(words, token, sensitive)

	return float64(cnt) / wordslen
}

// FrequencyOfTokens will compute each frequency of token in words.
func FrequencyOfTokens(words, tokens []string, sensitive bool) (probs []float64) {
	if len(words) == 0 || len(tokens) == 0 {
		return
	}

	probs = make([]float64, len(tokens))

	for x := 0; x < len(tokens); x++ {
		probs[x] = FrequencyOfToken(words, tokens[x], sensitive)
	}

	return probs
}

// IsContain return true if elemen `el` is in slice of string `ss`,
// otherwise return false.
func IsContain(ss []string, el string) bool {
	for x := 0; x < len(ss); x++ {
		if ss[x] == el {
			return true
		}
	}
	return false
}

// IsEqual compare elements of two slice of string without regard to
// their order.
//
// Return true if each both slice have the same elements, false otherwise.
func IsEqual(a, b []string) bool {
	alen := len(a)

	if alen != len(b) {
		return false
	}

	check := make([]bool, alen)

	for x, ls := range a {
		for _, rs := range b {
			if ls == rs {
				check[x] = true
			}
		}
	}

	for _, v := range check {
		if !v {
			return false
		}
	}
	return true
}

// Longest find the longest word in words and return their value and index.
//
// If words is empty return nil string with negative (-1) index.
func Longest(words []string) (string, int) {
	if len(words) == 0 {
		return "", -1
	}

	var (
		outlen, idx int
		out         string
	)
	for x := 0; x < len(words); x++ {
		vlen := len(words[x])
		if vlen > outlen {
			outlen = vlen
			out = words[x]
			idx = x
		}
	}
	return out, idx
}

// MostFrequentTokens return the token that has highest frequency in words.
//
// For example, given input
//
//	words:  [A A B A B C C]
//	tokens: [A B]
//
// it will return A as the majority tokens in words.
// If tokens has equal frequency, then the first token in order will returned.
func MostFrequentTokens(words []string, tokens []string, sensitive bool) string {
	if len(words) == 0 || len(tokens) == 0 {
		return ""
	}

	tokensCount := CountTokens(words, tokens, sensitive)
	_, maxIdx := slices.Max2(tokensCount)

	return tokens[maxIdx]
}

// SortByIndex will sort the slice of string in place using list of index.
func SortByIndex(ss *[]string, sortedListID []int) {
	newd := make([]string, len(*ss))

	for x := 0; x < len(sortedListID); x++ {
		newd[x] = (*ss)[sortedListID[x]]
	}

	(*ss) = newd
}

// Swap two indices value of string.
// If x or y is less than zero, it will return unchanged slice.
// If x or y is greater than length of slice, it will return unchanged slice.
func Swap(ss []string, x, y int) {
	if x == y {
		return
	}
	if x < 0 || y < 0 {
		return
	}
	if x > len(ss) || y > len(ss) {
		return
	}

	ss[x], ss[y] = ss[y], ss[x]
}

// TotalFrequencyOfTokens return total frequency of list of token in words.
func TotalFrequencyOfTokens(words, tokens []string, sensitive bool) float64 {
	if len(words) == 0 || len(tokens) == 0 {
		return 0
	}

	var sumfreq float64

	for x := 0; x < len(tokens); x++ {
		sumfreq += FrequencyOfToken(words, tokens[x], sensitive)
	}

	return sumfreq
}

// Uniq remove duplicate string from `words`.  It modify the content of slice
// in words by replacing duplicate word with empty string ("") and return only
// unique words.
// If sensitive is true then compare the string with case sensitive.
func Uniq(words []string, sensitive bool) (uniques []string) {
	var xcmp, ycmp string

	for x := 0; x < len(words); x++ {
		if len(words[x]) == 0 {
			continue
		}

		if sensitive {
			xcmp = words[x]
		} else {
			xcmp = strings.ToLower(words[x])
		}

		for y := x + 1; y < len(words); y++ {
			if len(words[y]) == 0 {
				continue
			}

			if sensitive {
				ycmp = words[y]
			} else {
				ycmp = strings.ToLower(words[y])
			}

			if xcmp == ycmp {
				words[y] = ""
			}
		}

		uniques = append(uniques, words[x])
	}

	return uniques
}
