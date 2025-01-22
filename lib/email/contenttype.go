// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package email

import (
	"bytes"
	"fmt"
	"strings"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

var (
	topText  = `text`
	subPlain = `plain`
	subHTML  = `html`
)

// ContentType represent MIME header "Content-Type" field.
type ContentType struct {
	Top    string
	Sub    string
	Params []Param
}

// ParseContentType parse content type from raw bytes.
func ParseContentType(raw []byte) (ct *ContentType, err error) {
	raw = bytes.TrimSpace(raw)

	ct = &ContentType{}
	if len(raw) == 0 {
		ct.Top = `text`
		ct.Sub = `plain`
		ct.Params = []Param{{
			Key:   `charset`,
			Value: `us-ascii`,
		}}
		return ct, nil
	}

	var (
		logp   = `ParseContentType`
		parser = libbytes.NewParser(raw, []byte{'/', ';'})
		tok    []byte
		c      byte
	)

	tok, c = parser.Read()
	if c != '/' {
		return nil, fmt.Errorf(`%s: missing subtype`, logp)
	}
	if !isValidToken(tok, false) {
		return nil, fmt.Errorf(`%s: invalid type '%s'`, logp, tok)
	}
	ct.Top = string(tok)

	tok, c = parser.Read()
	if !isValidToken(tok, false) {
		return nil, fmt.Errorf(`%s: invalid subtype '%s'`, logp, tok)
	}
	ct.Sub = string(tok)
	if c == 0 {
		return ct, nil
	}
	if c != ';' {
		return nil, fmt.Errorf(`%s: invalid character '%c'`, logp, c)
	}

	parser.SetDelimiters([]byte{'=', '"', ';'})
	for c == ';' {
		param := Param{}

		tok, c = parser.ReadNoSpace()
		if c == 0 {
			// Ignore key without value.
			param.Key = string(tok)
			break
		}

		if !isValidToken(tok, false) {
			return nil, fmt.Errorf(`%s: invalid parameter key '%s'`, logp, tok)
		}
		if c != '=' {
			return nil, fmt.Errorf(`%s: expecting '=', got '%c'`, logp, c)
		}
		param.Key = string(tok)

		tok, c = parser.ReadNoSpace()
		if c == '"' {
			if len(tok) != 0 {
				return nil, fmt.Errorf(`%s: invalid parameter value '%s'`, logp, tok)
			}

			// The param value may contain '=' or ';', remove it
			// temporarily.
			parser.RemoveDelimiters([]byte{'=', ';'})

			tok, c = parser.Read()
			if c != '"' {
				return nil, fmt.Errorf(`%s: missing closing quote`, logp)
			}
			param.Quoted = true

			parser.AddDelimiters([]byte{'=', ';'})

			c = parser.Skip()
		}
		if !isValidToken(tok, param.Quoted) {
			return nil, fmt.Errorf(`%s: invalid parameter value '%s'`, logp, tok)
		}
		param.Value = string(tok)
		ct.Params = append(ct.Params, param)
	}

	return ct, nil
}

func isValidToken(tok []byte, quoted bool) bool {
	if len(tok) == 0 {
		return false
	}
	for x := range len(tok) {
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
func (ct *ContentType) GetParamValue(name string) string {
	for _, p := range ct.Params {
		if strings.EqualFold(p.Key, name) {
			return p.Value
		}
	}
	return ``
}

// isEqual will return true if the Top and Sub matched with other, in
// case-insensitive matter.
func (ct *ContentType) isEqual(other *ContentType) bool {
	if other == nil {
		return false
	}
	if !strings.EqualFold(ct.Top, other.Top) {
		return false
	}
	return strings.EqualFold(ct.Sub, other.Sub)
}

// SetBoundary set or replace the Value for Key "boundary".
func (ct *ContentType) SetBoundary(boundary string) {
	for x := range len(ct.Params) {
		if strings.EqualFold(ct.Params[x].Key, ParamNameBoundary) {
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

	sb.WriteString(ct.Top)
	sb.WriteByte('/')
	sb.WriteString(ct.Sub)
	for _, p := range ct.Params {
		sb.WriteByte(';')
		sb.WriteByte(' ')
		sb.WriteString(p.Key)
		sb.WriteByte('=')
		if p.Quoted {
			sb.WriteByte('"')
		}
		sb.WriteString(p.Value)
		if p.Quoted {
			sb.WriteByte('"')
		}
	}

	return sb.String()
}
