// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package big

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/runes"
)

var ratZero = NewRat(0)

// Rat extend the standard big.Rat using rounding mode ToZero and without
// panic.
type Rat struct {
	big.Rat
}

// AddRat return the sum of `f[0]+f[1]+...`.
// It will return nil if the first parameter is not convertable to Rat.
func AddRat(f ...any) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := toRat(f[0])
	if total == nil {
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x])
		if rx == nil {
			continue
		}
		total.Add(rx)
	}
	return total
}

// NewRat create and initialize new Rat value from v.
// It will return nil if v is not convertable to Rat.
//
// Empty string or empty []byte still considered as valid, and it will return
// it as zero.
func NewRat(v any) (r *Rat) {
	return toRat(v)
}

// MulRat return the result of multiplication `f[0]*f[1]*...`.
// It will return nil if the first parameter is not convertable to Rat.
func MulRat(f ...any) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := toRat(f[0])
	if total == nil {
		// Its equal to `0*...`
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x])
		if rx == nil {
			continue
		}
		total.Mul(rx)
	}
	return total
}

// QuoRat return the quotient of `f[0]/f[1]/...` as new Rat.
// It will return nil if the first parameter is not convertable to Rat.
// If the second or rest of parameters can not be converted to Rat or zero it
// will return nil instead of panic.
func QuoRat(f ...any) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := toRat(f[0])
	if total == nil {
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x])
		if rx == nil || rx.IsZero() {
			return nil
		}
		total.Quo(rx)
	}
	return total
}

// SubRat return the result of subtraction `f[0]-f[1]-...` as new Rat.
// It will return nil if the first parameter is not convertable to Rat.
func SubRat(f ...any) *Rat {
	if len(f) == 0 {
		return nil
	}
	total := toRat(f[0])
	if total == nil {
		return nil
	}
	for x := 1; x < len(f); x++ {
		rx := toRat(f[x])
		if rx == nil {
			continue
		}
		total.Sub(rx)
	}
	return total
}

// Abs sets r to |r| (the absolute value of r) and return it.
func (r *Rat) Abs() *Rat {
	if r == nil {
		return nil
	}
	r.Rat.Abs(&r.Rat)
	return r
}

// Add sets r to `r+g` and return the r as the result.
// If g is not convertable to Rat it will equal to r+0.
func (r *Rat) Add(g any) *Rat {
	if r == nil {
		return nil
	}
	y := toRat(g)
	if y == nil {
		// Its equal to `r+0`
		return r
	}
	r.Rat.Add(&r.Rat, &y.Rat)
	return r
}

// Humanize format the r into string with custom thousand and decimal
// separator.
func (r *Rat) Humanize(thousandSep, decimalSep string) string {
	if r == nil {
		return "0"
	}
	var (
		raw     = r.String()
		parts   = strings.SplitN(raw, ".", 2)
		intPart = []rune(parts[0])
		out     = make([]rune, 0, len(intPart)+(len(intPart)/3))
		i       = 0
	)
	for x := len(intPart) - 1; x >= 0; x-- {
		if i%3 == 0 && len(out) > 0 {
			out = append(out, []rune(thousandSep)...)
		}
		out = append(out, intPart[x])
		i++
	}
	out = runes.Inverse(out)
	if len(parts) == 2 {
		out = append(out, []rune(decimalSep)...)
		out = append(out, []rune(parts[1])...)
	}
	return string(out)
}

// Int64 return the integer resulting from truncating r towards zero.
// It will return math.MaxInt64, if the value is larger than MaxInt64.
// It will return math.MinInt64, if the value is lower than MinInt64.
func (r *Rat) Int64() int64 {
	if r == nil {
		return 0
	}
	s := strings.Split(r.String(), ".")[0]
	i64, _ := strconv.ParseInt(s, 10, 64)
	return i64
}

