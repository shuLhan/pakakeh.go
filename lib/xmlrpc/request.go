// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// Request represent the XML-RPC request, including method name and optional
// parameters.
type Request struct {
	MethodName string
	Params     []*Value
}

// NewRequest create and initialize new request.
func NewRequest(methodName string, params []interface{}) (req Request, err error) {
	req = Request{
		MethodName: methodName,
		Params:     make([]*Value, 0, len(params)),
	}

	for _, p := range params {
		v := NewValue(p)
		if v == nil {
			return req, fmt.Errorf("NewRequest: cannot convert parameter %v", p)
		}

		req.Params = append(req.Params, v)
	}

	return req, nil
}

// MarshalText implement the encoding.TextMarshaler interface.
func (req Request) MarshalText() (out []byte, err error) {
	var buf bytes.Buffer

	buf.WriteString(xml.Header)
	buf.WriteString("<methodCall>")
	buf.WriteString("<methodName>" + req.MethodName + "</methodName>")
	if len(req.Params) > 0 {
		buf.WriteString("<params>")
	}

	for _, p := range req.Params {
		fmt.Fprintf(&buf, "<param>%s</param>", p.String())
	}

	if len(req.Params) > 0 {
		buf.WriteString("</params>")
	}
	buf.WriteString("</methodCall>")

	return buf.Bytes(), nil
}

// UnmarshalText parse the XML request.
func (req *Request) UnmarshalText(text []byte) (err error) {
	var (
		logp      = "xmlrpc: Request"
		dec       = xml.NewDecoder(bytes.NewReader(text))
		hasParams bool
	)

	err = xmlBegin(dec)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = xmlMustStart(dec, elNameMethodCall)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	req.MethodName, err = xmlMustCData(dec, elNameMethodName)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	req.Params, hasParams, err = xmlParseParams(dec, elNameMethodCall)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	if hasParams {
		err = xmlMustEnd(dec, elNameMethodCall)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}

	return nil
}
