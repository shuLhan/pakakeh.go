// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package big

import (
	"bytes"
	"log"
	"math/big"
)

// DefaultBitPrecision define the maximum number of mantissa bits available to
// represent the value.
//
// In standard library this value is 24 for float32 or 53 for float64.
//
// One should change this value before using the new extended Float in the
// program.
var DefaultBitPrecision uint = 128

// DefaultRoundingMode define the default rounding mode for all instance of
// Float.
//
// One should change this value before using the new extended Float in the
// program.
var DefaultRoundingMode = big.ToNearestAway

// Float extend the standard big.Float by setting each instance precision to
// DefaultBitPrecision, rounding mode to DefaultRoundingMode, and using
// DefaultDigitPrecision value after decimal point when converted to string.
type Float struct {
	big.Float
}

// AddFloat return the rounded sum `f[0]+f[1]+...`.
// It will return nil if the first parameter is not convertable to Float.
func AddFloat(f ...any) *Float {
	if len(f) == 0 {
		return nil
	}
	total := toFloat(f[0])
	if total == nil {
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toFloat(f[x])
		if rx == nil {
			continue
		}
		total.Add(rx)
	}
	return total
}

// NewFloat create and initialize new Float with default bit precision,
// and rounding mode.
func NewFloat(v any) *Float {
	return toFloat(v)
}

// CreateFloat create Float with default bit precision and rounding mode.
func CreateFloat(v float64) Float {
	f := Float{}
	f.SetPrec(DefaultBitPrecision)
	f.SetMode(DefaultRoundingMode)
	f.SetFloat64(v)
	return f
}

// MulFloat return the result of multiplication `f*g`.
// It will return nil if `f` or `g` is not convertible to Float.
func MulFloat(f, g any) *Float {
	ff := toFloat(f)
	if ff == nil {
		return nil
	}
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	h := ff.Clone()
	return h.Mul(gf)
}

// MustParseFloat convert the string `s` into Float or panic.
func MustParseFloat(s string) (f *Float) {
	f = NewFloat(0)
	_, _, err := f.Float.Parse(s, 10)
	if err != nil {
		log.Fatal("MustParseFloat:", err)
	}
	return f
}

// ParseFloat the string s into Float value.
func ParseFloat(s string) (f *Float, err error) {
	f = NewFloat(0)
	_, _, err = f.Float.Parse(s, 10)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// QuoFloat return the quotient of `f/g` as new Float.
func QuoFloat(f, g any) *Float {
	ff := toFloat(f)
	if ff == nil {
		return nil
	}
	gf := toFloat(g)
	if gf == nil {
		return nil
	}

	h := ff.Clone()
	return h.Quo(gf)
}

// SubFloat return the result of subtraction `f-g` as new Float.
func SubFloat(f, g *Float) *Float {
	h := f.Clone()
	return h.Sub(g)
}

// Add sets f to `f+g` and return the f as the result.
func (f *Float) Add(g any) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Add(&f.Float, &gf.Float)
	return f
}

// Clone the instance to new Float.
func (f *Float) Clone() *Float {
	g := NewFloat(0)
	g.Float.Set(&f.Float)
	return g
}

// Int64 return the integer resulting from truncating x towards zero.
func (f *Float) Int64() int64 {
	i64, _ := f.Float.Int64()
	return i64
}

// IsEqual will return true if `f == g`.
func (f *Float) IsEqual(g any) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) == 0
}

// IsGreater will return true if `f > g`.
func (f *Float) IsGreater(g any) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) > 0
}

// IsGreaterOrEqual will return true if `f >= g`.
func (f *Float) IsGreaterOrEqual(g any) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) >= 0
}

// IsLess will return true if `f < g`.
func (f *Float) IsLess(g any) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) < 0
}

// IsLessOrEqual will return true if `f <= g`.
func (f *Float) IsLessOrEqual(g any) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) <= 0
}

// IsZero will return true if `f == 0`.
func (f *Float) IsZero() bool {
	gf := NewFloat(0)
	return f.Cmp(&gf.Float) == 0
}

// MarshalJSON implement the json.Marshaler interface and return the output of
// String method.
func (f *Float) MarshalJSON() ([]byte, error) {
	s := f.String()
	if MarshalJSONAsString {
		s = `"` + s + `"`
	}
	return []byte(s), nil
}

// Mul sets f to product of `f * g` and return the result as f.
// If g is not convertible to Float it will return nil.
func (f *Float) Mul(g any) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Mul(&f.Float, &gf.Float)
	return f
}

// ParseFloat parse the string into Float value.
func (f *Float) ParseFloat(s string) (err error) {
	f.SetPrec(DefaultBitPrecision)
	f.SetMode(DefaultRoundingMode)
	_, _, err = f.Float.Parse(s, 10)
	if err != nil {
		return err
	}
	return nil
}

// Quo sets f to quotient of `f/g` and return the result as f.
// If g is not convertible to Float it will return nil.
func (f *Float) Quo(g any) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Quo(&f.Float, &gf.Float)
	return f
}

// String format the Float value into string with maximum mantissa is set by
// digit precision option.
//
// Unlike standard String method, this method will trim trailing zero digit or
// decimal point at the end of mantissa.
func (f *Float) String() string {
	b := []byte(f.Text('f', DefaultDigitPrecision))

	pointIdx := bytes.Index(b, []byte{'.'})
	if pointIdx < 0 {
		return string(b)
	}

	b = bytes.TrimRight(b, "0")
	b = bytes.TrimRight(b, ".")

	return string(b)
}

// Sub sets f to rounded difference `f-g` and return f.
func (f *Float) Sub(g any) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Sub(&f.Float, &gf.Float)
	return f
}

// UnmarshalJSON convert the JSON byte value into Float.
func (f *Float) UnmarshalJSON(in []byte) (err error) {
	if f == nil {
		f = NewFloat(0)
	}
	err = f.ParseFloat(string(in))
	return err
}

// toFloat convert v type to Float or nil if v type is unknown.
func toFloat(g any) (out *Float) {
	out = &Float{}
	out.SetPrec(DefaultBitPrecision).SetMode(DefaultRoundingMode)

	switch v := g.(type) {
	case []byte:
		if len(v) == 0 {
			out.SetInt64(0)
		} else {
			_, ok := out.Float.SetString(string(v))
			if !ok {
				return nil
			}
		}

	case string:
		if len(v) == 0 {
			out.SetInt64(0)
		} else {
			_, ok := out.Float.SetString(v)
			if !ok {
				return nil
			}
		}

	case byte:
		out.SetInt64(int64(v))
	case int:
		out.SetInt64(int64(v))
	case int32:
		out.SetInt64(int64(v))
	case int64:
		out.SetInt64(v)
	case uint:
		out.SetUint64(uint64(v))
	case uint16:
		out.SetUint64(uint64(v))
	case uint32:
		out.SetUint64(uint64(v))
	case uint64:
		out.SetUint64(v)
	case float32:
		out.SetFloat64(float64(v))
	case float64:
		out.SetFloat64(v)
	case Float:
		out.Float.Set(&v.Float)
	case *Float:
		if v == nil {
			out.SetInt64(0)
		} else {
			out.Float.Set(&v.Float)
		}
	case big.Int:
		out.Float.SetInt(&v)
	case *big.Int:
		if v == nil {
			out.Float.SetInt64(0)
		} else {
			out.Float.SetInt(v)
		}
	case big.Rat:
		out.Float.SetRat(&v)
	case *big.Rat:
		if v == nil {
			out.Float.SetInt64(0)
		} else {
			out.Float.SetRat(v)
		}
	default:
		return nil
	}
	return out
}
