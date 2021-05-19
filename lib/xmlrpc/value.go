// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"bytes"
	"fmt"
	"reflect"
)

//
// Value represent dynamic value of XML-RPC type.
//
type Value struct {
	Kind Kind
	// In contains scalar value for Base64, Boolean, Double, Integer,
	// String, and DateTime.
	// It would be nil for Kind of Array and Struct.
	In            interface{}
	StructMembers []*Member // List of struct's member.
	ArrayValues   []*Value  // List of array's items.
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
		for x := 0; x < reft.NumField(); x++ {
			m := &Member{}

			field := reft.Field(x)
			tag := field.Tag.Get(tagXML)
			if len(tag) > 0 {
				m.Name = tag
			} else {
				m.Name = field.Name
			}

			v := NewValue(refv.Field(x).Interface())
			if v != nil {
				m.Value = v
				out.StructMembers = append(out.StructMembers, m)
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
func (v Value) GetFieldAsFloat(key string) float64 {
	for _, m := range v.StructMembers {
		if m.Name == key {
			s, _ := m.Value.In.(float64)
			return s
		}
	}

	return 0
}

//
// GetFieldAsInteger get struct's field value by name as int32.
//
func (v Value) GetFieldAsInteger(key string) int32 {
	for _, m := range v.StructMembers {
		if m.Name == key {
			s, _ := m.Value.In.(int32)
			return s
		}
	}

	return 0
}

//
// GetFieldAsString get struct's field value by name as string.
//
func (v Value) GetFieldAsString(key string) string {
	for _, m := range v.StructMembers {
		if m.Name == key {
			s, _ := m.Value.In.(string)
			return s
		}
	}

	return ""
}

func (v Value) String() string {
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
		for _, member := range v.StructMembers {
			buf.WriteString(member.String())
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
