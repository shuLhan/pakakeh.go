// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestAddRat(t *testing.T) {
	cases := []struct {
		ins []interface{}
		exp *Rat
	}{{
		ins: nil,
	}, {
		ins: []interface{}{"a"},
	}, {
		ins: []interface{}{0, 0.0001},
		exp: NewRat(0.0001),
	}, {
		ins: []interface{}{"1.007", "a", "2.003"},
		exp: NewRat("3.01"),
	}}

	for _, c := range cases {
		got := AddRat(c.ins...)
		test.Assert(t, "AddRat", c.exp, got, true)
	}
}

func TestMulRat(t *testing.T) {
	cases := []struct {
		ins []interface{}
		exp *Rat
	}{{
		ins: nil,
	}, {
		ins: []interface{}{"a"},
	}, {
		ins: []interface{}{0, 1},
		exp: NewRat(0),
	}, {
		ins: []interface{}{
			NewRat(6),
			"a",
			NewRat("0.3"),
		},
		exp: NewRat("1.8"),
	}}

	for _, c := range cases {
		got := MulRat(c.ins...)
		test.Assert(t, "MulRat", c.exp, got, true)
	}
}

func TestQuoRat(t *testing.T) {
	cases := []struct {
		ins []interface{}
		exp *Rat
	}{{
		ins: nil,
	}, {
		ins: []interface{}{"a"},
	}, {
		ins: []interface{}{0, 1},
		exp: NewRat(0),
	}, {
		ins: []interface{}{
			NewRat(6),
			"a",
			NewRat("0.3"),
		},
		exp: NewRat("20"),
	}}

	for _, c := range cases {
		got := QuoRat(c.ins...)
		test.Assert(t, "QuoRat", c.exp, got, true)
	}
}

func TestSubRat(t *testing.T) {
	cases := []struct {
		ins []interface{}
		exp *Rat
	}{{
		ins: nil,
	}, {
		ins: []interface{}{"a"},
	}, {
		ins: []interface{}{0, 1},
		exp: NewRat(-1),
	}, {
		ins: []interface{}{
			NewRat(6),
			"a",
			NewRat("0.3"),
		},
		exp: NewRat("5.7"),
	}}

	for _, c := range cases {
		got := SubRat(c.ins...)
		test.Assert(t, "SubRat", c.exp, got, true)
	}
}

func TestNewRat(t *testing.T) {
	cases := []struct {
		v   interface{}
		exp *Rat
	}{{
		v: []byte("a"),
	}, {
		v:   []byte("14687233442.06916608"),
		exp: NewRat("14687233442.06916608"),
	}, {
		v:   "14687233442.06916608",
		exp: NewRat("14687233442.06916608"),
	}, {
		v:   NewRat("14687233442.06916608"),
		exp: NewRat("14687233442.06916608"),
	}, {
		v:   *(NewRat("14687233442.06916608")),
		exp: NewRat("14687233442.06916608"),
	}, {
		v:   big.NewRat(14687233442, 100_000_000),
		exp: NewRat("146.87233442"),
	}, {
		v:   *(big.NewRat(14687233442, 100_000_000)),
		exp: NewRat("146.87233442"),
	}}

	for _, c := range cases {
		got := NewRat(c.v)
		test.Assert(t, fmt.Sprintf("NewRat: %T(%v)", c.v, c.v),
			c.exp, got, true)
	}
}

func TestRat_Add(t *testing.T) {
	cases := []struct {
		got *Rat
		in  interface{}
		exp *Rat
	}{{
		got: NewRat(1),
		in:  nil,
		exp: NewRat(1),
	}, {
		got: NewRat(1),
		in:  1,
		exp: NewRat(2),
	}}

	for _, c := range cases {
		t.Logf("Add %T(%v)", c.in, c.in)

		c.got.Add(c.in)

		test.Assert(t, "Add", c.exp, c.got, true)
	}
}

