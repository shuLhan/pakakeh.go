// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

// Package paseto provide a simple, ready to use, opinionated implementation
// of Platform-Agnostic SEcurity TOkens (PASETO) version 2 and version 4 as
// defined in [paseto-v2] and [paseto-v4].
//
// The paseto version 2 is available under sub package
// [kilabit.info/pakakeh.go/lib/paseto/v2] and the paseto version 4
// is available under sub package
// [kilabit.info/pakakeh.go/lib/paseto/v4].
//
// [paseto-v2]: https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version2.md
// [paseto-v4]: https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version4.md
package paseto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// List of errors when unpacking token.
var (
	ErrToken        = errors.New(`invalid token`)
	ErrTokenFooter  = errors.New(`invalid token footer`)
	ErrTokenHeader  = errors.New(`invalid token header`)
	ErrTokenLen     = errors.New(`invalid token payload length`)
	ErrTokenPayload = errors.New(`invalid token payload`)
	ErrTokenSigLen  = errors.New(`invalid token signature length`)
)

// DefaultTTL define the time-to-live of a message, by setting ExpiredAt to
// current time + [DefaultTTL].
// This value can be changes to increase the message expiration time.
var DefaultTTL int64 = 60

// PreAuthEncode encodes each piece into single block.
func PreAuthEncode(pieces ...[]byte) (b []byte, err error) {
	logp := `PreAuthEncode`
	var buf bytes.Buffer

	err = binary.Write(&buf, binary.LittleEndian, uint64(len(pieces)))
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	for x := range len(pieces) {
		err = binary.Write(&buf, binary.LittleEndian, uint64(len(pieces[x])))
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		_, err = buf.Write(pieces[x])
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}
	return buf.Bytes(), nil
}
