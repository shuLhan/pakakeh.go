// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package json provide a library for working with JSON.
//
// This is an extension to standard "encoding/json" package.
//
package json

import (
	"strconv"
	"strings"
)

//
// ToMapStringFloat64 convert the map of string-interface{} into map of
// string-float64.
// This function convert the map's key to lower-cases and ignore zero value in
// interface{}.
// The interface{} value only accept basic numeric types and slice of byte.
//
func ToMapStringFloat64(in map[string]interface{}) (out map[string]float64, err error) {
	out = make(map[string]float64, len(in))

	for k, v := range in {
		var (
			f64 float64
			err error
		)

		switch vv := v.(type) {
		case string:
			f64, err = strconv.ParseFloat(vv, 64)
		case []byte:
			f64, err = strconv.ParseFloat(string(vv), 64)
		case byte:
			f64 = float64(vv)
		case float32:
			f64 = float64(vv)
		case float64:
			f64 = vv
		case int8:
			f64 = float64(vv)
		case int16:
			f64 = float64(vv)
		case int32:
			f64 = float64(vv)
		case int:
			f64 = float64(vv)
		case int64:
			f64 = float64(vv)
		case uint16:
			f64 = float64(vv)
		case uint32:
			f64 = float64(vv)
		case uint64:
			f64 = float64(vv)
		}
		if err != nil {
			return nil, err
		}
		if f64 == 0 {
			continue
		}

		k = strings.ToLower(k)
		out[k] = f64
	}
	return out, nil
}