func TestRat_Int64(t *testing.T) {
	cases := []struct {
		r   *Rat
		exp int64
	}{{
		r:   NewRat("0.000000001"),
		exp: 0,
	}, {
		r:   NewRat("0.5"),
		exp: 0,
	}, {
		r:   NewRat("0.6"),
		exp: 0,
	}, {
		r:   NewRat("4011144900.02438879").Mul(100000000),
		exp: 401114490002438879,
	}, {
		r:   QuoRat("128_900", "0.000_0322"),
		exp: 4003105590,
	}, {
		r:   QuoRat(128900, 3220).Mul(100000000),
		exp: 4003105590,
	}}

	for _, c := range cases {
		got := c.r.Int64()
		test.Assert(t, fmt.Sprintf("Int64 of %s", c.r), c.exp, got, true)
	}
}

func TestRat_IsEqual(t *testing.T) {
	f := NewRat(1)

	cases := []struct {
		g   interface{}
		exp bool
	}{{
		g: "a",
	}, {
		g: 1.1,
	}, {
		g:   byte(1),
		exp: true,
	}, {
		g:   int(1),
		exp: true,
	}, {
		g:   int32(1),
		exp: true,
	}, {
		g:   int64(1),
		exp: true,
	}, {
		g:   float32(1),
		exp: true,
	}, {
		g:   NewRat(1),
		exp: true,
	}}

	for _, c := range cases {
		got := f.IsEqual(c.g)
		test.Assert(t, "IsEqual", c.exp, got, true)
	}
}

func TestRat_IsGreater(t *testing.T) {
	r := NewRat("0.000_000_5")

	cases := []struct {
		in  interface{}
		exp bool
	}{{
		in: nil,
	}, {
		in: "0.000_000_5",
	}, {
		in:  "0.000_000_49",
		exp: true,
	}}
	for _, c := range cases {
		got := r.IsGreater(c.in)
		test.Assert(t, fmt.Sprintf("IsGreater %s", c.in),
			c.exp, got, true)
	}
}

func TestRat_IsGreaterOrEqual(t *testing.T) {
	r := NewRat("0.000_000_5")

	cases := []struct {
		in  interface{}
		exp bool
	}{{
		in: nil,
	}, {
		in: "0.000_000_500_000_000_001",
	}, {
		in:  "0.000_000_5",
		exp: true,
	}, {
		in:  "0.000_000_49",
		exp: true,
	}}
	for _, c := range cases {
		got := r.IsGreaterOrEqual(c.in)
		test.Assert(t, fmt.Sprintf("IsGreaterOrEqual %s", c.in),
			c.exp, got, true)
	}
}

func TestRat_IsGreaterThanZero(t *testing.T) {
	cases := []struct {
		in  interface{}
		exp bool
	}{{
		in: byte(0),
	}, {
		in: "-0.000_000_000_000_000_001",
	}, {
		in:  "0.000_000_000_000_000_001",
		exp: true,
	}, {
		in:  "0.000_000_5",
		exp: true,
	}}
	for _, c := range cases {
		got := NewRat(c.in).IsGreaterThanZero()
		test.Assert(t, fmt.Sprintf("IsGreaterThanZero %s", c.in),
			c.exp, got, true)
	}
}

func TestRat_IsLess(t *testing.T) {
	r := NewRat("0.000_000_5")

	cases := []struct {
		in  interface{}
		exp bool
	}{{
		in: nil,
	}, {
		in: "0.000_000_5",
	}, {
		in: "0.000_000_49",
	}, {
		in:  "0.000_000_500_000_000_001",
		exp: true,
	}}
	for _, c := range cases {
		got := r.IsLess(c.in)
		test.Assert(t, fmt.Sprintf("IsLess %s", c.in),
			c.exp, got, true)
	}
}

func TestRat_IsLessOrEqual(t *testing.T) {
	r := NewRat("0.000_000_5")

	cases := []struct {
		in  interface{}
		exp bool
	}{{
		in: nil,
	}, {
		in:  "0.000_000_5",
		exp: true,
	}, {
		in: "0.000_000_49",
	}, {
		in:  "0.000_000_500_000_000_001",
		exp: true,
	}}
	for _, c := range cases {
		got := r.IsLessOrEqual(c.in)
		test.Assert(t, fmt.Sprintf("IsLessOrEqual %s", c.in),
			c.exp, got, true)
	}
}

