// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"strings"
	"unicode"

	"git.sr.ht/~shulhan/pakakeh.go/lib/runes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/slices"
)

// CountAlnum return number of alpha-numeric character in text.
func CountAlnum(text string) (n int) {
	if len(text) == 0 {
		return
	}

	for _, v := range text {
		if unicode.IsDigit(v) || unicode.IsLetter(v) {
			n++
		}
	}
	return
}

// CountAlnumDistribution count distribution of alpha-numeric characters in
// text.
//
// Example, given a text "abbcccddddeeeee", it will return [a b c d e] and
// [1 2 3 4 5].
func CountAlnumDistribution(text string) (chars []rune, counts []int) {
	var found bool

	for _, v := range text {
		if !(unicode.IsDigit(v) || unicode.IsLetter(v)) {
			continue
		}
		found = false
		for y, c := range chars {
			if v == c {
				counts[y]++
				found = true
				break
			}
		}
		if !found {
			chars = append(chars, v)
			counts = append(counts, 1)
		}
	}
	return
}

// CountCharSequence given a string, count number of repeated character more
// than one in sequence and return list of characters and their counts.
func CountCharSequence(text string) (chars []rune, counts []int) {
	var lastv rune
	count := 1
	for _, v := range text {
		if v == lastv {
			if !unicode.IsSpace(v) {
				count++
			}
		} else {
			if count > 1 {
				chars = append(chars, lastv)
				counts = append(counts, count)
				count = 1
			}
		}
		lastv = v
	}
	if count > 1 {
		chars = append(chars, lastv)
		counts = append(counts, count)
	}
	return
}

// CountDigit return number of digit in text.
func CountDigit(text string) (n int) {
	if len(text) == 0 {
		return 0
	}

	for _, v := range text {
		if unicode.IsDigit(v) {
			n++
		}
	}
	return
}

// CountNonAlnum return number of non alpha-numeric character in text.
// If `withspace` is true, it will be counted as non-alpha-numeric, if it
// false it will be ignored.
func CountNonAlnum(text string, withspace bool) (n int) {
	if len(text) == 0 {
		return
	}

	for _, v := range text {
		if unicode.IsDigit(v) || unicode.IsLetter(v) {
			continue
		}
		if unicode.IsSpace(v) {
			if withspace {
				n++
			}
			continue
		}
		n++
	}
	return
}

// CountUniqChar count number of character in text without duplication.
func CountUniqChar(text string) (n int) {
	if len(text) == 0 {
		return
	}

	uchars := make([]rune, 0, len(text))

	for _, v := range text {
		yes, _ := runes.Contain(uchars, v)
		if yes {
			continue
		}
		uchars = append(uchars, v)
		n++
	}
	return
}

// CountUpperLower return number of uppercase and lowercase in text.
func CountUpperLower(text string) (upper, lower int) {
	for _, v := range text {
		if !unicode.IsLetter(v) {
			continue
		}
		if unicode.IsUpper(v) {
			upper++
		} else {
			lower++
		}
	}
	return
}

// MaxCharSequence return character which have maximum sequence in `text`.
func MaxCharSequence(text string) (rune, int) {
	if len(text) == 0 {
		return 0, 0
	}

	chars, counts := CountCharSequence(text)

	if len(chars) == 0 {
		return 0, 0
	}

	_, idx := slices.Max2(counts)

	return chars[idx], counts[idx]
}

// RatioAlnum compute and return ratio of alpha-numeric within all character
// in text.
func RatioAlnum(text string) float64 {
	textlen := len(text)
	if textlen == 0 {
		return 0
	}

	n := CountAlnum(text)

	return float64(n) / float64(textlen)
}

// RatioDigit compute and return digit ratio to all characters in text.
func RatioDigit(text string) float64 {
	textlen := len(text)

	if textlen == 0 {
		return 0
	}

	n := CountDigit(text)

	if n == 0 {
		return 0
	}

	return float64(n) / float64(textlen)
}

// RatioUpper compute and return ratio of uppercase character to all character
// in text.
func RatioUpper(text string) float64 {
	if len(text) == 0 {
		return 0
	}
	up, lo := CountUpperLower(text)

	total := up + lo
	if total == 0 {
		return 0
	}

	return float64(up) / float64(total)
}

// RatioNonAlnum return ratio of non-alphanumeric character to all
// character in text.
//
// If `withspace` is true then white-space character will be counted as
// non-alpha numeric, otherwise it will be skipped.
func RatioNonAlnum(text string, withspace bool) float64 {
	textlen := len(text)
	if textlen == 0 {
		return 0
	}

	n := CountNonAlnum(text, withspace)

	return float64(n) / float64(textlen)
}

// RatioUpperLower compute and return ratio of uppercase with lowercase
// character in text.
func RatioUpperLower(text string) float64 {
	if len(text) == 0 {
		return 0
	}

	up, lo := CountUpperLower(text)

	if lo == 0 {
		return float64(up)
	}

	return float64(up) / float64(lo)
}

// TextSumCountTokens given a text, count how many tokens inside of it and
// return sum of all counts.
func TextSumCountTokens(text string, tokens []string, sensitive bool) (
	cnt int,
) {
	if len(text) == 0 {
		return 0
	}

	if !sensitive {
		text = strings.ToLower(text)
	}

	for _, v := range tokens {
		if !sensitive {
			v = strings.ToLower(v)
		}
		cnt += strings.Count(text, v)
	}

	return
}

// TextFrequencyOfTokens return frequencies of tokens by counting each
// occurrence of token and divide it with total words in text.
func TextFrequencyOfTokens(text string, tokens []string, sensitive bool) (
	freq float64,
) {
	if len(text) == 0 {
		return 0
	}

	words := Split(text, false, false)
	wordsLen := float64(len(words))

	tokensCnt := float64(TextSumCountTokens(text, tokens, sensitive))

	freq = tokensCnt / wordsLen

	return
}
