// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"errors"
	"fmt"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

type empty struct{}

// parser for DKIM tags.
//
// Rules,
//
// *  Tags MUST NOT duplicate, otherwise entire value is invalid.
// *  Unrecognized tags MUST be ignored
// *  Tag with an empty value explicitly designates the empty string as the
// value.
//
// Simplified syntax,
//
//	tags  = tag *( ";" tag )
//	tag   = [FWS] key [FWS] "=" [FWS] value [FWS]
//	key   = ALPHA *(ALPHA / DIGIT / "_")
//	value = tval *( tval / WSP / FWS )
//
//	tval  = %x21-3A / %x3C-7E
//	FWS   = *(*(WSP) "\r\n")
//	WSP   = " " / "\t"
type parser struct {
	tags   map[tagKey]empty // Map to check duplicate tags.
	parser *libbytes.Parser
}

// newParser create and initialize new parser for DKIM Signature.
func newParser(value []byte) (p *parser) {
	p = &parser{
		parser: libbytes.NewParser(value, nil),
		tags:   make(map[tagKey]empty),
	}

	return p
}

// fetchTag parse and return single tag from reader.
func (p *parser) fetchTag() (t *tag, err error) {
	var (
		token []byte
		d     byte
	)

	p.parser.SetDelimiters([]byte{'='})

	token, d = p.parser.ReadNoSpace()
	if d == 0 {
		return nil, nil
	}
	if d != '=' {
		return nil, errors.New(`dkim: missing '='`)
	}

	t, err = newTag(token)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}
	if t.key != tagUnknown {
		var ok bool
		_, ok = p.tags[t.key]
		if ok {
			return nil, fmt.Errorf(`dkim: duplicate tag: '%s'`, token)
		}
		p.tags[t.key] = empty{}
	}

	p.parser.SetDelimiters([]byte{';'})

	token, _ = p.parser.ReadNoSpace()

	err = t.setValue(token)
	if err != nil {
		return nil, err
	}

	return t, nil
}
