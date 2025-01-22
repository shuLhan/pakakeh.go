// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package xmlrpc

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
)

// Value represent dynamic value of XML-RPC type.
type Value struct {
	// In contains scalar value for Base64, Boolean, Double, Integer,
	// String, and DateTime.
	// It would be nil for Kind of Array and Struct.
	In any

	// Pair of struct member name and its value.
	StructMembers map[string]*Value

	// List of array values.
	ArrayValues []*Value

	Kind Kind
}

// NewValue convert Go type data into XML-RPC value.
func NewValue(in any) (out *Value) {
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

	case reflect.Uint8, reflect.Uint16:
		out.Kind = Integer
		out.In = int32(refv.Uint())

	case reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
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
		for x := range reft.NumField() {
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
		for x := range refv.Len() {
			v := NewValue(refv.Index(x).Interface())
			if v != nil {
				out.ArrayValues = append(out.ArrayValues, v)
			}
		}

	case reflect.Interface, reflect.Ptr:
		return NewValue(refv.Elem())

	case reflect.Invalid, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Map,
		reflect.UnsafePointer:
		return nil
	}

	return out
}

// GetFieldAsFloat get struct's field value by name as float64.
func (v *Value) GetFieldAsFloat(key string) float64 {
	if v == nil || v.StructMembers == nil {
		return 0
	}
	mv := v.StructMembers[key]
	if mv == nil {
		return 0
	}
	f64, ok := mv.In.(float64)
	if !ok {
		return 0
	}
	return f64
}

// GetFieldAsBoolean get the struct's field value by its key as boolean.
func (v *Value) GetFieldAsBoolean(key string) bool {
	if v == nil || v.StructMembers == nil {
		return false
	}
	mv := v.StructMembers[key]
	if mv == nil {
		return false
	}
	abool, ok := mv.In.(bool)
	if !ok {
		return false
	}
	return abool
}

// GetFieldAsInteger get struct's field value by name as int.
func (v *Value) GetFieldAsInteger(key string) int {
	if v == nil || v.StructMembers == nil {
		return 0
	}
	mv := v.StructMembers[key]
	if mv == nil {
		return 0
	}
	i32, ok := mv.In.(int32)
	if !ok {
		return 0
	}
	return int(i32)
}

// GetFieldAsString get struct's field value by name as string.
func (v *Value) GetFieldAsString(key string) string {
	if v == nil || v.StructMembers == nil {
		return ""
	}
	mv := v.StructMembers[key]
	if mv == nil {
		return ""
	}
	s, ok := mv.In.(string)
	if !ok {
		return fmt.Sprintf("%v", mv.In)
	}
	return s
}

func (v *Value) String() string {
	var buf bytes.Buffer

	buf.WriteString("<value>")

	switch v.Kind {
	case Unset:
		// NOOP.
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
