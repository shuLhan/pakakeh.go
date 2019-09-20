// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package sanitize provide a function to sanitize markup document into plain
// text.
//
package sanitize

import (
	"bytes"

	"golang.org/x/net/html"
)

//
// HTML sanitize the content of HTML into plain text.
//
func HTML(in []byte) (plain []byte) {
	if len(in) == 0 {
		return plain
	}
	var (
		r         = bytes.NewReader(in)
		w         bytes.Buffer
		twoSpaces = []byte("  ")
	)

	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			goto out

		case html.TextToken:
			w.Write(z.Text())

		case html.StartTagToken:
			btag, _ := z.TagName()

			switch string(btag) {
			case "body", "head", "meta", "link":
			case "title", "script":
				z.Next()
			}
		}
	}
out:
	plain = w.Bytes()
	plain = bytes.Replace(plain, []byte("\r"), nil, -1)
	plain = bytes.Replace(plain, []byte("\n"), []byte(" "), -1)
	plain = bytes.Replace(plain, []byte("\t"), []byte(" "), -1)
	for {
		s := bytes.Index(plain, twoSpaces)
		if s < 0 {
			break
		}
		plain = bytes.Replace(plain, twoSpaces, []byte(" "), -1)
	}

	return plain
}
