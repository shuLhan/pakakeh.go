// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"math/rand"
	"time"

	"github.com/shuLhan/share/lib/ascii"
)

const (
	contentTypeMultipartAlternative = "multipart/alternative"
	contentTypeTextPlain            = `text/plain; charset="utf-8"`
	contentTypeTextHTML             = `text/html; charset="utf-8"`
	DateFormat                      = "Mon, 2 Jan 2006 15:04:05 -0700"
	encodingQuotedPrintable         = "quoted-printable"
	mimeVersion1                    = "1.0"
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

//
// Epoch return the UNIX timestamp in seconds.
//
// This variable is exported to allow function that use time and math/rand can
// be tested with fixed, predictable value.
//
var Epoch = func() int64 {
	return time.Now().Unix()
}

// randomChars generate n random characters.
func randomChars(n int) []byte {
	rand.Seed(Epoch())
	return ascii.Random([]byte(ascii.LettersNumber), n)
}
