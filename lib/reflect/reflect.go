// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package reflect extends the standard reflect package.
//
package reflect

import "reflect"

//
// IsNil will return true if v's type is chan, func, interface, map, pointer,
// or slice and its value is `nil`; otherwise it will return false.
//
func IsNil(v interface{}) bool {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return val.IsNil()
	}
	return false
}
