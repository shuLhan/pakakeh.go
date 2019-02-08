// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"fmt"

	libbytes "github.com/shuLhan/share/lib/bytes"
	libio "github.com/shuLhan/share/lib/io"
)

type empty struct{}

//
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
//
type parser struct {
	tags   map[tagKey]empty // Map to check duplicate tags.
	sepKey []byte
	sepVal []byte

	r      *libio.Reader
	c      byte
	isTerm bool
	tok    []byte
}

//
// newParser create and initialize new parser for DKIM Signature.
//
func newParser(value []byte) (p *parser) {
	p = &parser{
		r:      &libio.Reader{},
		sepKey: []byte{'='},
		sepVal: []byte{';'},
		tags:   make(map[tagKey]empty),
	}
	p.r.Init(value)

	return p
}

//
// fetchTag parse and return single tag from reader.
//
func (p *parser) fetchTag() (t *tag, err error) {
	p.c = p.r.SkipSpaces()
	if p.c == 0 {
		return nil, nil
	}

	t, err = p.fetchTagKey()
	if err != nil || p.c == 0 {
		return t, err
	}

	err = p.fetchTagValue(t)

	return t, err
}

//
// fetchTagKey parse and fetch tag's key.
//
func (p *parser) fetchTagKey() (t *tag, err error) {
	p.tok, p.isTerm, p.c = p.r.ReadUntil(p.sepKey, libbytes.ASCIISpaces)

	t, err = newTag(p.tok)
	if err != nil || t == nil {
		return nil, err
	}

	if p.isTerm || p.c == 0 {
		p.c = p.r.SkipSpaces()
		if p.c != '=' {
			return nil, fmt.Errorf("dkim: missing '=': '%s'", p.r.Rest())
		}
		p.r.SkipN(1)
	}
	if t.key != tagUnknown {
		_, ok := p.tags[t.key]
		if ok {
			return nil, fmt.Errorf("dkim: duplicate tag: '%s'", p.tok)
		}
		p.tags[t.key] = empty{}
	}

	p.c = p.r.SkipSpaces()

	return t, nil
}

//
// fetchTagValue parse and fetch tag's value.
//
func (p *parser) fetchTagValue(t *tag) (err error) {
	var v []byte
	sepCR := []byte{'\r'}
	for {
		p.tok, p.isTerm, p.c = p.r.ReadUntil(sepCR, p.sepVal)
		v = append(v, p.tok...)
		if p.isTerm || p.c == 0 {
			break
		}
		p.c = p.r.SkipSpaces()
	}
	err = t.setValue(v)

	return err
}
