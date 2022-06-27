// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package html extends the golang.org/x/net/html by providing simplified
// methods to Node.
//
// The x/net/html package currently only provide bare raw functionalities
// to iterate tree, there is no check for empty node, and no function to
// get attribute by name without looping it manually.
//
// This package extends the parent package by adding methods to get node's
// attribute by name, get the first non-empty child, get the next
// non-empty sibling, and method to iterate the tree.
package html

import (
	"bytes"

	"golang.org/x/net/html"

	"github.com/shuLhan/share/lib/ascii"
)

// NormalizeForID given an input string normalize it to HTML ID.
// The normalization follow [Mozilla specification] rules,
//
//   - it must not contain whitespace (spaces, tabs etc.),
//   - only ASCII letters, digits, '_', and '-' should be used, and
//   - it should start with a letter.
//
// This function,
//
//   - Return "_" if input is empty string,
//   - replace unknown characters with '_',
//   - prefix output with '_' unless it start with '-', '_', or letters, and
//   - convert letters to lower cases.
//
// [Mozilla specification]: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/id.
func NormalizeForID(in string) (out string) {
	var (
		bin = []byte(in)
		x   int
		b   byte
	)

	for x, b = range bin {
		if ascii.IsAlpha(b) {
			if b >= 'A' && b <= 'Z' {
				bin[x] = b + 32
			}
		} else if !(ascii.IsDigit(b) || b == '-' || b == '_') {
			bin[x] = '_'
		}
	}
	if len(bin) == 0 {
		bin = append(bin, '_')
	} else if !ascii.IsAlpha(bin[0]) && bin[0] != '_' {
		bin = append(bin, '_')
		copy(bin[1:], bin[:])
		bin[0] = '_'
	}

	return string(bin)
}

// Sanitize the content of HTML into plain text.
func Sanitize(in []byte) (plain []byte) {
	if len(in) == 0 {
		return plain
	}

	var (
		r         = bytes.NewReader(in)
		twoSpaces = []byte("  ")

		w         bytes.Buffer
		htmlToken *html.Tokenizer
		tokenType html.TokenType
		tagName   []byte
		x         int
	)

	htmlToken = html.NewTokenizer(r)
	for {
		tokenType = htmlToken.Next()
		switch tokenType {
		case html.ErrorToken:
			goto out

		case html.TextToken:
			w.Write(htmlToken.Text())

		case html.StartTagToken:
			tagName, _ = htmlToken.TagName()

			if bytes.Equal(tagName, []byte("title")) ||
				bytes.Equal(tagName, []byte("script")) {
				htmlToken.Next()
			}
		}
	}
out:
	plain = w.Bytes()
	plain = bytes.Replace(plain, []byte("\r"), nil, -1)
	plain = bytes.Replace(plain, []byte("\n"), []byte(" "), -1)
	plain = bytes.Replace(plain, []byte("\t"), []byte(" "), -1)
	for {
		x = bytes.Index(plain, twoSpaces)
		if x < 0 {
			break
		}
		plain = bytes.Replace(plain, twoSpaces, []byte(" "), -1)
	}

	return plain
}
