// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"fmt"
	"strings"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

var (
	topText  = []byte("text")
	subPlain = []byte("plain")
	subHtml  = []byte("html")
)

// ContentType represent MIME header "Content-Type" field.
type ContentType struct {
	Top    []byte
	Sub    []byte
	Params []Param
}

// ParseContentType parse content type from raw bytes.
func ParseContentType(raw []byte) (ct *ContentType, err error) {
	raw = bytes.TrimSpace(raw)

	ct = &ContentType{}
	if len(raw) == 0 {
		ct.Top = []byte("text")
		ct.Sub = []byte("plain")
		ct.Params = []Param{{
			Key:   []byte("charset"),
			Value: []byte("us-ascii"),
		}}
		return ct, nil
	}

	var (
		logp   = `ParseContentType`
		parser = libbytes.NewParser(raw, []byte{'/', ';'})
		c      byte
	)

	ct.Top, c = parser.Read()
	if c != '/' {
		return nil, fmt.Errorf(`%s: missing subtype`, logp)
	}
	if !isValidToken(ct.Top, false) {
		return nil, fmt.Errorf(`%s: invalid type '%s'`, logp, ct.Top)
	}

	ct.Sub, c = parser.Read()
	if !isValidToken(ct.Sub, false) {
		return nil, fmt.Errorf(`%s: invalid subtype '%s'`, logp, ct.Sub)
	}
	if c == 0 {
		return ct, nil
	}
	if c != ';' {
		return nil, fmt.Errorf(`%s: invalid character '%c'`, logp, c)
	}

	parser.SetDelimiters([]byte{'=', '"', ';'})
	for c == ';' {
		param := Param{}

		param.Key, c = parser.ReadNoSpace()
		if c == 0 {
			// Ignore key without value.
			break
		}

		if !isValidToken(param.Key, false) {
			return nil, fmt.Errorf(`%s: invalid parameter key '%s'`, logp, param.Key)
		}

		if c != '=' {
			return nil, fmt.Errorf(`%s: expecting '=', got '%c'`, logp, c)
		}

		param.Value, c = parser.ReadNoSpace()
		if c == '"' {
			if len(param.Value) != 0 {
				return nil, fmt.Errorf(`%s: invalid parameter value '%s'`, logp, param.Value)
			}

			// The param value may contain '=' or ';', remove it
			// temporarily.
			parser.RemoveDelimiters([]byte{'=', ';'})

			param.Value, c = parser.Read()
			if c != '"' {
				return nil, fmt.Errorf(`%s: missing closing quote`, logp)
			}
			param.Quoted = true

			parser.AddDelimiters([]byte{'=', ';'})

			c = parser.Skip()
		}
		if !isValidToken(param.Value, param.Quoted) {
			return nil, fmt.Errorf(`%s: invalid parameter value '%s'`, logp, param.Value)
		}

		ct.Params = append(ct.Params, param)
	}

	return ct, nil
}

func isValidToken(tok []byte, quoted bool) bool {
	if len(tok) == 0 {
		return false
	}
	for x := 0; x < len(tok); x++ {
		if quoted && tok[x] == ' ' {
			continue
		}
		if tok[x] < 33 {
			return false
		}
		if quoted {
			continue
		}
		switch tok[x] {
		case '(', ')', '<', '>', '@', ',', ';', ':', '\\', '"', '/',
			'[', ']', '?', '=':
			return false
		}
	}
	return true
}

// GetParamValue return parameter value related to specific name.
func (ct *ContentType) GetParamValue(name []byte) []byte {
	for _, p := range ct.Params {
		if bytes.EqualFold(p.Key, name) {
			return p.Value
		}
	}
	return nil
}

// isEqual will return true if the Top and Sub matched with other, in
// case-insensitive matter.
func (ct *ContentType) isEqual(other *ContentType) bool {
	if other == nil {
		return false
	}
	if !bytes.EqualFold(ct.Top, other.Top) {
		return false
	}
	return bytes.EqualFold(ct.Sub, other.Sub)
}

// SetBoundary set or replace the Value for Key "boundary".
func (ct *ContentType) SetBoundary(boundary []byte) {
	for x := 0; x < len(ct.Params); x++ {
		if bytes.EqualFold(ct.Params[x].Key, ParamNameBoundary) {
			ct.Params[x].Value = boundary
			return
		}
	}
	paramBoundary := Param{
		Key:   ParamNameBoundary,
		Value: boundary,
	}
	ct.Params = append(ct.Params, paramBoundary)
}

// String return text representation of content type with its parameters.
func (ct *ContentType) String() string {
	var sb strings.Builder

	sb.Write(ct.Top)
	sb.WriteByte('/')
	sb.Write(ct.Sub)
	for _, p := range ct.Params {
		sb.WriteByte(';')
		sb.WriteByte(' ')
		sb.Write(p.Key)
		sb.WriteByte('=')
		if p.Quoted {
			sb.WriteByte('"')
		}
		sb.Write(p.Value)
		if p.Quoted {
			sb.WriteByte('"')
		}
	}

	return sb.String()
}
