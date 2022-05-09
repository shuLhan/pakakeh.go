// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"

	liberrors "github.com/shuLhan/share/lib/errors"
)

type Response struct {
	liberrors.E
	Param *Value
}

// MarshalText encode the Response instance into XML text.
func (resp *Response) MarshalText() (out []byte, err error) {
	var buf bytes.Buffer

	buf.WriteString(xml.Header)
	buf.WriteString("<methodResponse>")

	if resp.Code <= 200 {
		fmt.Fprintf(&buf, "<params><param>%s</param></params>",
			resp.Param)
	} else {
		buf.WriteString("<fault><value><struct>")
		fmt.Fprintf(&buf, "<member><name>faultCode</name><value><int>%d</int></value></member>",
			resp.Code)
		fmt.Fprintf(&buf, "<member><name>faultString</name><value><string>%s</string></value></member>",
			resp.Message)
		buf.WriteString("</struct></value></fault>")
	}

	buf.WriteString("</methodResponse>")

	return buf.Bytes(), nil
}

func (resp *Response) UnmarshalText(text []byte) (err error) {
	var (
		logp = "xmlrpc: Response"
		dec  = xml.NewDecoder(bytes.NewReader(text))
	)

	err = xmlBegin(dec)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = xmlMustStart(dec, elNameMethodResponse)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	found := false
	for !found {
		switch tok := token.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case elNameFault:
				err = resp.unmarshalFault(dec)
				if err != nil {
					return fmt.Errorf("%s: %w", logp, err)
				}
				found = true

			case elNameParams:
				resp.Param, err = xmlParseParam(dec, elNameParams)
				if err != nil {
					return fmt.Errorf("%s: %w", logp, err)
				}
				found = true

			default:
				return fmt.Errorf("%s: expecting <params> or <fault> got <%s>",
					logp, tok.Name.Local)
			}

		case xml.Comment, xml.CharData:
			token, err = dec.Token()
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}

		default:
			return fmt.Errorf("%s: expecting <params> or <fault>, got token %T %+v",
				logp, token, tok)
		}
	}

	return nil
}

// Unwrap return the error as instance of *liberror.E.
func (resp *Response) Unwrap() (err error) {
	return &resp.E
}

// unmarshalFault parse the XML fault error code and message.
func (resp *Response) unmarshalFault(dec *xml.Decoder) (err error) {
	v, err := xmlParseValue(dec, elNameFault)
	if err != nil {
		return fmt.Errorf("unmarshalFault: %w", err)
	}

	resp.Code = v.GetFieldAsInteger(memberNameFaultCode)
	resp.Message = v.GetFieldAsString(memberNameFaultString)

	return nil
}
