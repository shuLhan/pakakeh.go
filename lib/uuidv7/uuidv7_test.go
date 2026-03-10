// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package uuidv7

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestUUIDv7_Scan(t *testing.T) {
	listCase := []struct {
		desc     string
		data     any
		expError string
	}{{
		desc:     `With empty data`,
		data:     `017F22E2-79B0-7CC3-98C4-DC0C0C07398F`,
		expError: `uuidv7: Scan: invalid type string`,
	}, {
		desc:     `With invalid version`,
		data:     []byte(`5c146b14-3c52-4afd-938a-375d0df1fbf6`),
		expError: `uuidv7: invalid version 2`,
	}}
	for _, tc := range listCase {
		var id UUIDv7
		err := id.Scan(tc.data)
		if err != nil {
			test.Assert(t, tc.desc, tc.expError, err.Error())
			continue
		}
	}
}

func TestUUIDv7_UnmarshalBinary(t *testing.T) {
	listCase := []struct {
		desc     string
		expError string
		data     []byte
	}{{
		desc:     `With empty data`,
		expError: `uuidv7: invalid length 0`,
	}, {
		desc:     `With non-version 7`,
		data:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 1, 2, 3, 4},
		expError: `uuidv7: invalid version 0`,
	}}
	for _, tc := range listCase {
		var id UUIDv7
		err := id.UnmarshalBinary(tc.data)
		if err != nil {
			test.Assert(t, tc.desc, tc.expError, err.Error())
			continue
		}
	}
}

func TestUUIDv7_UnmarshalText(t *testing.T) {
	listCase := []struct {
		desc     string
		expError string
		data     []byte
	}{{
		desc:     `With empty data`,
		expError: `uuidv7: invalid length 0`,
	}, {
		desc:     `With non-version 7`,
		data:     []byte(`5c146b14-3c52-4afd-938a-375d0df1fbf6`),
		expError: `uuidv7: invalid version 2`,
	}, {
		desc:     `With invalid hex (high)`,
		data:     []byte(`X17F22E2-79B0-7CC3-98C4-DC0C0C07398F`),
		expError: `uuidv7: encoding/hex: invalid byte: U+0058 'X'`,
	}, {
		desc:     `With invalid hex (low)`,
		data:     []byte(`017F22E2-79B0-7CC3-98C4-DC0C0C07398X`),
		expError: `uuidv7: encoding/hex: invalid byte: U+0058 'X'`,
	}}
	for _, tc := range listCase {
		var id UUIDv7
		err := id.UnmarshalText(tc.data)
		if err != nil {
			test.Assert(t, tc.desc, tc.expError, err.Error())
			continue
		}
	}
}
