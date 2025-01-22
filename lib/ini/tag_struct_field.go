// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package ini

import (
	"reflect"
	"sort"
	"strings"
	"time"
)

type tagStructField struct {
	v map[string]*structField
}

// unpackTagStructField read each ini tag in the struct's field and store its section,
// subsection, and/or key along with their reflect type and value into
// structField.
func unpackTagStructField(rtype reflect.Type, rval reflect.Value) (out *tagStructField) {
	var (
		tags []string
		tsf  *tagStructField
		key  string
		sf   *structField
	)

	numField := rtype.NumField()
	out = &tagStructField{
		v: make(map[string]*structField, numField),
	}
	if numField == 0 {
		return out
	}

	for x := range numField {
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
				tsf = unpackTagStructField(ftype, fval)
				for key, sf = range tsf.v {
					out.v[key] = sf
				}

			case reflect.Ptr:
				if fval.IsNil() {
					continue
				}
				ftype = ftype.Elem()
				fval = fval.Elem()
				kind := ftype.Kind()
				if kind == reflect.Struct {
					tsf = unpackTagStructField(ftype, fval)
					for key, sf = range tsf.v {
						out.v[key] = sf
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

		tags = ParseTag(tag)
		sfield.sec = tags[0]
		sfield.sub = tags[1]
		sfield.key = strings.ToLower(tags[2])

		if len(sfield.key) == 0 {
			sfield.sec = tags[0]
			sfield.key = sfield.fname
		}

		out.v[tag] = sfield
	}
	return out
}

func (tsf tagStructField) getByKey(key string) (sf *structField) {
	for _, sf = range tsf.v {
		if sf.key == key {
			return sf
		}
	}
	return nil
}

// keys return the map keys sorted in ascending order.
func (tsf tagStructField) keys() (out []string) {
	var (
		key string
	)

	out = make([]string, 0, len(tsf.v))
	for key = range tsf.v {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
