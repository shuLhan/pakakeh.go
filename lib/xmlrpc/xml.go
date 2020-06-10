// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func xmlBegin(dec xml.TokenReader) (err error) {
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("xmlBegin: %w", err)
	}

	el, ok := token.(xml.ProcInst)
	if !ok {
		return fmt.Errorf("xmlBegin: expecting xml.ProcInst got %v",
			el)
	}

	token, err = dec.Token()
	if err != nil {
		return fmt.Errorf("xmlBegin: %w", err)
	}

	_, ok = token.(xml.CharData)
	if !ok {
		return fmt.Errorf("xmlNeBegin: expecting xml.CharData got %v", el)
	}

	return nil
}

func xmlNext(dec xml.TokenReader) (start xml.StartElement, end xml.EndElement, err error) {
	var (
		token xml.Token
		ok    bool
	)

	// Skip any elements that is not StartElement or EndElement ...
	for {
		token, err = dec.Token()
		if err != nil {
			return start, end, fmt.Errorf("xmlNext: %w", err)
		}

		end, ok = token.(xml.EndElement)
		if ok {
			end.Name.Local = strings.ToLower(end.Name.Local)
			break
		}

		start, ok = token.(xml.StartElement)
		if ok {
			start.Name.Local = strings.ToLower(start.Name.Local)
			break
		}
	}

	return start, end, nil
}

func xmlData(dec xml.TokenReader) (data string, err error) {
	token, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("xmlData: %w", err)
	}

	el, ok := token.(xml.CharData)
	if !ok {
		return "", fmt.Errorf("xmlData: expecting xml.CharData got %v", el)
	}

	return string(el), nil
}

func xmlParseScalarValue(dec *xml.Decoder) (v Value, err error) {
	// Get the element type.
	el, _, err := xmlNext(dec)
	if err != nil {
		return v, fmt.Errorf("xmlParseScalarValue: %w", err)
	}

	var data string

	switch el.Name.Local {
	case typeNameBoolean, typeNameBase64, typeNameDateTime,
		typeNameDouble, typeNameInteger, typeNameInteger4,
		typeNameString:

		data, err = xmlData(dec)
		if err != nil {
			return v, fmt.Errorf("xmlParseScalarValue: %w", err)
		}
	}

	switch el.Name.Local {
	case typeNameArray:
		v, err = xmlParseArray(dec)

	case typeNameBoolean:
		v.Kind = Boolean
		if strings.ToLower(data) == boolTrue {
			v.In = true
		} else {
			v.In = false
		}

	case typeNameBase64, typeNameString:
		v.Kind = String
		v.In = data

	case typeNameDateTime:
		v.Kind = DateTime
		v.In, err = time.Parse(timeLayoutISO8601, data)

	case typeNameDouble:
		v.Kind = Double
		v.In, err = strconv.ParseFloat(data, 10)

	case typeNameInteger, typeNameInteger4:
		var i64 int64
		v.Kind = Integer
		i64, err = strconv.ParseInt(data, 10, 64)
		v.In = int32(i64)

	case typeNameStruct:
		v, err = xmlParseStruct(dec)

	default:
		return v, fmt.Errorf("xmlParseScalarValue: unknown scalar type %q",
			el.Name.Local)
	}
	if err != nil {
		return v, fmt.Errorf("xmlParseScalarValue: %w", err)
	}

	// Skip element type closed tag ...
	switch el.Name.Local {
	case typeNameBoolean, typeNameBase64, typeNameDateTime,
		typeNameDouble, typeNameInteger, typeNameInteger4,
		typeNameString:

		err = dec.Skip()
		if err != nil {
			return v, fmt.Errorf("xmlParseScalarValue: %w", err)
		}
	}

	// Skip </value> ...
	err = dec.Skip()
	if err != nil {
		return v, fmt.Errorf("xmlParseScalarValue: %w", err)
	}

	return v, nil
}

func xmlParseArray(dec *xml.Decoder) (arr Value, err error) {
	el, _, err := xmlNext(dec)
	if err != nil {
		return arr, fmt.Errorf("xmlParseArray: %w", err)
	}
	if el.Name.Local != elNameData {
		return arr, fmt.Errorf("xmlParseArray: expecting <%s> got %v",
			elNameData, el)
	}

	arr.Kind = Array

	for {
		start, end, err := xmlNext(dec)
		if err != nil {
			return arr, fmt.Errorf("xmlParseArray: %w", err)
		}
		if end.Name.Local == elNameData {
			break
		}
		if start.Name.Local != elNameValue {
			return arr, fmt.Errorf("xmlParseArray: expecting '<%s>' got %v",
				elNameValue, el)
		}

		v, err := xmlParseScalarValue(dec)
		if err != nil {
			return v, fmt.Errorf("xmlParseArray: %w", err)
		}

		arr.Values = append(arr.Values, v)
	}

	return arr, nil
}

func xmlParseStruct(dec *xml.Decoder) (v Value, err error) {
	v.Kind = Struct

	for {
		start, end, err := xmlNext(dec)
		if err != nil {
			return v, fmt.Errorf("xmlParseStruct: %w", err)
		}
		if end.Name.Local == typeNameStruct {
			break
		}
		if start.Name.Local != elNameMember {
			return v, fmt.Errorf("xmlParseStruct: expecting '<%s>' got %v",
				elNameMember, start.Name.Local)
		}

		m, err := xmlParseStructMember(dec)
		if err != nil {
			return v, fmt.Errorf("xmlParseStruct: %w", err)
		}

		v.Members = append(v.Members, m)

		// skip '</member>'...
		err = dec.Skip()
		if err != nil {
			return v, fmt.Errorf("xmlParseStruct: %w", err)
		}
	}

	return v, nil
}

func xmlParseStructMember(dec *xml.Decoder) (m Member, err error) {
	// member's <name> ...
	el, _, err := xmlNext(dec)
	if err != nil {
		return m, fmt.Errorf("xmlParseStructMember: %w", err)
	}
	if el.Name.Local != elNameName {
		return m, fmt.Errorf("xmlParseStructMember: expecting '<%s>' got %v",
			elNameName, el.Name.Local)
	}

	data, err := xmlData(dec)
	if err != nil {
		return m, fmt.Errorf("xmlParseStructMember: %w", err)
	}

	m.Name = strings.ToLower(data)

	// skip </name> ...
	err = dec.Skip()
	if err != nil {
		return m, fmt.Errorf("xmlParseStructMember: %w", err)
	}

	el, _, err = xmlNext(dec)
	if err != nil {
		return m, fmt.Errorf("xmlParseStructMember: %w", err)
	}
	if el.Name.Local != elNameValue {
		return m, fmt.Errorf("xmlParseStructMember: expecting '<%s>' got %v",
			elNameValue, el.Name.Local)
	}

	m.Value, err = xmlParseScalarValue(dec)
	if err != nil {
		return m, fmt.Errorf("xmlParseStructMember: %w", err)
	}

	return m, nil
}