func TestRat_IsLessThanZero(t *testing.T) {
	cases := []struct {
		in  interface{}
		exp bool
	}{{
		in: byte(0),
	}, {
		in:  "-0.000_000_000_000_000_001",
		exp: true,
	}, {
		in: "0.000_000_000_000_000_001",
	}, {
		in: "0.000_000_5",
	}}
	for _, c := range cases {
		got := NewRat(c.in).IsLessThanZero()
		test.Assert(t, fmt.Sprintf("IsLessThanZero %s", c.in),
			c.exp, got, true)
	}
}

func TestRat_IsZero(t *testing.T) {
	cases := []struct {
		in  interface{}
		exp bool
	}{{
		in:  byte(0),
		exp: true,
	}, {
		in:  byte(-0),
		exp: true,
	}, {
		in: "-0.000_000_000_000_000_001",
	}, {
		in: "0.000_000_000_000_000_001",
	}}
	for _, c := range cases {
		got := NewRat(c.in).IsZero()
		test.Assert(t, fmt.Sprintf("IsZero %s", c.in),
			c.exp, got, true)
	}
}

func TestRat_MarshalJSON(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		exp: "0",
	}, {
		in:  "0.00000000",
		exp: "0",
	}, {
		in:  "0.1",
		exp: "0.1",
	}, {
		in:  "0.0000001",
		exp: "0.0000001",
	}, {
		in:  "0.00000001",
		exp: "0.00000001",
	}, {
		in:  "0.000000001",
		exp: "0",
	}, {
		in:  "1234567890.0",
		exp: "1234567890",
	}, {
		in:  "64.23738872403",
		exp: "64.23738872",
	}, {
		in:  "0.1234567890",
		exp: "0.12345679",
	}, {
		in:  "142660378.65368736",
		exp: "142660378.65368736",
	}, {
		in:  "9193394308.85771370",
		exp: "9193394308.8577137",
	}, {
		in:  "14687233442.06916608",
		exp: "14687233442.06916608",
	}}
	for _, c := range cases {
		var (
			got []byte
			err error
		)
		r := NewRat(c.in)
		got, err = r.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, fmt.Sprintf("MarshalJSON: %s", c.in),
			[]byte(c.exp), got, true)
	}
}

func TestRat_Mul(t *testing.T) {
	const (
		defValue = "14687233442.06916608"
	)

	cases := []struct {
		got *Rat
		in  interface{}
		exp *Rat
	}{{
		got: NewRat(defValue),
		in:  "a",
		exp: NewRat(0),
	}, {
		got: NewRat(defValue),
		in:  "0",
		exp: NewRat(0),
	}, {
		got: NewRat(defValue),
		in:  defValue,
		exp: NewRat("215714826181834884090.46087867"),
	}, {
		got: NewRat("1.06916608"),
		in:  "1.06916608",
		exp: NewRat("1.14311611"),
	}}

	for _, c := range cases {
		t.Logf("Mul %T(%v)", c.in, c.in)

		c.got.Mul(c.in)

		test.Assert(t, "Mul", c.exp, c.got, true)
	}
}

func TestRat_Quo(t *testing.T) {
	const (
		defValue = "14687233442.06916608"
	)

	cases := []struct {
		g   interface{}
		exp *Rat
	}{{
		g: "a",
	}, {
		g:   defValue,
		exp: NewRat(1),
	}, {
		g:   "100000000",
		exp: NewRat("146.87233442"),
	}}

	for _, c := range cases {
		r := NewRat(defValue)
		got := r.Quo(c.g)

		test.Assert(t, "Quo", c.exp, got, true)
	}
}

func TestRat_Scan(t *testing.T) {
	cases := []struct {
		in       interface{}
		exp      *Rat
		expError string
	}{{
		in:       nil,
		expError: "Rat.Scan: unknown type <nil>",
	}, {
		in:  "0.0001",
		exp: NewRat("0.0001"),
	}, {
		in:  float64(0.0001),
		exp: NewRat("0.0001"),
	}, {
		in:  (1.0 / 10000.0),
		exp: NewRat("0.0001"),
	}}

	for _, c := range cases {
		r := NewRat(0)
		err := r.Scan(c.in)
		if err != nil {
			test.Assert(t, "Scan error", c.expError, err.Error(), true)
			continue
		}
		test.Assert(t, fmt.Sprintf("Scan(%T(%v))", c.in, c.in),
			c.exp, r, true)
	}
}

