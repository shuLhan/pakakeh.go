// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package big

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"math/big"
	"strings"
)

var intZero = NewInt(0)

// Int extends the standard big.Int package.
type Int struct {
	big.Int
}

// NewInt create and initialize new Int value from v or nil if v is invalid
// type that cannot be converted to Int.
func NewInt(v any) (i *Int) {
	i = &Int{}

	got := toInt(v, i)
	if got == nil {
		return nil
	}
	return i
}

// Add set the i value to i + v and return the i as the result.
func (i *Int) Add(v any) *Int {
	vv := toInt(v, nil)
	if vv == nil {
		// Its equal to `i+0`
		return i
	}
	i.Int.Add(&i.Int, &vv.Int)
	return i
}

// IsGreater will return true if i > v.
func (i *Int) IsGreater(v any) bool {
	vv := toInt(v, nil)
	if vv == nil {
		return false
	}
	return i.Cmp(&vv.Int) > 0
}

// IsLess will return true if i < v.
func (i *Int) IsLess(v any) bool {
	vv := toInt(v, nil)
	if vv == nil {
		return false
	}
	return i.Cmp(&vv.Int) < 0
}

// IsZero will return true if `i == 0`.
func (i *Int) IsZero() bool {
	return i.Cmp(&intZero.Int) == 0
}

// MarshalJSON implement the json.Marshaler interface and return the output of
// String method.
//
// If the global variable MarshalJSONAsString is true, the Int value will
// be encoded as string.
func (i *Int) MarshalJSON() ([]byte, error) {
	var s string
	if i == nil {
		s = "0"
	} else {
		s = i.String()
	}
	if MarshalJSONAsString {
		s = `"` + s + `"`
	}
	return []byte(s), nil
}

// Scan implement the database's sql.Scan interface.
func (i *Int) Scan(src any) error {
	got := toInt(src, i)
	if got == nil {
		return fmt.Errorf("Int.Scan: unknown type %T", src)
	}
	return nil
}

// UnmarshalJSON convert the JSON byte value into Int.
func (i *Int) UnmarshalJSON(in []byte) (err error) {
	in = bytes.Trim(in, `"`)
	i.SetInt64(0)
	_, ok := i.Int.SetString(string(in), 10)
	if !ok {
		return fmt.Errorf("Int.UnmarshalJSON: cannot convert %T(%v) to Int", in, in)
	}
	return nil
}

// Value implement the sql/driver.Valuer.
func (i *Int) Value() (driver.Value, error) {
	var s = `0`
	if i != nil {
		s = i.String()
	}
	return []byte(s), nil
}

// toInt convert any type to Int or nil if type is unknown.
// If in is not nil, it will be set to out.
func toInt(v any, in *Int) (out *Int) {
	out = &Int{}

	switch v := v.(type) {
	case []byte:
		if len(v) == 0 {
			out.SetInt64(0)
		} else {
			v = bytes.ReplaceAll(v, []byte{'_'}, nil)
			vals := bytes.Split(v, []byte{'.'})
			_, ok := out.Int.SetString(string(vals[0]), 10)
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
			vals := strings.Split(v, ".")
			_, ok := out.Int.SetString(vals[0], 10)
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
	case uint64:
		out.SetUint64(v)
	case float32:
		out.SetInt64(int64(v))
	case float64:
		out.SetInt64(int64(v))
	case *Int:
		*out = *v
	case *big.Int:
		out.Set(v)
	case *Rat:
		vals := strings.Split(v.String(), ".")
		out.SetString(vals[0], 10)
	case *big.Rat:
		vals := strings.Split(v.String(), ".")
		out.SetString(vals[0], 10)
	default:
		return nil
	}
	if in != nil {
		in.Set(&out.Int)
	} else {
		in = out
	}
	return out
}
