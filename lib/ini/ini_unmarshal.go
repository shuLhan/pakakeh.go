// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// unmarshal set each section-subsection variables into the struct
// fields.
func (in *Ini) unmarshal(tagField *tagStructField) {
	var (
		sec    *Section
		sfield *structField
		v      *variable
		tag    string
		ok     bool
	)
	for _, sec = range in.secs {
		// Search field that tagged with subsection first.
		tag = fmt.Sprintf("%s:%s", sec.nameLower, sec.sub)
		sfield, ok = tagField.v[tag]
		if !ok {
			// Search field that tagged with section name only.
			tag = sec.nameLower
			sfield, ok = tagField.v[tag]
			if !ok {
				// Unmarshal each variable in section-sub into
				// field directly.
				for _, v = range sec.vars {
					tag = fmt.Sprintf("%s:%s:%s", sec.nameLower, sec.sub, v.keyLower)
					sfield, ok = tagField.v[tag]
					if !ok {
						continue
					}
					sfield.set(v.value)
				}
				continue
			}
		}
		switch sfield.fkind {
		case reflect.Map:
			unmarshalToMap(sec, sfield.ftype, sfield.fval)

		case reflect.Ptr:
			for sfield.fkind == reflect.Ptr {
				sfield.ftype = sfield.ftype.Elem()
				sfield.fkind = sfield.ftype.Kind()
			}

			if sfield.fkind == reflect.Struct {
				if sfield.fval.IsNil() {
					ptrfval := reflect.New(sfield.ftype)
					sfield.fval.Set(ptrfval)
					sfield.fval = ptrfval.Elem()
				} else {
					sfield.fval = sfield.fval.Elem()
				}
				unmarshalToStruct(sec, sfield.ftype, sfield.fval)
			}

		case reflect.Slice:
			sliceElem := sfield.ftype.Elem()
			switch sliceElem.Kind() {
			case reflect.Struct:
				newStruct := reflect.New(sliceElem)
				unmarshalToStruct(sec, sliceElem, newStruct.Elem())
				newSlice := reflect.Append(sfield.fval, newStruct.Elem())
				sfield.fval.Set(newSlice)

			case reflect.Ptr:
				// V []*T
				for sliceElem.Kind() == reflect.Ptr {
					sliceElem = sliceElem.Elem()
				}

				if sliceElem.Kind() == reflect.Struct {
					ptrfval := reflect.New(sliceElem)
					unmarshalToStruct(sec, sliceElem, ptrfval.Elem())
					newSlice := reflect.Append(sfield.fval, ptrfval)
					sfield.fval.Set(newSlice)
				}
			}

		case reflect.Struct:
			unmarshalToStruct(sec, sfield.ftype, sfield.fval)
		}

	}
}

// unmarshalToMap unmarshal the Section into a map.
//
// The format is
//
//	V map[S]T `ini:"section:sub"`
//
// for non-struct value or
//
//	V map[S]T `ini:"section"`
//
// for map of struct.
func unmarshalToMap(sec *Section, rtype reflect.Type, rval reflect.Value) bool {
	if rtype.Key().Kind() != reflect.String {
		return false
	}

	var (
		elType = rtype.Elem()
		elKind = elType.Kind()

		v        *variable
		amap     reflect.Value
		astruct  reflect.Value
		mapValue reflect.Value
		isPtr    bool
		ok       bool
	)

	if rval.IsNil() {
		amap = reflect.MakeMap(rtype)
		rval.Set(amap)
	} else {
		amap = rval
	}
	for elKind == reflect.Ptr {
		elType = elType.Elem()
		elKind = elType.Kind()
		isPtr = true
	}
	if elKind == reflect.Struct {
		astruct = reflect.New(elType)

		unmarshalToStruct(sec, elType, astruct.Elem())
		if isPtr {
			amap.SetMapIndex(reflect.ValueOf(sec.sub), astruct)
		} else {
			amap.SetMapIndex(reflect.ValueOf(sec.sub), astruct.Elem())
		}
		return true
	}

	for _, v = range sec.vars {
		if len(v.keyLower) == 0 {
			continue
		}

		mapValue, ok = unmarshalValue(elType, v.value)
		if ok {
			amap.SetMapIndex(reflect.ValueOf(v.keyLower), mapValue)
		}
	}
	rval.Set(amap)
	return true
}

func unmarshalToStruct(sec *Section, rtype reflect.Type, rval reflect.Value) {
	var (
		tagField = unpackTagStructField(rtype, rval)

		v      *variable
		sfield *structField
	)

	for _, v = range sec.vars {
		sfield = tagField.getByKey(v.keyLower)
		if sfield == nil {
			continue
		}
		sfield.set(v.value)
	}
}

// unmarshalValue convert the value from string to primitive type based on its
// kind.
func unmarshalValue(rtype reflect.Type, val string) (rval reflect.Value, ok bool) {
	switch rtype.Kind() {
	case reflect.Bool:
		if IsValueBoolTrue(val) {
			return reflect.ValueOf(true), true
		}
		return reflect.ValueOf(false), true

	case reflect.String:
		return reflect.ValueOf(val), true

	case reflect.Int:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(int(v)), true

	case reflect.Int8:
		v, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(int8(v)), true

	case reflect.Int16:
		v, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(int16(v)), true

	case reflect.Int32:
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(int32(v)), true

	case reflect.Int64:
		vi := reflect.Zero(rtype)
		_, ok := vi.Interface().(time.Duration)
		if ok {
			dur, err := time.ParseDuration(val)
			if err != nil {
				return reflect.Zero(rtype), false
			}
			return reflect.ValueOf(dur), true
		}

		i64, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(i64), true

	case reflect.Uint:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(uint(v)), true

	case reflect.Uint8:
		v, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(uint8(v)), true

	case reflect.Uint16:
		v, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(uint16(v)), true

	case reflect.Uint32:
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(uint32(v)), true

	case reflect.Uint64:
		u64, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(u64), true

	case reflect.Float32:
		f64, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(float32(f64)), true

	case reflect.Float64:
		f64, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return reflect.Zero(rtype), false
		}
		return reflect.ValueOf(f64), true
	}
	return reflect.Zero(rtype), false
}
