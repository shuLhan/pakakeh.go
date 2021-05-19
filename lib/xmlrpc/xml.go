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

func xmlBegin(dec *xml.Decoder) (err error) {
	token, err := dec.Token()
	if err != nil {
		return err
	}

	_, ok := token.(xml.ProcInst)
	if !ok {
		return fmt.Errorf(`xmlBegin: expecting <?xml version="1.0"?> got %T %+v`, token, token)
	}

	return nil
}

//
// xmlMustCData parse the CDATA inside the tag.
//
func xmlMustCData(dec *xml.Decoder, tag string) (cdata string, err error) {
	err = xmlMustStart(dec, tag)
	if err != nil {
		return "", err
	}

	token, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("expecting CDATA, got an error %w", err)
	}

	found := false
	for !found {
		switch tok := token.(type) {
		case xml.CharData:
			cdata = string(tok)
			found = true

		case xml.EndElement:
			if tok.Name.Local != tag {
				return "", fmt.Errorf("expecting </%s>, got token %+v", tag, tok)
			}
			return "", nil

		default:
			token, err = dec.Token()
			if err != nil {
				return "", fmt.Errorf("expecting CDATA, got an error %w", err)
			}
		}
	}

	err = xmlMustEnd(dec, tag)
	if err != nil {
		return "", err
	}

	return cdata, nil
}

//
// xmlMustStart parse the first XML element that must start with passed
// openTag.
//
func xmlMustStart(dec *xml.Decoder, tag string) (err error) {
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("expecting <%s>, got an error %w", tag, err)
	}

	found := false
	for !found {
		switch tok := token.(type) {
		case xml.StartElement:
			if tok.Name.Local != tag {
				return fmt.Errorf("expecting <%s> got token %+v", tag, tok)
			}
			found = true

		case xml.Comment, xml.CharData:
			token, err = dec.Token()
			if err != nil {
				return fmt.Errorf("expecting <%s>, got an error %w", tag, err)
			}

		default:
			return fmt.Errorf("expecting <%s>, got token %T", tag, tok)
		}
	}
	return nil
}

//
// xmlMustEnd parse the next XML element that must be a closed tag.
//
func xmlMustEnd(dec *xml.Decoder, tag string) (err error) {
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("expecting </%s>, got an error %w", tag, err)
	}

	found := false
	for !found {
		switch tok := token.(type) {
		case xml.EndElement:
			if tok.Name.Local != tag {
				return fmt.Errorf("expecting </%s>, got token </%s>", tag, tok.Name.Local)
			}
			found = true

		case xml.Comment, xml.CharData:
			token, err = dec.Token()
			if err != nil {
				return fmt.Errorf("expecting </%s>, got an error %w", tag, err)
			}

		default:
			return fmt.Errorf("expecting </%s>, got an error %w", tag, err)
		}
	}
	return nil
}

func xmlStart(dec *xml.Decoder, openTag, closeTag string) (isOpen bool, err error) {
	token, err := dec.Token()
	if err != nil {
		return false, fmt.Errorf("expecting <%s>, got an error %w", openTag, err)
	}

	for !isOpen {
		switch tok := token.(type) {
		case xml.StartElement:
			if tok.Name.Local != openTag {
				return false, fmt.Errorf("expecting <%s>, got <%s>",
					openTag, tok.Name.Local)
			}
			isOpen = true

		case xml.EndElement:
			if tok.Name.Local == closeTag {
				return false, nil
			}
			return false, fmt.Errorf("expecting </%s>, got </%s>",
				closeTag, tok.Name.Local)

		case xml.CharData, xml.Comment:
			token, err = dec.Token()
			if err != nil {
				return false, fmt.Errorf("expecting <%s>, got an error %w",
					openTag, err)
			}
		default:
			return false, fmt.Errorf("expecting <%s>, got token %T %+v",
				openTag, token, tok)
		}
	}

	return true, nil
}

//
// xmlParseParams parse the optional <params> elements.
//
func xmlParseParams(dec *xml.Decoder, closeTag string) (params []*Value, hasParams bool, err error) {
	isOpen, err := xmlStart(dec, elNameParams, closeTag)
	if err != nil {
		return nil, false, err
	}
	if !isOpen {
		return nil, false, nil
	}

	for {
		param, err := xmlParseParam(dec, elNameParams)
		if err != nil {
			return nil, hasParams, err
		}
		if param == nil {
			break
		}
		params = append(params, param)
	}

	return params, true, nil
}

//
// xmlParseParam parse the <param> element.
//
func xmlParseParam(dec *xml.Decoder, closeTag string) (param *Value, err error) {
	isOpen, err := xmlStart(dec, elNameParam, closeTag)
	if err != nil {
		return nil, err
	}
	if !isOpen {
		return nil, nil
	}

	param, err = xmlParseValue(dec, elNameParam)
	if err != nil {
		return nil, err
	}

	if param != nil {
		err = xmlMustEnd(dec, elNameParam)
		if err != nil {
			return nil, err
		}
	}

	return param, nil
}

