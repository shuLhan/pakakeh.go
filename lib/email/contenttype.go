// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	libio "github.com/shuLhan/share/lib/io"
)

//
// ContentType represent MIME header "Content-Type" field.
//
type ContentType struct {
	Top    []byte
	Sub    []byte
	Params []Param
}

//
// ParseContentType parse content type from raw bytes.
//
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

	r := &libio.Reader{}
	r.Init(raw)
	var c byte

	ct.Top, _, c = r.ReadUntil([]byte{'/'}, nil)
	if c == 0 {
		return nil, errors.New("ParseContentType: missing subtype")
	}
	if !isValidToken(ct.Top, false) {
		return nil, fmt.Errorf("ParseContentType: invalid type: '%s'", ct.Top)
	}

	ct.Sub, _, c = r.ReadUntil([]byte{';'}, nil)
	if !isValidToken(ct.Sub, false) {
		return nil, fmt.Errorf("ParseContentType: invalid subtype: '%s'", ct.Sub)
	}
	if c == 0 {
		return ct, nil
	}

	c = r.SkipSpace()
	ksep := []byte{'='}
	qsep := []byte{'"'}
	vsep := []byte{' '}
	for c != 0 {
		param := Param{}

		param.Key, _, c = r.ReadUntil(ksep, nil)
		if c == 0 {
			// Ignore key without value
			break
		}
		if !isValidToken(param.Key, false) {
			err = fmt.Errorf("ParseContentType: invalid parameter key: '%s'", param.Key)
			return nil, err
		}

		c = r.Current()
		if c == '"' {
			r.SkipN(1)
			param.Value, _, c = r.ReadUntil(qsep, nil)
			if c != '"' {
				return nil, errors.New("ParseContentType: missing closing quote")
			}
			param.Quoted = true
		} else {
			param.Value, _, _ = r.ReadUntil(vsep, nil)
		}
		if !isValidToken(param.Value, param.Quoted) {
			err = fmt.Errorf("ParseContentType: invalid parameter value: '%s'", param.Value)
			return nil, err
		}

		param.Key = bytes.ToLower(param.Key)
		ct.Params = append(ct.Params, param)

		c = r.SkipSpace()
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

//
// GetParamValue return parameter value related to specific name.
//
func (ct *ContentType) GetParamValue(name []byte) []byte {
	name = bytes.ToLower(name)
	for _, p := range ct.Params {
		if bytes.Equal(p.Key, name) {
			return p.Value
		}
	}
	return nil
}

//
// String return text representation of this instance.
//
func (ct *ContentType) String() string {
	var sb strings.Builder

	sb.Write(ct.Top)
	sb.WriteByte('/')
	sb.Write(ct.Sub)
	sb.WriteByte(';')
	for _, p := range ct.Params {
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