func TestRat_String_fromString(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in:  "12345",
		exp: "12345",
	}, {
		in:  "0.00000000",
		exp: "0",
	}, {
		in:  "0.1",
		exp: "0.1",
	}, {
		in:  "0.0000001",
		exp: "0.0000001",
	}, {
		in:  "0.00000001",
		exp: "0.00000001",
	}, {
		in:  "0.000000001",
		exp: "0",
	}, {
		in:  "1234567890.0",
		exp: "1234567890",
	}, {
		in:  "64.23738872403",
		exp: "64.23738872",
	}, {
		in:  "0.1234567890",
		exp: "0.12345679",
	}, {
		in:  "142660378.65368736",
		exp: "142660378.65368736",
	}, {
		in:  "9193394308.85771370",
		exp: "9193394308.8577137",
	}, {
		in:  "14687233442.06916608",
		exp: "14687233442.06916608",
	}}

	for _, c := range cases {
		got := MustRat(c.in)
		test.Assert(t, c.in, c.exp, got.String(), true)
	}
}

func TestRat_String_fromFloat64(t *testing.T) {
	cases := []struct {
		in  float64
		exp string
	}{{
		in:  0.00000000,
		exp: "0",
	}, {
		in:  0.1,
		exp: "0.1",
	}, {
		in:  0.000_000_1,
		exp: "0.0000001",
	}, {
		in:  0.000_000_01,
		exp: "0.00000001",
	}, {
		in:  0.000000001,
		exp: "0",
	}, {
		in:  1234567890.0,
		exp: "1234567890",
	}, {
		in:  64.23738872403,
		exp: "64.23738872",
	}, {
		in:  0.1234567890,
		exp: "0.12345679",
	}, {
		in:  142660378.65368736,
		exp: "142660378.65368736",
	}, {
		in:  9193394308.85771370,
		exp: "9193394308.8577137",
	}}

	for _, c := range cases {
		got := NewRat(c.in)
		test.Assert(t, c.exp, c.exp, got.String(), true)
	}
}

func TestRat_Sub(t *testing.T) {
	cases := []struct {
		got *Rat
		in  interface{}
		exp *Rat
	}{{
		got: NewRat(1),
		in:  nil,
		exp: NewRat(1),
	}, {
		got: NewRat(1),
		in:  1,
		exp: NewRat(0),
	}}

	for _, c := range cases {
		t.Logf("Sub %T(%v)", c.in, c.in)

		c.got.Sub(c.in)

		test.Assert(t, "Sub", c.exp, c.got, true)
	}
}

func TestRat_UnmarshalJSON(t *testing.T) {
	type T struct {
		V *Rat `json:"V"`
	}

	cases := []struct {
		in       []byte
		exp      *Rat
		expError string
	}{{
		in:       []byte(`{"V":"ab"}`),
		expError: "Rat.UnmarshalJSON: cannot convert []uint8([97 98]) to Rat",
	}, {
		in: []byte(`{}`),
	}, {
		in:       []byte(`{"V":}`),
		expError: `invalid character '}' looking for beginning of value`,
	}, {
		in:  []byte(`{"V":0}`),
		exp: NewRat(0),
	}, {
		in:  []byte(`{"V":"1"}`),
		exp: NewRat(1),
	}, {
		in:  []byte(`{"V":0.00000001}`),
		exp: MustRat("0.00000001"),
	}}

	for _, c := range cases {
		t.Logf("%q", c.in)

		got := &T{}
		err := json.Unmarshal(c.in, &got)
		if err != nil {
			test.Assert(t, fmt.Sprintf("error %s", c.in),
				c.expError, err.Error(), true)
			continue
		}

		test.Assert(t, "", c.exp, got.V, true)
	}
}
