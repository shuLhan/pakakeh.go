// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package email

import (
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

// DateFormat define the default date layout in email.
const DateFormat = "Mon, 2 Jan 2006 15:04:05 -0700"

const (
	contentTypeMultipartAlternative = "multipart/alternative"
	contentTypeTextPlain            = `text/plain; charset="utf-8"`
	contentTypeTextHTML             = `text/html; charset="utf-8"`
	mimeVersion1                    = "1.0"
)

// List of content type encoding.
const (
	encodingQuotedPrintable = `quoted-printable`
	encodingBase64          = `base64`
)

const (
	cr byte = '\r'
	lf byte = '\n'
)

var boundSeps = []byte{'-', '-'}

// dateInUtc if set to true, the Date header will be set to UTC instead of
// local time.
// This variable is used to make testing works on all zones.
var dateInUtc bool

// Epoch return the UNIX timestamp in seconds.
//
// This variable is exported to allow function that use time can
// be tested with fixed, predictable value.
var Epoch = func() int64 {
	return time.Now().Unix()
}

// randomChars generate n random characters.
func randomChars(n int) []byte {
	return ascii.Random([]byte(ascii.LettersNumber), n)
}

// randomString generate random string with n characters.
func randomString(n int) string {
	var v = ascii.Random([]byte(ascii.LettersNumber), n)
	return string(v)
}

// sanitize remove comment from in and merge multiple spaces into one.
// A comment start with '(' and end with ')' and can be nested
// "(...(...(...)...)".
func sanitize(in []byte) (out []byte) {
	var (
		c         byte
		inComment int
		hasSpace  bool
	)
	out = make([]byte, 0, len(in))
	for _, c = range in {
		if inComment != 0 {
			if c == ')' {
				inComment--
			} else if c == '(' {
				inComment++
			}
			continue
		}
		if c == '(' {
			inComment++
			continue
		}
		if ascii.IsSpace(c) {
			hasSpace = true
			continue
		}
		if hasSpace {
			out = append(out, ' ')
			hasSpace = false
		}
		out = append(out, c)
	}
	return out
}
