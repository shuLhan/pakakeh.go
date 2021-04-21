// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"bytes"
	"database/sql/driver"
	"math/big"
	"strings"
)

//
// Int extends the standard big.Int package.
//
type Int struct {
	big.Int
}

//
// NewInt create and initialize new Int value from v or nil if v is invalid
// type that cannot be converted to Int.
//
func NewInt(v interface{}) (i *Int) {
	i = &Int{}

	got := toInt(v, i)
	if got == nil {
		return nil
	}
	return i
}

//
// Value implement the sql/driver.Valuer.
//
func (i *Int) Value() (driver.Value, error) {
	var s string = "0"
	if i != nil {
		s = i.String()
	}
	return []byte(s), nil
}

//
// toInt convert any type to Int or nil if type is unknown.
// If in is not nil, it will be set to out.
//
func toInt(v interface{}, in *Int) (out *Int) {
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
