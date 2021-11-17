// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"encoding/json"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestFloat_Clone(t *testing.T) {
	const (
		defValue = "14687233442.06916608"
	)

	f, err := ParseFloat(defValue)
	if err != nil {
		t.Fatal(err)
	}

	got := f.Clone()

	test.Assert(t, "Clone", f.String(), got.String())
}

func TestFloat_IsEqual(t *testing.T) {
	f := NewFloat(1)

	cases := []struct {
		g   interface{}
		exp bool
	}{{
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
		g:   NewFloat(1),
		exp: true,
	}, {
		g:   CreateFloat(1),
		exp: true,
	}}

	for _, c := range cases {
		got := f.IsEqual(c.g)
		test.Assert(t, "IsEqual", c.exp, got)
	}
}

func TestFloat_Mul(t *testing.T) {
	const (
		defValue = "14687233442.06916608"
	)

	cases := []struct {
		g   string
		exp string
	}{{
		g:   "0",
		exp: "0",
	}, {
		g:   defValue,
		exp: "215714826181834884090.46087867",
	}}

	for _, c := range cases {
		f, err := ParseFloat(defValue)
		if err != nil {
			t.Fatal(err)
		}

		g, err := ParseFloat(c.g)
		if err != nil {
			t.Fatal(err)
		}

		f.Mul(g)
		got := f.String()

		test.Assert(t, "Mul", c.exp, got)
	}
}

func TestFloat_MulFloat64(t *testing.T) {
	const (
		defValue = "1.06916608"
	)

	cases := []struct {
		exp string
		g   float64
	}{{
		g:   0,
		exp: "0",
	}, {
		g:   1.06916608,
		exp: "1.14311611",
	}}

	for _, c := range cases {
		f, err := ParseFloat(defValue)
		if err != nil {
			t.Fatal(err)
		}

		f.Mul(c.g)
		got := f.String()

		test.Assert(t, "MulFloat64", c.exp, got)
	}
}

func TestFloat_Quo(t *testing.T) {
	const (
		defValue = "14687233442.06916608"
	)

	cases := []struct {
		in  string
		exp string
	}{{
		in:  "0",
		exp: "+Inf",
	}, {
		in:  defValue,
		exp: "1",
	}, {
		in:  "100000000",
		exp: "146.87233442",
	}}

	for _, c := range cases {
		f, err := ParseFloat(defValue)
		if err != nil {
			t.Fatal(err)
		}

		g, err := ParseFloat(c.in)
		if err != nil {
			t.Fatal(err)
		}

		f.Quo(g)

		got := f.String()

		test.Assert(t, "Quo", c.exp, got)
	}
}

func TestFloat_QuoFloat64(t *testing.T) {
	const (
		defValue = "14687233442.06916608"
	)

	cases := []struct {
		in  string
		exp string
	}{{
		in:  "0",
		exp: "+Inf",
	}, {
		in:  defValue,
		exp: "1",
	}, {
		in:  "100000000",
		exp: "146.87233442",
	}}

	for _, c := range cases {
		f, err := ParseFloat(defValue)
		if err != nil {
			t.Fatal(err)
		}

		g, err := ParseFloat(c.in)
		if err != nil {
			t.Fatal(err)
		}

		f.Quo(g)

		got := f.String()

		test.Assert(t, "Quo", c.exp, got)
	}
}

func TestFloat_String_fromString(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
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

	bf := CreateFloat(0)

	for _, c := range cases {
		err := bf.ParseFloat(c.in)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.in, c.exp, bf.String())
	}
}

func TestFloat_String_fromFloat(t *testing.T) {
	cases := []struct {
		exp string
		in  float64
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

	bf := CreateFloat(0)

	for _, c := range cases {
		bf.SetFloat64(c.in)
		test.Assert(t, c.exp, c.exp, bf.String())
	}
}

func TestFloat_UnmarshalJSON(t *testing.T) {
	type T struct {
		V *Float `json:"F"`
	}

	cases := []struct {
		exp *Float
		in  []byte
	}{{
		in: []byte(`{}`),
	}, {
		in:  []byte(`{"F":0}`),
		exp: NewFloat(0),
	}, {
		in:  []byte(`{"F":0.00000001}`),
		exp: MustParseFloat("0.00000001"),
	}}

	for _, c := range cases {
		t.Logf("%q", c.in)

		got := &T{}
		err := json.Unmarshal(c.in, &got)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "", c.exp, got.V)
	}
}
