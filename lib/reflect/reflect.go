// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package reflect extends the standard reflect package.
package reflect

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// DoEqual is a naive interfaces comparison that check and use Equaler
// interface and return an error if its not match.
func DoEqual(x, y interface{}) (err error) {
	if x == nil && y == nil {
		return nil
	}

	var (
		logp               = "DoEqual"
		v1   reflect.Value = reflect.ValueOf(x)
		v2   reflect.Value = reflect.ValueOf(y)
	)
	err = doEqual(v1, v2)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	return nil
}

// IsEqual is a naive interfaces comparison that check and use Equaler
// interface.
func IsEqual(x, y interface{}) bool {
	if x == nil && y == nil {
		return true
	}

	var (
		v1 reflect.Value = reflect.ValueOf(x)
		v2 reflect.Value = reflect.ValueOf(y)
	)
	return isEqual(v1, v2)
}

// IsNil will return true if v's type is chan, func, interface, map, pointer,
// or slice and its value is `nil`; otherwise it will return false.
func IsNil(v interface{}) bool {
	var (
		val reflect.Value = reflect.ValueOf(v)
	)

	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return val.IsNil()
	}
	return v == nil
}

// Set the obj value by converting the string val to the obj type.
//
// If the obj type is an interface or struct, its value will be set by calling
// Unmarshal function.
//
// It will return an error if,
//   - obj is not setable, variable is passed without pointer or pointer
//     not initialized.
//   - val is overflow
//   - obj Kind is Invalid, Array, Chan, Func, Map, or UnsafePointer.
func Set(obj reflect.Value, val string) (err error) {
	var (
		logp                 = "Set"
		objType reflect.Type = obj.Type()
		objKind reflect.Kind = obj.Kind()

		objValue reflect.Value
	)

	if objKind != reflect.Ptr {
		// Variable passed value (V T).
		return fmt.Errorf("%s: object %T is not setable", logp, obj.Interface())
	}

	objValue = obj
	obj = obj.Elem()
	objType = objType.Elem()
	objKind = objType.Kind()

	if objKind == reflect.Ptr {
		// Variable is passed as **T.
		if obj.IsNil() {
			objType = objType.Elem()
			objValue = reflect.New(objType)
			obj.Set(objValue)
		} else {
			objValue = obj
		}
	} else {
		if objValue.IsNil() {
			// Variable is passed as pointer (V *T) but not
			// initialized.
			return fmt.Errorf("%s: object %T is not initialized", logp, obj.Interface())
		}
	}

	switch objKind {
	case reflect.Invalid:
		return fmt.Errorf("%s: object %T is invalid", logp, obj)

	case reflect.Array, reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
		return fmt.Errorf("%s: object %T is not setable", logp, obj.Interface())

	case reflect.Slice:
		objType = objType.Elem()
		objKind = objType.Kind()
		if objKind == reflect.Uint8 {
			err = setValue(objValue, val)
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}
		} else {
			obj, err = setSlice(obj, val)
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}
			objValue.Elem().Set(obj)
		}

	default:
		err = setValue(objValue, val)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}
	return nil
}

// setSlice append the string value of val to the slice.
func setSlice(slice reflect.Value, val string) (sliceOut reflect.Value, err error) {
	var (
		sliceKind reflect.Kind = slice.Kind()
		sliceType reflect.Type = slice.Type()

		elValue  reflect.Value
		ptrValue reflect.Value
	)

	if sliceKind != reflect.Slice {
		return sliceOut, fmt.Errorf("expecting slice, got %T", slice.Interface())
	}

	sliceType = sliceType.Elem() // T = []T
	sliceKind = sliceType.Kind()

	if sliceKind == reflect.Ptr {
		sliceType = sliceType.Elem()
		sliceKind = sliceType.Kind()

		if sliceKind == reflect.Ptr {
			elValue = reflect.New(sliceType)  // var t = new(*T)
			sliceType = sliceType.Elem()      // *T <= **T
			ptrValue = reflect.New(sliceType) // var pt = new(T)
			elValue.Elem().Set(ptrValue)      // *t = pt
		} else {
			ptrValue = reflect.New(sliceType) // t = new(T)
			elValue = ptrValue
		}
	} else {
		ptrValue = reflect.New(sliceType) // var pt = new(T)
		elValue = ptrValue.Elem()         // var t = *pt
	}

	err = setValue(ptrValue, val)
	if err != nil {
		return sliceOut, err
	}

	sliceOut = reflect.Append(slice, elValue)

	return sliceOut, nil
}

