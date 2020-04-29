// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
)

//nolint: gochecknoglobals
var ratZero = NewRat(0)

//
// Rat extend the standard big.Rat using rounding mode ToZero.
//
type Rat struct {
	big.Rat
}

//
// AddRat return the sum of `f+g+...`.
// It will return nil if `f` or `g` is not convertable to Rat.
//
func AddRat(f ...interface{}) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := NewRat(f[0])
	if total == nil {
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x], nil)
		if rx == nil {
			continue
		}
		total.Add(rx)
	}
	return total
}

//
// NewRat create and initialize new Rat value from v or nil if v is invalid
// type that cannot be converted to Rat.
//
func NewRat(v interface{}) (r *Rat) {
	r = &Rat{}

	got := toRat(v, r)
	if got == nil {
		return nil
	}

	return r
}

//
// MulRat return the result of multiplication `f*...`.
// It will return nil if parameter is empty or `f` is not convertable to Rat.
//
func MulRat(f ...interface{}) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := NewRat(f[0])
	if total == nil {
		// Its equal to `0*...`
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x], nil)
		if rx == nil {
			continue
		}
		total.Mul(rx)
	}
	return total
}

//
// MustRat create and initialize new Rat value from v or panic if v is
// unknown type that cannot be converted to Rat.
//
func MustRat(v interface{}) (r *Rat) {
	r = NewRat(v)
	if r == nil {
		log.Fatalf("MustRat: cannot convert %v to Rat", v)
	}
	return r
}

//
// QuoRat return the quotient of `f/g/...` as new Rat.
//
func QuoRat(f ...interface{}) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := NewRat(f[0])
	if total == nil {
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x], nil)
		if rx == nil {
			continue
		}
		total.Quo(rx)
	}
	return total
}

//
// SubRat return the result of subtraction `f-g-...` as new Rat.
//
func SubRat(f ...interface{}) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := NewRat(f[0])
	if total == nil {
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x], nil)
		if rx == nil {
			continue
		}
		total.Sub(rx)
	}
	return total
}

//
// Add sets r to `r+g` and return the r as the result.
//
func (r *Rat) Add(g interface{}) *Rat {
	y := toRat(g, nil)
	if y == nil {
		// Its equal to `r+0`
		return r
	}
	r.Rat.Add(&r.Rat, &y.Rat)
	return r
}

//
// Int64 return the integer resulting from truncating r towards zero.
//
func (r *Rat) Int64() int64 {
	s := strings.Split(r.String(), ".")[0]
	i64, _ := strconv.ParseInt(s, 10, 64)
	return i64
}

//
// IsEqual will return true if `r == g`.
//
func (r *Rat) IsEqual(g interface{}) bool {
	y := toRat(g, nil)
	if y == nil {
		return r == nil
	}
	if r == nil {
		return false
	}
	if r.Cmp(&y.Rat) == 0 {
		return true
	}
	if r.String() == y.String() {
		return true
	}
	return false
}

//
// IsGreater will return true if `r > g`.
//
func (r *Rat) IsGreater(g interface{}) bool {
	y := toRat(g, nil)
	if y == nil {
		return false
	}
	return r.Cmp(&y.Rat) > 0
}

//
// IsGreaterOrEqual will return true if `r >= g`.
//
func (r *Rat) IsGreaterOrEqual(g interface{}) bool {
	y := toRat(g, nil)
	if y == nil {
		return false
	}
	return r.Cmp(&y.Rat) >= 0
}

//
// IsGreaterThanZero will return true if `r > 0`.
//
func (r *Rat) IsGreaterThanZero() bool {
	if r == nil {
		return false
	}
	return r.Cmp(&ratZero.Rat) > 0
}

//
// IsLess will return true if `r < g`.
//
func (r *Rat) IsLess(g interface{}) bool {
	y := toRat(g, nil)
	if y == nil {
		return false
	}
	return r.Cmp(&y.Rat) < 0
}

//
// IsLessOrEqual will return true if `r <= g`.
//
func (r *Rat) IsLessOrEqual(g interface{}) bool {
	y := toRat(g, nil)
	if y == nil {
		return false
	}
	return r.Cmp(&y.Rat) <= 0
}

//
// IsLessThanZero return true if `r < 0`.
//
func (r *Rat) IsLessThanZero() bool {
	return r.Cmp(&ratZero.Rat) < 0
}

