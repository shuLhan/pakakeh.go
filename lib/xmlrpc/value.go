// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
)

//
// Value represent dynamic value of XML-RPC type.
//
type Value struct {
	Kind Kind
	// In contains scalar value for Base64, Boolean, Double, Integer,
	// String, and DateTime.
	// It would be nil for Kind of Array and Struct.
	In interface{}

	// Pair of struct member name and its value.
	StructMembers map[string]*Value

	// List of array values.
	ArrayValues []*Value
}

//
// NewValue convert Go type data into XML-RPC value.
//
func NewValue(in interface{}) (out *Value) {
	reft := reflect.TypeOf(in)
	if reft == nil {
		return nil
	}

	refv := reflect.ValueOf(in)

	out = &Value{}

	switch refv.Kind() {
	case reflect.Bool:
		out.Kind = Boolean
		out.In = refv.Bool()

	case reflect.Int:
		if reft.Size() <= 4 {
			out.Kind = Integer
			out.In = int32(refv.Int())
		} else {
			out.Kind = Double
			out.In = float64(refv.Int())
		}

	case reflect.Int8, reflect.Int16, reflect.Int32:
		out.Kind = Integer
		out.In = int32(refv.Int())

	case reflect.Int64:
		out.Kind = Double
		out.In = float64(refv.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		out.Kind = Double
		out.In = float64(refv.Uint())

	case reflect.Float32, reflect.Float64:
		out.Kind = Double
		out.In = refv.Float()

	case reflect.String:
		out.Kind = String
		out.In = refv.String()

	case reflect.Struct:
		out.Kind = Struct
		out.StructMembers = make(map[string]*Value, reft.NumField())
		for x := 0; x < reft.NumField(); x++ {
			var name string

			field := reft.Field(x)
			tag := field.Tag.Get(tagXML)
			if len(tag) > 0 {
				name = tag
			} else {
				name = field.Name
			}

			v := NewValue(refv.Field(x).Interface())
			if v != nil {
				out.StructMembers[name] = v
			}
		}

	case reflect.Array, reflect.Slice:
		out.Kind = Array
		for x := 0; x < refv.Len(); x++ {
			v := NewValue(refv.Index(x).Interface())
			if v != nil {
				out.ArrayValues = append(out.ArrayValues, v)
			}
		}

	case reflect.Interface, reflect.Ptr:
		return NewValue(refv.Elem())

	default:
		return nil
	}

	return out
}

//
// GetFieldAsFloat get struct's field value by name as float64.
//
func (v *Value) GetFieldAsFloat(key string) float64 {
	if v == nil || v.StructMembers == nil {
		return 0
	}
	mv, _ := v.StructMembers[key]
	if mv == nil {
		return 0
	}
	f64, _ := mv.In.(float64)
	return f64
}

//
// GetFieldAsInteger get struct's field value by name as int32.
//
func (v *Value) GetFieldAsInteger(key string) int32 {
	if v == nil || v.StructMembers == nil {
		return 0
	}
	mv, _ := v.StructMembers[key]
	if mv == nil {
		return 0
	}
	i32, _ := mv.In.(int32)
	return i32
}

//
// GetFieldAsString get struct's field value by name as string.
//
func (v *Value) GetFieldAsString(key string) string {
	if v == nil || v.StructMembers == nil {
		return ""
	}
	mv, _ := v.StructMembers[key]
	if mv == nil {
		return ""
	}
	s, _ := mv.In.(string)
	return s
}

func (v *Value) String() string {
	var buf bytes.Buffer

	buf.WriteString("<value>")

	switch v.Kind {
	case String:
		fmt.Fprintf(&buf, "<string>%s</string>", v.In.(string))
	case Boolean:
		fmt.Fprintf(&buf, "<boolean>%t</boolean>", v.In.(bool))
	case Integer:
		fmt.Fprintf(&buf, "<int>%d</int>", v.In.(int32))
	case Double:
		fmt.Fprintf(&buf, "<double>%f</double>", v.In.(float64))
	case DateTime:
		fmt.Fprintf(&buf, "<dateTime.iso8601>%s</dateTime.iso8601>",
			v.In.(string))
	case Base64:
		fmt.Fprintf(&buf, "<base64>%s</base64>", v.In.(string))
	case Struct:
		buf.WriteString("<struct>")
		keys := make([]string, 0, len(v.StructMembers))
		for key := range v.StructMembers {
			if len(key) == 0 {
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&buf, `<member><name>%s</name>%s</member>`,
				key, v.StructMembers[key])
		}
		buf.WriteString("</struct>")
	case Array:
		buf.WriteString("<array><data>")
		for _, val := range v.ArrayValues {
			fmt.Fprintf(&buf, "%s", val.String())
		}
		buf.WriteString("</data></array>")
	}

	buf.WriteString("</value>")

	return buf.String()
}
