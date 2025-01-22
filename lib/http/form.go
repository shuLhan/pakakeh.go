// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package http

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	libreflect "git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
)

const (
	structTagKey = "form"
)

// MarshalForm marshal struct fields tagged with `form:` into [url.Values].
//
// The rules for marshaling follow the same rules as in [UnmarshalForm].
//
// It will return an error if the input is not pointer to or a struct.
func MarshalForm(in any) (out url.Values, err error) {
	var (
		logp    = `MarshalForm`
		inValue = reflect.ValueOf(in)
		inType  = inValue.Type()
		inKind  = inType.Kind()

		listField []reflect.StructField
		x         int
	)

	if inKind == reflect.Ptr {
		inType = inType.Elem()
		inKind = inType.Kind()
	}
	if inKind != reflect.Struct {
		return nil, fmt.Errorf(`%s: expecting struct got %T`, logp, in)
	}

	out = url.Values{}
	listField = reflect.VisibleFields(inType)

	for ; x < len(listField); x++ {
		var field = listField[x]

		if field.Anonymous {
			// Skip embedded field.
			continue
		}

		var (
			key    string
			opts   []string
			hasTag bool
		)

		key, opts, hasTag = libreflect.Tag(field, structTagKey)
		if len(key) == 0 && !hasTag {
			// Field is unexported.
			continue
		}

		var fval = inValue.FieldByIndex(field.Index)

		if libstrings.IsContain(opts, `omitempty`) {
			if libreflect.IsNil(fval) {
				continue
			}
			if fval.IsZero() {
				continue
			}
		}

		var (
			fkind = fval.Kind()

			val  string
			valb []byte
		)

		// Try using one of the method: MarshalBinary, MarshalJSON,
		// or MarshalText; in respective order.
		valb, err = libreflect.Marshal(fval)
		if err != nil {
			return nil, fmt.Errorf(`%s: error marshaling: %w`, logp, err)
		}
		if len(valb) != 0 {
			out.Add(key, string(valb))
			continue
		}

		if fkind == reflect.Slice {
			var (
				sliceType = fval.Type()
				elType    = sliceType.Elem()
			)

			fkind = elType.Kind()

			if fkind == reflect.Uint8 {
				val = fmt.Sprintf(`%s`, fval.Interface())
				out.Add(key, val)
				continue
			}

			var (
				size = fval.Len()

				sliceEl reflect.Value
				y       int
			)
			for ; y < size; y++ {
				sliceEl = fval.Index(y)
				val = fmt.Sprintf(`%v`, sliceEl.Interface())
				out.Add(key, val)
			}
			continue
		}

		val = fmt.Sprintf(`%v`, fval.Interface())
		out.Add(key, val)
	}

	return out, nil
}

// UnmarshalForm read struct fields tagged with `form:` from out as key and
// set its using the value from [url.Values] based on that key.
// If the field does not have `form:` tag but it is exported, then it will
// use the field name, in case insensitive manner.
//
// Only the following types are supported: bool, int/intX, uint/uintX,
// floatX, string, []byte, or type that implement
// [encoding.BinaryUnmarshaler] (UnmarshalBinary), [json.Unmarshaler]
// (UnmarshalJSON), or [encoding.TextUnmarshaler] (UnmarshalText).
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
func UnmarshalForm(in url.Values, out any) (err error) {
	var (
		logp  = `UnmarshalForm`
		vout  = reflect.ValueOf(out)
		rtype = vout.Type()
		rkind = rtype.Kind()

		field      reflect.StructField
		fval       reflect.Value
		fptr       reflect.Value
		key, val   string
		listFields []reflect.StructField
		vals       []string
		x          int
		hasTag     bool
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
	} else if rkind != reflect.Struct {
		return fmt.Errorf(`%s: expecting *T or **T got %T`, logp, out)
	}

	listFields = reflect.VisibleFields(rtype)

	for ; x < len(listFields); x++ {
		field = listFields[x]

		if field.Anonymous {
			// Skip embedded field.
			continue
		}

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
		fval = vout.FieldByIndex(field.Index)
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
