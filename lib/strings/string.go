// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"bytes"
	"strings"
	"unicode"

	librunes "github.com/shuLhan/share/lib/runes"
)

// Alnum remove non alpha-numeric character from text and return it.
// If withSpace is true then white space is allowed, otherwise it would also
// be removed from text.
func Alnum(text string, withSpace bool) string {
	var buf bytes.Buffer
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		} else if withSpace && unicode.IsSpace(r) {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// CleanURI remove known links from text and return it.
// This function assume that space in URI is using '%20' not literal space,
// as in ' '.
//
// List of known links: http, https, ftp, ftps, ssh, file, rsync, and ws.
func CleanURI(text string) string {
	if len(text) == 0 {
		return ""
	}

	var URIPrefixes = []string{
		"http://", "https://", "ftp://", "ftps://", "ssh://", "file://", "rsync://", "ws://",
	}
	ctext := []rune(text)

	for _, uri := range URIPrefixes {
		startat := 0
		curi := []rune(uri)
		newtext := []rune{}

		for {
			begin := librunes.TokenFind(ctext, curi, startat)
			if begin < 0 {
				if startat > 0 {
					newtext = append(newtext, ctext[startat:]...)
				}
				break
			}

			newtext = append(newtext, ctext[startat:begin]...)

			end := librunes.FindSpace(ctext, begin)
			if end < 0 {
				break
			}

			startat = end
		}
		if len(newtext) > 0 {
			ctext = newtext
		}
	}
	return string(ctext)
}

// CleanWikiMarkup remove wiki markup from text.
//
//	List of known wiki markups,
//	- [[Category: ... ]]
//	- [[:Category: ... ]]
//	- [[File: ... ]]
//	- [[Help: ... ]]
//	- [[Image: ... ]]
//	- [[Special: ... ]]
//	- [[Wikipedia: ... ]]
//	- {{DEFAULTSORT: ... }}
//	- {{Template: ... }}
//	- <ref ... />
func CleanWikiMarkup(text string) string {
	ctext := []rune(text)

	for _, mu := range listWikiMarkup {
		ctext, _ = librunes.EncloseRemove(ctext, []rune(mu.begin), []rune(mu.end))
	}

	return string(ctext)
}

// MergeSpaces replace two or more horizontal spaces (' ', '\t', '\v', '\f',
// '\r') with single space.
// If withline is true it also replace two or more new lines with single
// new-line.
func MergeSpaces(text string, withline bool) string {
	var (
		isspace   bool
		isnewline bool
	)

	out := make([]rune, 0, len(text))

	for _, v := range text {
		switch v {
		case ' ', '\t', '\v', '\f', '\r':
			if isspace {
				continue
			}
			isspace = true
		default:
			isspace = false
		}
		if withline {
			if v == '\n' {
				if isnewline {
					continue
				}
				isnewline = true
			} else if isnewline {
				isnewline = false
			}
		}
		out = append(out, v)
	}
	return string(out)
}

// Reverse the string.
func Reverse(input string) string {
	r := []rune(input)
	x := 0
	y := len(r) - 1
	for x < len(r)/2 {
		r[x], r[y] = r[y], r[x]
		x++
		y--
	}
	return string(r)
}

// SingleSpace convert all sequences of white spaces into single space ' '.
func SingleSpace(in string) string {
	var isspace bool
	out := make([]rune, 0, len(in))
	for _, r := range in {
		if unicode.IsSpace(r) {
			if isspace {
				continue
			}
			isspace = true
			continue
		}
		if isspace {
			out = append(out, ' ')
			isspace = false
		}
		out = append(out, r)
	}
	if isspace {
		out = append(out, ' ')
	}
	return string(out)
}

// Split given a text, return all words in text.
//
// A word is any sequence of character which have length equal or greater than
// one and separated by white spaces.
//
// If cleanit is true remove any non-alphanumeric in the start and the end of
// each words.
//
// If uniq is true remove duplicate words, in case insensitive manner.
func Split(text string, cleanit bool, uniq bool) (words []string) {
	words = strings.Fields(text)

	if cleanit {
		// Remove non-alphanumeric character from start and end of each word.
		for x, word := range words {
			words[x] = TrimNonAlnum(word)
		}
	}

	if uniq {
		return Uniq(words, false)
	}

	return words
}

// TrimNonAlnum remove non alpha-numeric character at the beginning and
// end of `text`.
func TrimNonAlnum(text string) string {
	r := []rune(text)
	rlen := len(r)
	start := 0

	for ; start < rlen; start++ {
		if unicode.IsLetter(r[start]) || unicode.IsDigit(r[start]) {
			break
		}
	}

	if start >= rlen {
		return ""
	}

	r = r[start:]
	rlen = len(r)
	end := rlen - 1
	for ; end >= 0; end-- {
		if unicode.IsLetter(r[end]) || unicode.IsDigit(r[end]) {
			break
		}
	}

	if end < 0 {
		return ""
	}

	r = r[:end+1]

	return string(r)
}