// IsEqual will return true if `r == g`, including when r and g are both nil.
//
// Unlike the standard Cmp(), if the first call to Cmp is not 0, it will try
// to compare the string values of r and g, truncated by
// DefaultDigitPrecision.
func (r *Rat) IsEqual(g any) bool {
	y := toRat(g)
	if y == nil {
		return r == nil
	}
	if r == nil {
		return false
	}
	if r.Rat.Cmp(&y.Rat) == 0 {
		return true
	}
	if r.String() == y.String() {
		return true
	}
	return false
}

// IsGreater will return true if `r > g`.
// If g is not convertable to Rat it will return false.
func (r *Rat) IsGreater(g any) bool {
	if r == nil {
		return false
	}
	y := toRat(g)
	if y == nil {
		return false
	}
	return r.Rat.Cmp(&y.Rat) > 0
}

// IsGreaterOrEqual will return true if `r >= g`.
// If g is not convertable to Rat it will return false.
func (r *Rat) IsGreaterOrEqual(g any) bool {
	if r == nil {
		return false
	}
	y := toRat(g)
	if y == nil {
		return false
	}
	return r.Rat.Cmp(&y.Rat) >= 0
}

// IsGreaterThanZero will return true if `r > 0`.
func (r *Rat) IsGreaterThanZero() bool {
	if r == nil {
		return false
	}
	return r.Rat.Cmp(&ratZero.Rat) > 0
}

// IsLess will return true if `r < g`.
// If r is nill or g is not convertable to Rat it will return false.
func (r *Rat) IsLess(g any) bool {
	if r == nil {
		return false
	}
	y := toRat(g)
	if y == nil {
		return false
	}
	return r.Rat.Cmp(&y.Rat) < 0
}

// IsLessOrEqual will return true if `r <= g`.
// It r is nil or g is not convertable to Rat it will return false.
func (r *Rat) IsLessOrEqual(g any) bool {
	if r == nil {
		return false
	}
	y := toRat(g)
	if y == nil {
		return false
	}
	return r.Rat.Cmp(&y.Rat) <= 0
}

// IsLessThanZero return true if `r < 0`.
func (r *Rat) IsLessThanZero() bool {
	if r == nil {
		return false
	}
	return r.Rat.Cmp(&ratZero.Rat) < 0
}

// IsZero will return true if `r == 0`.
func (r *Rat) IsZero() bool {
	if r == nil {
		return false
	}
	return r.Rat.Cmp(&ratZero.Rat) == 0
}

// MarshalJSON implement the json.Marshaler interface.
// It will return the same result as String().
func (r *Rat) MarshalJSON() ([]byte, error) {
	var s string
	if r == nil {
		if MarshalJSONAsString {
			s = `"0"`
		} else {
			s = `0`
		}
	} else {
		s = r.String()
		if MarshalJSONAsString {
			s = `"` + s + `"`
		}
	}
	return []byte(s), nil
}

// Mul sets r to product of `r * g` and return the result as r.
// If g is not convertible to Rat it will return nil.
func (r *Rat) Mul(g any) *Rat {
	y := toRat(g)
	if y == nil {
		return nil
	}
	r.Rat.Mul(&r.Rat, &y.Rat)

	// This security issue has been fixed since Go 1.17.7,
	// - https://groups.google.com/g/golang-announce/c/SUsQn0aSgPQ
	// - https://github.com/golang/go/issues/50699
	r.Rat.SetString(r.String())

	return r
}

// Quo sets r to quotient of `r/g` and return the result as r.
// If r is nil or g is not convertible to Rat or zero it will return nil.
func (r *Rat) Quo(g any) *Rat {
	if r == nil {
		return nil
	}
	y := toRat(g)
	if y == nil || y.IsZero() {
		return nil
	}
	r.Rat.Quo(&r.Rat, &y.Rat)
	r.Rat.SetString(r.String())
	return r
}

