// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

const (
	contentTypeMultipartAlternative = "multipart/alternative"
	contentTypeTextPlain            = `text/plain; charset="utf-8"`
	contentTypeTextHTML             = `text/html; charset="utf-8"`
	encodingQuotedPrintable         = "quoted-printable"
	mimeVersion1                    = "1.0"
)

const (
	cr byte = '\r'
	lf byte = '\n'
)

//nolint:gochecknoglobals
var boundSeps = []byte{'-', '-'}
