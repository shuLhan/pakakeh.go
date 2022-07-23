// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"reflect"
	"sort"
	"strings"
	"time"
)

type tagStructField map[string]*structField

// unpackTagStructField read each ini tag in the struct's field and store its section,
// subsection, and/or key along with their reflect type and value into
// structField.
func unpackTagStructField(rtype reflect.Type, rval reflect.Value) (out tagStructField) {
	var (
		tags []string
	)

	numField := rtype.NumField()
	if numField == 0 {
		return nil
	}

	out = make(tagStructField, numField)

	for x := 0; x < numField; x++ {
		field := rtype.Field(x)
		fval := rval.Field(x)
		ftype := field.Type
		fkind := ftype.Kind()

		if !fval.CanSet() {
			continue
		}

		tag := strings.TrimSpace(field.Tag.Get(fieldTagName))
		if len(tag) == 0 {
			switch fkind {
			case reflect.Struct:
				for k, v := range unpackTagStructField(ftype, fval) {
					out[k] = v
				}

			case reflect.Ptr:
				if fval.IsNil() {
					continue
				}
				ftype = ftype.Elem()
				fval = fval.Elem()
				kind := ftype.Kind()
				if kind == reflect.Struct {
					for k, v := range unpackTagStructField(ftype, fval) {
						out[k] = v
					}
				}
			}
			continue
		}

		sfield := &structField{
			fname: strings.ToLower(field.Name),
			fkind: fkind,
			ftype: ftype,
			fval:  fval,
		}

		sfield.layout = field.Tag.Get("layout")
		if len(sfield.layout) == 0 {
			sfield.layout = time.RFC3339
		}

		tags = parseTag(tag)
		sfield.sec = tags[0]
		sfield.sub = tags[1]
		sfield.key = strings.ToLower(tags[2])

		if len(sfield.key) == 0 {
			sfield.sec = tags[0]
			sfield.key = sfield.fname
		}

		out[tag] = sfield
	}
	return out
}

func (tsf tagStructField) getByKey(key string) *structField {
	for _, f := range tsf {
		if f.key == key {
			return f
		}
	}
	return nil
}

// keys return the map keys sorted in ascending order.
func (tsf tagStructField) keys() (out []string) {
	out = make([]string, 0, len(tsf))
	for k := range tsf {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