// setValue the string value val into reflect.Value based on type of object
// pass in objType.
//
// It will return an error if objType Kind is Invalid, Array, Chan, Func, Map,
// Slice, or UnsafePointer.
func setValue(obj reflect.Value, val string) (err error) {
	var (
		objType = obj.Type()
		objKind = obj.Kind()

		objValue reflect.Value
		v        interface{}
	)

	for objKind == reflect.Ptr {
		objType = objType.Elem()
		objKind = objType.Kind()
	}

	switch objKind {
	case reflect.Bool:
		var vbool bool
		val = strings.ToLower(val)
		if val == "yes" || val == "true" || val == "1" {
			vbool = true
		}
		v = vbool

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var vint64 int64

		if len(val) != 0 {
			vint64, err = strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value: %s", val)
			}
			if obj.Elem().OverflowInt(vint64) {
				return fmt.Errorf("%s value is overflow: %s", objKind, val)
			}
		}
		switch objKind {
		case reflect.Int:
			v = int(vint64)
		case reflect.Int8:
			v = int8(vint64)
		case reflect.Int16:
			v = int16(vint64)
		case reflect.Int32:
			v = int32(vint64)
		default:
			v = vint64
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var vuint64 uint64

		if len(val) != 0 {
			vuint64, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unsigned integer value: %s", val)
			}
			if obj.Elem().OverflowUint(vuint64) {
				return fmt.Errorf("%s value is overflow: %s", objKind, val)
			}
		}
		switch objKind {
		case reflect.Uint:
			v = uint(vuint64)
		case reflect.Uint8:
			v = uint8(vuint64)
		case reflect.Uint16:
			v = uint16(vuint64)
		case reflect.Uint32:
			v = uint32(vuint64)
		default:
			v = vuint64
		}

	case reflect.Float32, reflect.Float64:
		var vf64 float64

		if len(val) != 0 {
			vf64, err = strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("invalid float value: %s", val)
			}
			if obj.Elem().OverflowFloat(vf64) {
				return fmt.Errorf("%s value is overflow: %s", objKind, val)
			}
		}
		if objKind == reflect.Float32 {
			var f32 = float32(vf64)
			v = f32
		} else {
			v = vf64
		}

	case reflect.Slice:
		var (
			sliceType = objType.Elem()
			sliceKind = sliceType.Kind()
		)
		if sliceKind != reflect.Uint8 {
			return fmt.Errorf("cannot convert %s to %s", val, objType)
		}
		v = []byte(val)

	case reflect.String:
		v = val

	case reflect.Interface, reflect.Struct:
		// If type implement UnmarshalBinary, UnmarshalJSON, or
		// UnmarshalText; use it to set the value.

		_, err = Unmarshal(obj, []byte(val))
		if err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("cannot convert %s to %s", val, objType)
	}

	objValue = reflect.ValueOf(v)
	if len(objType.PkgPath()) != 0 {
		// Type is not predeclared (builtin).
		objValue = objValue.Convert(objType)
	}
	obj.Elem().Set(objValue)

	return nil
}