//
// IsZero will return true if `r == 0`.
//
func (r *Rat) IsZero() bool {
	return r.Cmp(&ratZero.Rat) == 0
}

//
// MarshalJSON implement the json.Marshaler interface and return the output of
// String method.
//
func (r *Rat) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("0"), nil
	}
	s := r.String()
	return []byte(s), nil
}

//
// Mul sets r to product of `r * g` and return the result as r.
// If g is not convertible to Rat it will return nil.
//
func (r *Rat) Mul(g interface{}) *Rat {
	y := toRat(g, nil)
	if y == nil {
		y = ratZero
	}
	r.Rat.Mul(&r.Rat, &y.Rat)
	r.SetString(r.String())
	return r
}

//
// Quo sets r to quotient of `r/g` and return the result as r.
// If g is not convertible to Rat it will return nil.
//
func (r *Rat) Quo(g interface{}) *Rat {
	y := toRat(g, nil)
	if y == nil {
		return nil
	}
	r.Rat.Quo(&r.Rat, &y.Rat)
	r.SetString(r.String())
	return r
}

//
// Scan implement the database's sql.Scan interface.
//
func (r *Rat) Scan(src interface{}) error {
	got := toRat(src, r)
	if got == nil {
		return fmt.Errorf("Rat.Scan: unknown type %T", src)
	}
	return nil
}

//
// String format the Rat value into string with maximum mantissa is set by
// digit precision option with rounding mode set to zero.
//
// Unlike standard String method, this method will trim trailing zero digit or
// decimal point at the end of mantissa.
//
func (r *Rat) String() string {
	if r == nil {
		return "0"
	}

	b := []byte(r.FloatString(DefaultDigitPrecision + 1))

	nums := bytes.Split(b, []byte{'.'})
	out := string(nums[0])
	if len(nums) == 2 {
		nums[1] = bytes.TrimRight(nums[1], "0")
		nums[1] = bytes.TrimRight(nums[1], ".")

		if len(nums[1]) > DefaultDigitPrecision {
			nums[1] = nums[1][:DefaultDigitPrecision]
			nums[1] = bytes.TrimRight(nums[1], "0")
			nums[1] = bytes.TrimRight(nums[1], ".")
		}

		if len(nums[1]) > 0 {
			out += "." + string(nums[1])
		}
	}

	return out
}

//
// Sub sets r to rounded difference `r-g` and return r.
//
func (r *Rat) Sub(g interface{}) *Rat {
	y := toRat(g, nil)
	if y == nil {
		// Its equal to `r-0`.
		return r
	}
	r.Rat.Sub(&r.Rat, &y.Rat)
	return r
}

//
// UnmarshalJSON convert the JSON byte value into Rat.
//
func (r *Rat) UnmarshalJSON(in []byte) (err error) {
	in = bytes.Trim(in, `"`)
	r.SetInt64(0)
	_, ok := r.Rat.SetString(string(in))
	if !ok {
		return fmt.Errorf("Rat.UnmarshalJSON:"+
			" cannot convert %T(%v) to Rat", in, in)
	}
	return nil
}

//
// toRat convert any type to Rat or nil if type is unknown.
// If in is not nil, it will be set to out.
//
func toRat(v interface{}, in *Rat) (out *Rat) {
	out = &Rat{}

	switch v := v.(type) {
	case []byte:
		if len(v) == 0 {
			out.SetInt64(0)
		} else {
			v = bytes.ReplaceAll(v, []byte{'_'}, nil)
			_, ok := out.Rat.SetString(string(v))
			if !ok {
				return nil
			}
		}
	case string:
		if len(v) == 0 {
			out.SetInt64(0)
		} else {
			// Replace the underscore character, so we can write the
			// number as "0.000_000_1".
			v = strings.ReplaceAll(v, "_", "")
			_, ok := out.Rat.SetString(v)
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
	case float32:
		out.SetFloat64(float64(v))
	case float64:
		out.SetFloat64(v)
	case Rat:
		out = &v
	case *Rat:
		out = v
	case big.Rat:
		out.Rat = v
	case *big.Rat:
		out.Rat = *v
	default:
		return nil
	}
	if in != nil {
		in.Set(&out.Rat)
	} else {
		in = out
	}
	return out
}
