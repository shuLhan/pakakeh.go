// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	libreflect "github.com/shuLhan/share/lib/reflect"
)

const (
	structTagKey = "form"
)

// UnmarshalForm read struct fields tagged with `form:` from out as key and
// set its using the value from url.Values based on that key.
// If the field does not have `form:` tag but it is exported, then it will use
// the field name, in case insensitive.
//
// Only the following types are supported: bool, int/intX, uint/uintX,
// floatX, string, []byte, or type that implement BinaryUnmarshaler
// (UnmarshalBinary), json.Unmarshaler (UnmarshalJSON), or TextUnmarshaler
// (UnmarshalText).
//
// A bool type can be set to true using the following string value: "true",
// "yes", or "1".
//
// If the input contains multiple values but the field type is not slice,
// the field will be set using the first value.
//
// It will return an error if the out variable is not set-able (the type is
// not a pointer to a struct).
// It will not return an error if one of the input value is not match with
// field type.
func UnmarshalForm(in url.Values, out interface{}) (err error) {
	var (
		logp                = "UnmarshalForm"
		vout  reflect.Value = reflect.ValueOf(out)
		rtype reflect.Type  = vout.Type()
		rkind reflect.Kind  = rtype.Kind()

		tstruct  reflect.Type
		field    reflect.StructField
		fval     reflect.Value
		fptr     reflect.Value
		key, val string
		vals     []string
		x        int
		hasTag   bool
	)

	if rkind != reflect.Ptr {
		return fmt.Errorf("%s: expecting *T got %T", logp, out)
	}

	if vout.IsNil() {
		return fmt.Errorf("%s: %T is not initialized", logp, out)
	}

	vout = vout.Elem()
	rtype = rtype.Elem()
	rkind = rtype.Kind()
	if rkind == reflect.Ptr {
		rtype = rtype.Elem()
		rkind = rtype.Kind()
		if rkind != reflect.Struct {
			return fmt.Errorf("%s: expecting *T or **T got %T", logp, out)
		}

		if vout.IsNil() {
			vout.Set(reflect.New(rtype)) // vout = new(T)
			vout = vout.Elem()
		} else {
			vout = vout.Elem()
		}
	} else {
		if rkind != reflect.Struct {
			return fmt.Errorf("%s: expecting *T or **T got %T", logp, out)
		}
	}

	tstruct = rtype

	for ; x < vout.NumField(); x++ {
		field = tstruct.Field(x)

		key, _, hasTag = libreflect.Tag(field, structTagKey)
		if len(key) == 0 && !hasTag {
			// Field is unexported.
			continue
		}

		vals = in[key]
		if len(vals) == 0 {
			if hasTag {
				// Tag is defined and not empty, but no value
				// in input.
				continue
			}

			// Tag is not defined, search lower case field name.
			key = strings.ToLower(key)
			vals = in[key]
			if len(vals) == 0 {
				continue
			}
		}

		// Now that we have the value, store it into field by its
		// type.
		fval = vout.Field(x)
		rtype = fval.Type()
		rkind = fval.Kind()

		if rkind == reflect.Ptr {
			// F *T
			rtype = rtype.Elem() // T <= *T
			rkind = rtype.Kind()
			fptr = reflect.New(rtype) // f = new(T)
			fval.Set(fptr)            // F = f
		} else {
			// F T
			fptr = fval.Addr() // f = &F
		}

		if len(vals) > 1 {
			if rkind == reflect.Slice {
				for _, val = range vals {
					err = libreflect.Set(fptr, val)
					if err != nil {
						continue
					}
				}
			} else {
				// Form contains multiple values, use
				// only the first value to set the field.
				val = vals[0]
				err = libreflect.Set(fptr, val)
				if err != nil {
					continue
				}
			}
		} else {
			val = vals[0]
			err = libreflect.Set(fptr, val)
			if err != nil {
				continue
			}
		}
	}
	return nil
}