// Unmarshal set the obj value by calling one of the method:
// UnmarshalBinary, UnmarshalJSON, or UnmarshalText; in respective
// order.
//
// Just like reflect, the obj value must be pointer to initialized variable
// (&T) or pointer-to-pointer to uninitialized variable (**T).
//
// If obj implement one of the method, it will return (true, nil) if there is
// no error.
//
// If none of the method exist on obj, it will return (false, nil).
func Unmarshal(obj reflect.Value, val []byte) (ok bool, err error) {
	var (
		logp        = "Unmarshal"
		methodNames = []string{
			"UnmarshalBinary",
			"UnmarshalJSON",
			"UnmarshalText",
		}

		objKind = obj.Kind()
		objType = obj.Type()
		callIn  = []reflect.Value{reflect.ValueOf(val)}

		objValue   reflect.Value
		method     reflect.Value
		callOut    []reflect.Value
		callReturn interface{}
		methodName string
	)

	if objKind != reflect.Ptr {
		// Variable passed as is (V T).
		return false, nil
	}

	objValue = obj
	obj = obj.Elem()
	objType = objType.Elem()
	objKind = objType.Kind()

	if objKind == reflect.Ptr {
		// Variable is passed as **T.
		if obj.IsNil() {
			objType = objType.Elem()
			objValue = reflect.New(objType)
			obj.Set(objValue)
		} else {
			objValue = obj
		}
	} else {
		if objValue.IsNil() {
			// Variable is passed as pointer (V *T) but not
			// initialized.
			return false, nil
		}
	}

	if len(val) == 0 {
		obj.Set(reflect.Zero(objType))
		return true, nil
	}

	for _, methodName = range methodNames {
		method = objValue.MethodByName(methodName)
		if !method.IsValid() {
			// IsValid returns false if method is the zero Value
			// (method not found).
			continue
		}

		callOut = method.Call(callIn)
		if len(callOut) == 0 {
			// No error?
			return true, nil
		}
		callReturn = callOut[0].Interface()
		err, _ = callReturn.(error)
		if err != nil {
			return false, fmt.Errorf("%s: %w", logp, err)
		}
		return true, nil
	}

	return false, nil
}

// Tag simplify lookup on struct's field tag.
//
// Given a StructField and the name of tag, return the tag's value and
// options inside the tag.
// The options is any string after tag's value, separated by comma.
// For example, given the following field definition
//
//	F `tag:"name,opt1, opt2"`
//
// It will return (name, [opt1 opt2], true).
//
// If the field is exported but does not have tag, it will return the field
// name (as is without converting to lower case) in val with hasTag is false:
// (Name, nil, false).
//
// If the field is unexported it will return empty val with hasTag is false
// ("", nil, false).
func Tag(field reflect.StructField, tag string) (val string, opts []string, hasTag bool) {
	if len(field.PkgPath) != 0 {
		// field is unexported.
		return "", nil, false
	}

	var (
		x int
	)

	val, hasTag = field.Tag.Lookup(tag)
	if !hasTag {
		// Tag not defined, so we use field name as key.
		val = field.Name
	} else {
		opts = strings.Split(val, ",")
		for x, val = range opts {
			opts[x] = strings.TrimSpace(val)
		}

		val = opts[0]
		opts = opts[1:]
		if len(val) == 0 {
			// Tag is empty, use field name as key and
			// mark it as not OK.
			val = field.Name
			hasTag = false
		}
	}

	return val, opts, hasTag
}