// RoundNearestFraction round the fraction to the nearest non-zero value.
//
// The RoundNearestFraction does not require precision parameter, like in
// other rounds function, but it will figure it out based on the last non-zero
// value from fraction.
//
// See example for more information.
func (r *Rat) RoundNearestFraction() *Rat {
	if r == nil {
		return nil
	}
	b := []byte(r.String())
	nums := bytes.Split(b, []byte{'.'})
	x := 0
	if len(nums) == 2 {
		for ; x < len(nums[1]); x++ {
			if nums[1][x] != '0' {
				break
			}
		}
	}
	return r.RoundToNearestAway(x + 1)
}

// RoundToNearestAway round r to n digit precision using nearest away mode,
// where mantissa is accumulated by the last digit after precision.
// For example, using 2 digit precision, 0.555 would become 0.56.
func (r *Rat) RoundToNearestAway(prec int) *Rat {
	if r == nil {
		return nil
	}
	r.Rat.SetString(r.FloatString(prec))
	return r
}

// RoundToZero round r to n digit precision using to zero mode.
// For example, using 2 digit precision, 0.555 would become 0.55.
func (r *Rat) RoundToZero(prec int) *Rat {
	if r == nil {
		return nil
	}
	b := []byte(r.FloatString(prec + 1))
	nums := bytes.Split(b, []byte{'.'})
	b = b[:0]
	b = append(b, nums[0]...)
	if len(nums) == 2 && prec > 0 {
		b = append(b, '.')
		b = append(b, nums[1][:prec]...)
	}
	r.Rat.SetString(string(b))
	return r
}

// Scan implement the database's sql.Scan interface.
func (r *Rat) Scan(v any) error {
	got := toRat(v)
	if got == nil {
		return fmt.Errorf("Rat.Scan: unknown type %T", v)
	}
	r.Rat.Set(&got.Rat)
	return nil
}

// String format the Rat value into string with maximum mantissa is set by
// digit precision option with rounding mode set to zero.
//
// Unlike standard String method, this method will trim trailing zero digit or
// decimal point at the end of mantissa.
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

// Sub sets r to rounded difference `r-g` and return r.
// If g is not convertable to Rat, it will return as r-0.
func (r *Rat) Sub(g any) *Rat {
	if r == nil {
		return nil
	}
	y := toRat(g)
	if y == nil {
		// Its equal to `r-0`.
		return r
	}
	r.Rat.Sub(&r.Rat, &y.Rat)
	return r
}

// UnmarshalJSON convert the JSON byte value into Rat.
func (r *Rat) UnmarshalJSON(in []byte) (err error) {
	in = bytes.Trim(in, `"`)
	r.SetInt64(0)
	_, ok := r.Rat.SetString(string(in))
	if !ok {
		return fmt.Errorf("Rat.UnmarshalJSON: cannot convert %T(%v) to Rat", in, in)
	}
	return nil
}

// Value return the []byte value for database/sql, as defined in
// sql/driver.Valuer.
// It will return "0" if r is nil.
func (r *Rat) Value() (driver.Value, error) {
	var s string
	if r == nil {
		s = "0"
	} else {
		s = r.String()
	}
	return []byte(s), nil
}

// toRat convert v type to Rat or nil if v type is unknown.
func toRat(g any) (out *Rat) {
	out = &Rat{}

	switch v := g.(type) {
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
		if len(v) == 0 || v == "0" {
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
	case Rat:
		out.Rat.Set(&v.Rat)
	case *Rat:
		if v == nil {
			out.SetInt64(0)
		} else {
			out.Rat.Set(&v.Rat)
		}
	case big.Rat:
		out.Rat.Set(&v)
	case *big.Rat:
		if v == nil {
			out.SetInt64(0)
		} else {
			out.Rat.Set(v)
		}
	case *big.Int:
		if v == nil {
			out.Rat.SetInt64(0)
		} else {
			out.Rat.SetInt(v)
		}
	default:
		return nil
	}
	return out
}