func xmlParseValue(dec *xml.Decoder, closeTag string) (param *Value, err error) {
	var (
		cdata string
	)

	isOpen, err := xmlStart(dec, elNameValue, closeTag)
	if err != nil {
		return nil, err
	}
	if !isOpen {
		return nil, nil
	}

	param = &Value{}

	token, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("expecting CDATA, got an error %w", err)
	}

	found := false
	for !found {
		switch tok := token.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case typeNameArray:
				param, err = xmlParseArray(dec)
				if err != nil {
					return nil, err
				}
				err = xmlMustEnd(dec, elNameValue)
				if err != nil {
					return nil, err
				}
				return param, nil

			case typeNameStruct:
				param, err = xmlParseStruct(dec)
				if err != nil {
					return nil, err
				}
				err = xmlMustEnd(dec, elNameValue)
				if err != nil {
					return nil, err
				}
				return param, nil

			case typeNameBase64:
				param.Kind = Base64
				param.In = ""
			case typeNameBoolean:
				param.Kind = Boolean
				param.In = false
			case typeNameDateTime:
				param.Kind = DateTime
				param.In = time.Time{}
			case typeNameDouble:
				param.Kind = Double
				param.In = float64(0)
			case typeNameInteger, typeNameInteger4:
				param.Kind = Integer
				param.In = int32(0)
			case typeNameString:
				param.Kind = String
				param.In = ""

			default:
				return nil, fmt.Errorf("unknown type %s", tok.Name.Local)
			}

			cdata, err = xmlParseCData(dec, tok.Name.Local)
			if err != nil {
				return nil, err
			}
			found = true

		case xml.EndElement:
			if tok.Name.Local != elNameValue {
				return nil, fmt.Errorf("expecting </value>, got token </%s>", tok.Name.Local)
			}
			param.Kind = String
			param.In = ""
			return param, nil

		case xml.CharData:
			cdata = strings.TrimSpace(string(tok))
			if len(cdata) > 0 {
				found = true
			} else {
				token, err = dec.Token()
				if err != nil {
					return nil, fmt.Errorf("expecting CDATA, got an error %w", err)
				}
			}

		default:
			return nil, fmt.Errorf("expecting <value>, got token %+v", tok)
		}
	}

	switch param.Kind {
	case Unset, String, Base64:
		param.Kind = String
		param.In = cdata
	case Boolean:
		cdata = strings.ToLower(cdata)
		if cdata == "1" || cdata == "true" {
			param.In = true
		}
	case DateTime:
		param.In, err = time.Parse(timeLayoutISO8601, cdata)
		if err != nil {
			return nil, fmt.Errorf("invalid dateTime value %s: %w", cdata, err)
		}
	case Double:
		param.In, err = strconv.ParseFloat(cdata, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid double value %s: %w", cdata, err)
		}
	case Integer:
		i64, err := strconv.ParseInt(cdata, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value %s: %w", cdata, err)
		}
		param.In = int32(i64)
	}

	err = xmlMustEnd(dec, elNameValue)
	if err != nil {
		return nil, err
	}

	return param, nil
}

func xmlParseCData(dec *xml.Decoder, closeTag string) (cdata string, err error) {
	token, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("expecting CDATA, got error %w", err)
	}

	found := false
	for !found {
		switch tok := token.(type) {
		case xml.StartElement:
			return "", fmt.Errorf("expecting CDATA, got <%s>", tok.Name.Local)

		case xml.EndElement:
			if tok.Name.Local != closeTag {
				return "", fmt.Errorf("expecting </%s>, got </%s>", closeTag, tok.Name.Local)
			}
			return "", nil

		case xml.CharData:
			cdata = strings.TrimSpace(string(tok))
			found = true

		case xml.Comment:
			token, err = dec.Token()
			if err != nil {
				return "", fmt.Errorf("expecting CDATA, got an error %w", err)
			}
		default:
			return "", fmt.Errorf("expecting CDATA, got %T %+v", token, tok)
		}
	}

	err = xmlMustEnd(dec, closeTag)
	if err != nil {
		return cdata, err
	}

	return cdata, nil
}

func xmlParseArray(dec *xml.Decoder) (arr *Value, err error) {
	arr = &Value{}
	arr.Kind = Array

	err = xmlMustStart(dec, elNameData)
	if err != nil {
		return nil, err
	}

	for {
		v, err := xmlParseValue(dec, elNameData)
		if err != nil {
			return nil, err
		}
		if v == nil {
			break
		}
		arr.ArrayValues = append(arr.ArrayValues, v)
	}

	err = xmlMustEnd(dec, typeNameArray)
	if err != nil {
		return nil, err
	}

	return arr, nil
}

func xmlParseStruct(dec *xml.Decoder) (v *Value, err error) {
	v = &Value{}
	v.Kind = Struct

	for {
		member, err := xmlParseStructMember(dec)
		if err != nil {
			return nil, err
		}
		if member == nil {
			break
		}
		v.StructMembers = append(v.StructMembers, member)
	}

	return v, nil
}

func xmlParseStructMember(dec *xml.Decoder) (m *Member, err error) {
	isOpen, err := xmlStart(dec, elNameMember, typeNameStruct)
	if err != nil {
		return nil, err
	}
	if !isOpen {
		return nil, nil
	}

	cdata, err := xmlMustCData(dec, elNameName)
	if err != nil {
		return nil, err
	}

	m = &Member{}
	m.Name = strings.TrimSpace(cdata)

	m.Value, err = xmlParseValue(dec, elNameMember)
	if err != nil {
		return nil, err
	}

	err = xmlMustEnd(dec, elNameMember)
	if err != nil {
		return nil, err
	}

	return m, nil
}
