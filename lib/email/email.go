// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

var ( // nolint: gochecknoglobals
	crlf      = []byte{'\r', '\n'}
	boundSeps = []byte{'-', '-'}
)

//
// Email represent an internet message.
//
type Email struct {
	Header Header
	Body   Body
}

//
// Unpack the raw message header and body.
//
func (email *Email) Unpack(raw []byte) ([]byte, error) {
	var err error

	raw, err = email.Header.Unpack(raw)
	if err != nil {
		return raw, err
	}

	boundary := email.Header.Boundary()

	raw, err = email.Body.Unpack(raw, boundary)

	return raw, err
}
