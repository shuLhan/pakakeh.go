// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type Response struct {
	Param        Value
	FaultMessage string
	FaultCode    int32
	IsFault      bool
}

func (resp *Response) UnmarshalText(text []byte) (err error) {
	dec := xml.NewDecoder(bytes.NewReader(text))

	err = xmlBegin(dec)
	if err != nil {
		return fmt.Errorf("UnmarshalText: %w", err)
	}

	el, _, err := xmlNext(dec)
	if err != nil {
		return fmt.Errorf("UnmarshalText: %w", err)
	}
	if el.Name.Local != elNameMethodResponse {
		return fmt.Errorf("UnmarshalText: expecting '<%s>' got %v",
			elNameMethodResponse, el.Name.Local)
	}

	el, _, err = xmlNext(dec)
	if err != nil {
		return fmt.Errorf("UnmarshalText: %w", err)
	}

	switch el.Name.Local {
	case elNameFault:
		err = resp.unmarshalFault(dec)
		if err != nil {
			return fmt.Errorf("UnmarshalText: %w", err)
		}

	case elNameParams:
		el, _, err = xmlNext(dec)
		if err != nil {
			return fmt.Errorf("UnmarshalText: %w", err)
		}
		if el.Name.Local != elNameParam {
			return fmt.Errorf("UnmarshalText: expecting '<%s>' got %v",
				elNameParam, el)
		}

		el, _, err = xmlNext(dec)
		if err != nil {
			return fmt.Errorf("UnmarshalText: %w", err)
		}
		if el.Name.Local != elNameValue {
			return fmt.Errorf("UnmarshalText: expecting '<%s>' got %v",
				elNameValue, el)
		}

		resp.Param, err = xmlParseScalarValue(dec)
		if err != nil {
			return fmt.Errorf("UnmarshalText: %w", err)
		}
	default:
		return fmt.Errorf("UnmarshalText: expecting '<params>' or '<fault>' got %v",
			el.Name.Local)
	}

	return nil
}

//
// unmarshalFault parse the XML fault error code and message.
//
func (resp *Response) unmarshalFault(dec *xml.Decoder) (err error) {
	resp.IsFault = true

	el, _, err := xmlNext(dec)
	if err != nil {
		return fmt.Errorf("unmarshalFault: %w", err)
	}
	if el.Name.Local != elNameValue {
		return fmt.Errorf("expecting '<%s>' got %v", elNameValue,
			el.Name.Local)
	}

	el, _, err = xmlNext(dec)
	if err != nil {
		return fmt.Errorf("unmarshalFault: %w", err)
	}
	if el.Name.Local != typeNameStruct {
		return fmt.Errorf("unmarshalFault: expecting '<%s>' got %v",
			typeNameStruct, el.Name.Local)
	}

	v, err := xmlParseStruct(dec)
	if err != nil {
		return fmt.Errorf("unmarshalFault: %w", err)
	}

	resp.FaultCode = v.GetFieldAsInteger(memberNameFaultCode)
	resp.FaultMessage = v.GetFieldAsString(memberNameFaultString)

	return nil
}
