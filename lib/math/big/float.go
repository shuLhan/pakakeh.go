// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"bytes"
	"log"
	"math/big"
)

//
// DefaultBitPrecision define the maximum number of mantissa bits available to
// represent the value.
//
// In standard library this value is 24 for float32 or 53 for float64.
//
// One should change this value before using the new extended Float in the
// program.
//
//nolint: gochecknoglobals
var DefaultBitPrecision uint = 128

//
// DefaultRoundingMode define the default rounding mode for all instance of
// Float.
//
// One should change this value before using the new extended Float in the
// program.
//
//nolint: gochecknoglobals
var DefaultRoundingMode = big.ToNearestAway

//
// Float extend the standard big.Float by setting each instance precision to
// DefaultBitPrecision, rounding mode to DefaultRoundingMode, and using
// DefaultDigitPrecision value after decimal point when converted to string.
//
type Float struct {
	big.Float
}

//
// AddFloat return the rounded sum `f+g` and return f.
//
func AddFloat(f, g interface{}) *Float {
	ff := toFloat(f)
	gf := toFloat(g)
	if ff == nil {
		return nil
	}
	if gf == nil {
		return nil
	}
	h := gf.Clone()
	return h.Add(gf)
}

//
// NewFloat create and initialize new Float with default bit precision,
// and rounding mode.
//
func NewFloat(v float64) *Float {
	f := &Float{}
	f.SetPrec(DefaultBitPrecision)
	f.SetMode(DefaultRoundingMode)
	f.SetFloat64(v)
	return f
}

//
// Create Float with default bit precision and rounding mode.
//
func CreateFloat(v float64) Float {
	f := Float{}
	f.SetPrec(DefaultBitPrecision)
	f.SetMode(DefaultRoundingMode)
	f.SetFloat64(v)
	return f
}

//
// MulFloat return the result of multiplication `f*g`.
// It will return nil if `f` or `g` is not convertible to Float.
//
func MulFloat(f, g interface{}) *Float {
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

//
// MustParseFloat convert the string `s` into Float or panic.
//
func MustParseFloat(s string) (f *Float) {
	f = NewFloat(0)
	_, _, err := f.Float.Parse(s, 10)
	if err != nil {
		log.Fatal("MustParseFloat:", err)
	}
	return f
}

//
// ParseFloat the string s into Float value.
//
func ParseFloat(s string) (f *Float, err error) {
	f = NewFloat(0)
	_, _, err = f.Float.Parse(s, 10)
	if err != nil {
		return nil, err
	}
	return f, nil
}

//
// QuoFloat return the quotient of `f/g` as new Float.
//
func QuoFloat(f, g interface{}) *Float {
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

//
// SubFloat return the result of subtraction `f-g` as new Float.
//
func SubFloat(f, g *Float) *Float {
	h := f.Clone()
	return h.Sub(g)
}

//
// Add sets f to `f+g` and return the f as the result.
//
func (f *Float) Add(g interface{}) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Add(&f.Float, &gf.Float)
	return f
}

//
// Clone the instance to new Float.
//
func (f *Float) Clone() *Float {
	g := NewFloat(0)
	g.Float.Set(&f.Float)
	return g
}

//
// Int64 return the integer resulting from truncating x towards zero.
//
func (f *Float) Int64() int64 {
	i64, _ := f.Float.Int64()
	return i64
}

//
// IsEqual will return true if `f == g`.
//
func (f *Float) IsEqual(g interface{}) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) == 0
}

//
// IsGreater will return true if `f > g`.
//
func (f *Float) IsGreater(g interface{}) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) > 0
}

//
// IsGreaterOrEqual will return true if `f >= g`.
//
func (f *Float) IsGreaterOrEqual(g interface{}) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) >= 0
}

//
// IsLess will return true if `f < g`.
//
func (f *Float) IsLess(g interface{}) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) < 0
}

//
// IsLessOrEqual will return true if `f <= g`.
//
func (f *Float) IsLessOrEqual(g interface{}) bool {
	gf := toFloat(g)
	if gf == nil {
		return false
	}
	return f.Cmp(&gf.Float) <= 0
}

//
// IsZero will return true if `f == 0`.
//
func (f *Float) IsZero() bool {
	gf := NewFloat(0)
	return f.Cmp(&gf.Float) == 0
}

//
// MarshalJSON implement the json.Marshaler interface and return the output of
// String method.
//
func (f *Float) MarshalJSON() ([]byte, error) {
	s := f.String()
	return []byte(s), nil
}

//
// Mul sets f to product of `f * g` and return the result as f.
// If g is not convertible to Float it will return nil.
//
func (f *Float) Mul(g interface{}) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Mul(&f.Float, &gf.Float)
	return f
}

//
// Parse the string into Float value.
//
func (f *Float) ParseFloat(s string) (err error) {
	f.SetPrec(DefaultBitPrecision)
	f.SetMode(DefaultRoundingMode)
	_, _, err = f.Float.Parse(s, 10)
	if err != nil {
		return err
	}
	return nil
}

//
// Quo sets f to quotient of `f/g` and return the result as f.
// If g is not convertible to Float it will return nil.
//
func (f *Float) Quo(g interface{}) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Quo(&f.Float, &gf.Float)
	return f
}

//
// String format the Float value into string with maximum mantissa is set by
// digit precision option.
//
// Unlike standard String method, this method will trim trailing zero digit or
// decimal point at the end of mantissa.
//
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

//
// Sub sets f to rounded difference `f-g` and return f.
//
func (f *Float) Sub(g interface{}) *Float {
	gf := toFloat(g)
	if gf == nil {
		return nil
	}
	f.Float.Sub(&f.Float, &gf.Float)
	return f
}

//
// UnmarshalJSON convert the JSON byte value into Float.
//
func (f *Float) UnmarshalJSON(in []byte) (err error) {
	if f == nil {
		f = NewFloat(0)
	}
	err = f.ParseFloat(string(in))
	return err
}

//
// toFloat convert any type to Float or nil if type is unknown.
//
func toFloat(v interface{}) *Float {
	switch v := v.(type) {
	case string:
		vf, err := ParseFloat(v)
		if err != nil {
			log.Println("toFloat: ", err.Error())
			return nil
		}
		return vf

	case byte:
		return NewFloat(float64(v))
	case int:
		return NewFloat(float64(v))
	case int32:
		return NewFloat(float64(v))
	case int64:
		return NewFloat(float64(v))
	case float32:
		return NewFloat(float64(v))
	case float64:
		return NewFloat(v)
	case Float:
		return &v
	case *Float:
		return v
	}

	return nil
}
