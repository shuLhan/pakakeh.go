// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package big

import (
	"math/big"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewInt(t *testing.T) {
	cases := []struct {
		v   any
		exp string
	}{{
		v:   []byte("123.45"),
		exp: "123",
	}, {
		v:   "123.45",
		exp: "123",
	}, {
		v:   byte(255),
		exp: "255",
	}, {
		v:   int(-123),
		exp: "-123",
	}, {
		v:   int32(-123),
		exp: "-123",
	}, {
		v:   int64(-123),
		exp: "-123",
	}, {
		v:   uint64(12345),
		exp: "12345",
	}, {
		v:   float32(123.45),
		exp: "123",
	}, {
		v:   float64(123.45),
		exp: "123",
	}, {
		v:   NewInt("12345678901234567890"),
		exp: "12345678901234567890",
	}, {
		v:   big.NewInt(12345),
		exp: "12345",
	}, {
		v:   NewRat("1234567890.1234567890"),
		exp: "1234567890",
	}, {
		v:   big.NewRat(123456, 1),
		exp: "123456",
	}}

	for _, c := range cases {
		got := NewInt(c.v)
		test.Assert(t, "NewInt", c.exp, got.String())
	}
}

func TestInt_IsZero(t *testing.T) {
	cases := []struct {
		in  *Int
		exp bool
	}{{
		in:  NewInt(0),
		exp: true,
	}, {
		in: NewInt(1),
	}, {
		in: NewInt(-1),
	}}

	for _, c := range cases {
		test.Assert(t, "Int.IsZero", c.in.IsZero(), c.exp)
	}
}
