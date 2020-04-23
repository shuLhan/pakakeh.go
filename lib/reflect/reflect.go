// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package reflect extends the standard reflect package.
//
package reflect

import (
	"reflect"
)

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

//
// IsEqual is a naive interfaces comparison that check and use Equaler
// interface.
//
func IsEqual(x, y interface{}) bool {
	if x == nil && y == nil {
		return true
	}

	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)

	return isEqual(v1, v2)
}

//nolint: gocyclo
func isEqual(v1, v2 reflect.Value) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}

	t1 := v1.Type()
	t2 := v2.Type()
	if t1 != t2 {
		return false
	}

	k1 := v1.Kind()
	k2 := v2.Kind()
	if k1 != k2 {
		return false
	}

	// For debugging.
	//log.Printf("v1:%v(%s(%v)) v2:%v(%s(%v))", k1, t1.String(), v1,
	//	k2, t2.String(), v2)

	switch k1 {
	case reflect.Bool:
		return v1.Bool() == v2.Bool()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		return v1.Int() == v2.Int()

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() == v2.Uint()

	case reflect.Float32, reflect.Float64:
		return v1.Float() == v2.Float()

	case reflect.Complex64, reflect.Complex128:
		return v1.Complex() == v2.Complex()

	case reflect.Array:
		if v1.Len() != v2.Len() {
			return false
		}
		for x := 0; x < v1.Len(); x++ {
			if !isEqual(v1.Index(x), v2.Index(x)) {
				return false
			}
		}
		return true

	case reflect.Chan:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		return t1 == t2

	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v2.IsNil() {
			return false
		}
		return t1 == t2

	case reflect.Interface:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v2.IsNil() {
			return false
		}
		return isEqual(v1.Elem(), v2.Elem())

	case reflect.Map:
		return isEqualMap(v1, v2)

	case reflect.Ptr:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v2.IsNil() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		return isEqual(v1.Elem(), v2.Elem())

	case reflect.Slice:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v2.IsNil() {
			return false
		}

		l1 := v1.Len()
		l2 := v2.Len()
		if l1 != l2 {
			return false
		}

		for x := 0; x < l1; x++ {
			s1 := v1.Index(x)
			s2 := v2.Index(x)
			if !isEqual(s1, s2) {
				return false
			}
		}
		return true

	case reflect.String:
		return v1.String() == v2.String()

	case reflect.Struct:
		return isEqualStruct(v1, v2)

	case reflect.UnsafePointer:
		return v1.UnsafeAddr() == v2.UnsafeAddr()
	}

	return false
}

func isEqualMap(v1, v2 reflect.Value) bool {
	if v1.IsNil() && v2.IsNil() {
		return true
	}
	if v2.IsNil() {
		return false
	}
	if v1.Len() != v2.Len() {
		return false
	}
	keys := v1.MapKeys()
	for x := 0; x < len(keys); x++ {
		if !isEqual(v1.MapIndex(keys[x]), v2.MapIndex(keys[x])) {
			return false
		}
	}
	return true
}

func isEqualStruct(v1, v2 reflect.Value) bool {
	m1 := v1.MethodByName("IsEqual")
	if m1.IsValid() {
		res := m1.Call([]reflect.Value{
			v2.Addr(),
		})
		if len(res) == 1 && res[0].Kind() == reflect.Bool {
			return res[0].Bool()
		}
	}

	n := v1.NumField()
	for x := 0; x < n; x++ {
		f1 := v1.Field(x)
		f2 := v2.Field(x)
		if !isEqual(f1, f2) {
			return false
		}
	}
	return true
}
