// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package uuidv7

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestUUID_Equal(t *testing.T) {
	id := Parse(`019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF`)

	listCase := []struct {
		desc     string
		v        any
		expError string
	}{{
		desc:     `EmptyInterface`,
		expError: `uuidv7: expecting type *uuidv7.UUID, got <nil>`,
	}, {
		desc:     `NotEqualTime`,
		v:        Parse(`019CD2F8-1AE4-774E-BFFF-FFFFFFFFFFFF`),
		expError: `uuidv7: not equal, want 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF, got 019CD2F8-1AE4-774E-BFFF-FFFFFFFFFFFF`,
	}, {
		desc:     `NotEqualRand`,
		v:        Parse(`019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFF0`),
		expError: `uuidv7: not equal, want 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF, got 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFF0`,
	}, {
		desc: `AsPointerUUID`,
		v:    new(Parse(`019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF`)),
	}, {
		desc: `AsUUID`,
		v:    Parse(`019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF`),
	}}
	var gotError string
	for _, tc := range listCase {
		gotError = ``
		err := id.Equal(tc.v)
		if err != nil {
			gotError = err.Error()
		}
		test.Assert(t, tc.desc, tc.expError, gotError)
	}
}

func TestUUID_Scan(t *testing.T) {
	id := Parse(`019B76DA-A800-7000-8000-00000000001A`)

	listCase := []struct {
		desc     string
		data     any
		expError string
		exp      string
	}{{
		desc:     `With empty data`,
		data:     `017F22E2-79B0-7CC3-98C4-DC0C0C07398F`,
		expError: `uuidv7: Scan: invalid type string`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}, {
		desc:     `With invalid version`,
		data:     []byte(`5c146b14-3c52-4afd-938a-375d0df1fbf6`),
		expError: `uuidv7: invalid version 2`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}, {
		desc:     `With nil`,
		data:     nil,
		expError: ``,
		exp:      `00000000-0000-0000-0000-000000000000`,
	}}
	for _, tc := range listCase {
		err := id.Scan(tc.data)
		if err != nil {
			test.Assert(t, tc.desc, tc.expError, err.Error())
		}
		test.Assert(t, tc.desc, tc.exp, id.String())
	}
}

func TestUUID_UnmarshalBinary(t *testing.T) {
	id := Parse(`019B76DA-A800-7000-8000-00000000001A`)

	listCase := []struct {
		desc     string
		expError string
		exp      string
		data     []byte
	}{{
		desc:     `With empty data`,
		expError: `uuidv7: invalid length 0`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}, {
		desc:     `With non-version 7`,
		data:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 1, 2, 3, 4},
		expError: `uuidv7: invalid version 0`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}, {
		desc: `OK`,
		data: []byte{0xf, 0xf, 0xf, 0xf,
			0xf, 0xf,
			0x70, 0xf,
			0xf, 0xf,
			0xf, 0xf, 0xf, 0xf, 0xf, 0xf},
		exp: `0F0F0F0F-0F0F-700F-0F0F-0F0F0F0F0F0F`,
	}}
	for _, tc := range listCase {
		err := id.UnmarshalBinary(tc.data)
		if err != nil {
			test.Assert(t, tc.desc, tc.expError, err.Error())
		}
		test.Assert(t, tc.desc, tc.exp, id.String())
	}
}

func TestUUID_UnmarshalText(t *testing.T) {
	id := Parse(`019B76DA-A800-7000-8000-00000000001A`)

	listCase := []struct {
		desc     string
		expError string
		exp      string
		data     []byte
	}{{
		desc:     `With empty data`,
		expError: `uuidv7: invalid length 0`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}, {
		desc:     `With non-version 7`,
		data:     []byte(`5c146b14-3c52-4afd-938a-375d0df1fbf6`),
		expError: `uuidv7: invalid version 2`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}, {
		desc:     `With invalid hex (high)`,
		data:     []byte(`X17F22E2-79B0-7CC3-98C4-DC0C0C07398F`),
		expError: `uuidv7: encoding/hex: invalid byte: U+0058 'X'`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}, {
		desc:     `With invalid hex (low)`,
		data:     []byte(`017F22E2-79B0-7CC3-98C4-DC0C0C07398X`),
		expError: `uuidv7: encoding/hex: invalid byte: U+0058 'X'`,
		exp:      `019B76DA-A800-7000-8000-00000000001A`,
	}}
	for _, tc := range listCase {
		err := id.UnmarshalText(tc.data)
		if err != nil {
			test.Assert(t, tc.desc, tc.expError, err.Error())
		}
		test.Assert(t, tc.desc, tc.exp, id.String())
	}
}

func TestUUID_Value(t *testing.T) {
	var id UUID
	v, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `ValueZero`, v.(string), `00000000-0000-0000-0000-000000000000`)

	err = id.Scan([]byte(`019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF`))
	if err != nil {
		t.Fatal(err)
	}
	v, err = id.Value()
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, `ValueZero`, v.(string), `019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF`)
}