// doEqual compare two kind of objects and return nils if both are equal.
//
// If its not equal, it will return the interface{} value of v1 and v2 and
// additional error message which describe the type and value where its not
// matched.
func doEqual(v1, v2 reflect.Value) (err error) {
	var (
		in1, in2 interface{}
	)

	if !v1.IsValid() || !v2.IsValid() {
		if v1.IsValid() == v2.IsValid() {
			return nil
		}
		return fmt.Errorf("IsValid: expecting %s(%v), got %s(%v)",
			v1.String(), v1.IsValid(), v2.String(), v2.IsValid())
	}

	var (
		k1    reflect.Kind = v1.Kind()
		k2    reflect.Kind = v2.Kind()
		t1    reflect.Type = v1.Type()
		t2    reflect.Type = v2.Type()
		name1              = t1.Name()
		name2              = t2.Name()

		x int
	)
	if t1 != t2 {
		return fmt.Errorf("Type: expecting %s(%v), got %s(%v)",
			name1, t1.String(), name2, t2.String())
	}
	if k1 != k2 {
		return fmt.Errorf("Kind: expecting %s(%v), got %s(%v)",
			name1, v1.String(), name2, v2.String())
	}
	// For debugging.
	//log.Printf("v1:%v(%s(%v)) v2:%v(%s(%v))", k1, t1.String(), v1, k2, t2.String(), v2)

	if v1.CanSet() {
		in1 = v1.Interface()
	} else if v1.CanAddr() {
		in1 = reflect.NewAt(t1, unsafe.Pointer(v1.UnsafeAddr())).Elem()
	}

	if v2.CanSet() {
		in2 = v2.Interface()
	} else if v2.CanAddr() {
		in2 = reflect.NewAt(t2, unsafe.Pointer(v2.UnsafeAddr())).Elem()
	}

	switch k1 {
	case reflect.Bool:
		if v1.Bool() == v2.Bool() {
			return nil
		}
		return fmt.Errorf("expecting %s(%v), got %s(%v)", name1, v1.Bool(), name2, v2.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		if v1.Int() == v2.Int() {
			return nil
		}
		return fmt.Errorf("expecting %s(%v), got %s(%v)", name1, v1.Int(), name2, v2.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v1.Uint() == v2.Uint() {
			return nil
		}
		return fmt.Errorf("expecting %s(%v), got %s(%v)", name1, v1.Uint(), name2, v2.Uint())

	case reflect.Float32, reflect.Float64:
		if v1.Float() == v2.Float() {
			return nil
		}
		return fmt.Errorf("expecting %s(%v), got %s(%v)", name1, v1.Float(), name2, v2.Float())

	case reflect.Complex64, reflect.Complex128:
		if v1.Complex() == v2.Complex() {
			return nil
		}
		return fmt.Errorf("expecting %s(%v), got %s(%v)", name1, v1.Complex(), name2, v2.Complex())

	case reflect.Array:
		if v1.Len() != v2.Len() {
			return fmt.Errorf("len(%s): expecting %v, got %v", name1, v1.Len(), v2.Len())
		}
		for x = 0; x < v1.Len(); x++ {
			err = doEqual(v1.Index(x), v2.Index(x))
			if err != nil {
				return fmt.Errorf("%s[%d]: %s", name1, x, err)
			}
		}
		return nil

	case reflect.Chan:
		if v1.IsNil() && v2.IsNil() {
			return nil
		}
		if t1 == t2 {
			return nil
		}

	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return nil
		}
		if v2.IsNil() {
			return fmt.Errorf("%s(%v): expecting non-nil, got nil", name1, v1.String())
		}
		if t1 == t2 {
			return nil
		}

	case reflect.Interface:
		if v1.IsNil() && v2.IsNil() {
			return nil
		}
		if v2.IsNil() {
			return fmt.Errorf("%s(%v): expecting non-nil, got nil", name1, in1)
		}
		return doEqual(v1.Elem(), v2.Elem())

	case reflect.Map:
		return doEqualMap(v1, v2)

	case reflect.Ptr:
		if v1.IsNil() && v2.IsNil() {
			return nil
		}
		if v2.IsNil() {
			return fmt.Errorf("%s(%v): expecting non-nil got nil", name1, in1)
		}
		if v1.Pointer() == v2.Pointer() {
			return nil
		}
		return doEqual(v1.Elem(), v2.Elem())

	case reflect.Slice:
		if v1.IsNil() && v2.IsNil() {
			return nil
		}
		if v2.IsNil() {
			return fmt.Errorf("%s(%v): expecting non-nil, got nil", name1, in1)
		}

		var (
			l1 = v1.Len()
			l2 = v2.Len()

			s1, s2 reflect.Value
		)
		if l1 != l2 {
			return fmt.Errorf("len(%s): expecting %v, got %v", name1, l1, l2)
		}

		for x = 0; x < l1; x++ {
			s1 = v1.Index(x)
			s2 = v2.Index(x)
			err = doEqual(s1, s2)
			if err != nil {
				return fmt.Errorf("%s[%d]: %s", name1, x, err)
			}
		}
		return nil

	case reflect.String:
		if v1.String() == v2.String() {
			return nil
		}
		return fmt.Errorf("expecting %s(%v), got %s(%v)", name1, v1.String(), name2, v2.String())

	case reflect.Struct:
		return doEqualStruct(v1, v2)

	case reflect.UnsafePointer:
		if v1.UnsafeAddr() == v2.UnsafeAddr() {
			return nil
		}
	}

	return fmt.Errorf("expecting %s(%v), got %s(%v)", name1, in1, name2, in2)
}

func doEqualMap(v1, v2 reflect.Value) (err error) {
	if v1.IsNil() && v2.IsNil() {
		return nil
	}
	if v2.IsNil() {
		return fmt.Errorf("Map(%s) expecting non-nil, got nil", v1.String())
	}
	if v1.Len() != v2.Len() {
		return fmt.Errorf("len(map(%s)): expecting %d, got %d", v1.Type().Name(), v1.Len(), v2.Len())
	}

	var (
		keys []reflect.Value = v1.MapKeys()

		tipe reflect.Type
		name string
		x    int
	)

	for ; x < len(keys); x++ {
		tipe = keys[x].Type()
		name = tipe.Name()
		err = doEqual(v1.MapIndex(keys[x]), v2.MapIndex(keys[x]))
		if err != nil {
			return fmt.Errorf("Map[%s(%v)] %s", name, keys[x].Interface(), err)
		}
	}
	return nil
}

func doEqualStruct(v1, v2 reflect.Value) (err error) {
	var (
		m1 reflect.Value = v1.MethodByName("IsEqual")

		callIn  []reflect.Value
		callOut []reflect.Value
	)

	if m1.IsValid() {
		callIn = append(callIn, v2.Addr())
		callOut = m1.Call(callIn)
		if len(callOut) == 1 && callOut[0].Kind() == reflect.Bool {
			if callOut[0].Bool() {
				return nil
			}
			return fmt.Errorf("IsEqual: %s.IsEqual(%s) return false", v1.String(), v2.String())
		}
	}

	var (
		t1     reflect.Type = v1.Type()
		v1Name string       = t1.Name()
		n      int          = v1.NumField()

		f1, f2 reflect.Value
		f1Name string
		x      int
	)
	for ; x < n; x++ {
		f1 = v1.Field(x)
		f1Name = t1.Field(x).Name
		f2 = v2.Field(x)

		err = doEqual(f1, f2)
		if err != nil {
			return fmt.Errorf("%s.%s: %s", v1Name, f1Name, err)
		}
	}
	return nil
}

func isEqual(v1, v2 reflect.Value) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}

	var (
		t1 reflect.Type = v1.Type()
		t2 reflect.Type = v2.Type()
	)
	if t1 != t2 {
		return false
	}

	var (
		k1 reflect.Kind = v1.Kind()
		k2 reflect.Kind = v2.Kind()
	)
	if k1 != k2 {
		return false
	}

	// For debugging.
	//log.Printf("v1:%v(%s(%v)) v2:%v(%s(%v))", k1, t1.String(), v1,
	//	k2, t2.String(), v2)

	var (
		x int
	)

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
		for x = 0; x < v1.Len(); x++ {
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

		var (
			l1 = v1.Len()
			l2 = v2.Len()

			s1, s2 reflect.Value
		)
		if l1 != l2 {
			return false
		}

		for x = 0; x < l1; x++ {
			s1 = v1.Index(x)
			s2 = v2.Index(x)
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
	var (
		keys []reflect.Value = v1.MapKeys()

		x int
	)
	for ; x < len(keys); x++ {
		if !isEqual(v1.MapIndex(keys[x]), v2.MapIndex(keys[x])) {
			return false
		}
	}
	return true
}

func isEqualStruct(v1, v2 reflect.Value) bool {
	var (
		m1 reflect.Value = v1.MethodByName("IsEqual")

		callIn  []reflect.Value
		callOut []reflect.Value
		f1, f2  reflect.Value
		n, x    int
	)
	if m1.IsValid() {
		callIn = append(callIn, v2.Addr())
		callOut = m1.Call(callIn)
		if len(callOut) == 1 && callOut[0].Kind() == reflect.Bool {
			return callOut[0].Bool()
		}
	}

	n = v1.NumField()
	for ; x < n; x++ {
		f1 = v1.Field(x)
		f2 = v2.Field(x)
		if !isEqual(f1, f2) {
			return false
		}
	}
	return true
}
