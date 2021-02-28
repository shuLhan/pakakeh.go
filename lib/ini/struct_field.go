// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"fmt"
	"reflect"
	"time"
)

type structField struct {
	sec    string
	sub    string
	key    string
	layout string
	fname  string
	fkind  reflect.Kind
	ftype  reflect.Type
	fval   reflect.Value
}

func (sfield *structField) set(val string) bool {
	rval, ok := unmarshalValue(sfield.ftype, val)
	if ok {
		sfield.fval.Set(rval)
		return true
	}

	switch sfield.fkind {
	case reflect.Array, reflect.Slice:
		slice, ok := sfield.append(val)
		if ok {
			sfield.fval.Set(slice)
			return true
		}

	case reflect.Ptr:
		for sfield.fkind == reflect.Ptr {
			sfield.ftype = sfield.ftype.Elem()
			sfield.fkind = sfield.ftype.Kind()
		}

		if sfield.fval.IsNil() {
			ptrfval := reflect.New(sfield.ftype)
			sfield.fval.Set(ptrfval)
			sfield.fval = ptrfval.Elem()
		} else {
			sfield.fval = sfield.fval.Elem()
		}

		return sfield.set(val)

	case reflect.Struct:
		vi := sfield.fval.Interface()

		_, ok := vi.(time.Time)
		if ok {
			t, err := time.Parse(sfield.layout, val)
			if err != nil {
				return false
			}
			sfield.fval.Set(reflect.ValueOf(t))
			return true
		}
	}
	return false
}

func (sfield *structField) append(val string) (slice reflect.Value, ok bool) {
	ftype := sfield.ftype.Elem()
	slice = sfield.fval

	rval, ok := unmarshalValue(ftype, val)
	if ok {
		slice = reflect.Append(slice, rval)
		return slice, true
	}

	switch ftype.Kind() {
	case reflect.Struct:
		vi := reflect.Zero(ftype).Interface()
		_, ok := vi.(time.Time)
		if ok {
			t, err := time.Parse(sfield.layout, val)
			if err != nil {
				return slice, false
			}
			slice = reflect.Append(slice, reflect.ValueOf(t))
		}

	case reflect.Ptr:
		for ftype.Kind() == reflect.Ptr {
			ftype = ftype.Elem()
		}

		// Create new object with pointer to type, but assign its
		// Elem() to fval for set().
		ptrfval := reflect.New(ftype)
		sliceItem := &structField{
			layout: sfield.layout,
			fkind:  ftype.Kind(),
			ftype:  ftype,
			fval:   ptrfval.Elem(),
		}
		ok = sliceItem.set(val)
		if ok {
			slice = reflect.Append(slice, ptrfval)
		}

	default:
		// Do nothing for other types.
		fmt.Printf("ini: append: unknown type: %v, %s\n", ftype.Kind(), val)
		return slice, false
	}
	return slice, true
}
