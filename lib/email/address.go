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
	libnet "github.com/shuLhan/share/lib/net"
)

const (
	stateBegin       = 1 << iota // 1
	stateDisplayName             // 2
	stateLocalPart               // 4
	stateDomain                  // 8
	stateEnd                     // 16
	stateGroupEnd                // 32
)

//
// Mailbox represent an invidual mailbox.
//
type Mailbox struct {
	Name    []byte
	Local   []byte
	Domain  []byte
	Address []byte
	isAngle bool
}

//
// String return the text representation of mailbox.
//
func (mbox *Mailbox) String() string {
	var sb strings.Builder

	if len(mbox.Name) > 0 {
		sb.Write(mbox.Name)
		sb.WriteByte(' ')
	}
	sb.WriteByte('<')
	sb.Write(mbox.Local)
	if len(mbox.Domain) > 0 {
		sb.WriteByte('@')
		sb.Write(mbox.Domain)
	}
	sb.WriteByte('>')

	return sb.String()
}

//
// ParseAddress parse raw address into single or multiple mailboxes.
// Raw address can be a group of address, list of mailbox, or single mailbox.
//
// A group of address have the following syntax,
//
//	DisplayName ":" mailbox-list ";" [comment]
//
// List of mailbox (mailbox-list) have following syntax,
//
//	mailbox *("," mailbox)
//
// A single mailbox have following syntax,
//
//	[DisplayName] ["<"] local "@" domain [">"]
//
// The angle bracket is optional, but both must be provided.
//
// DisplayName, local, and domain can have comment before and/or after it,
//
//	[comment] text [comment]
//
// A comment have the following syntax,
//
//	"(" text [comment] ")"
//
func ParseAddress(raw []byte) (mboxes []*Mailbox, err error) { // nolint: gocyclo
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, errors.New("ParseAddress: empty address")
	}

	r := &libio.Reader{}
	r.Init(raw)

	var (
		seps    = []byte{'(', ':', '<', '@', '>', ',', ';'}
		tok     []byte
		value   []byte
		isGroup bool
		c       byte
		mbox    *Mailbox
		state   = stateBegin
	)

	_ = r.SkipSpace()
	tok, _, c = r.ReadUntil(seps, nil)
	for {
		switch c {
		case '(':
			_, err = skipComment(r)
			if err != nil {
				return nil, err
			}
			if len(tok) > 0 {
				value = append(value, tok...)
			}

		case ':':
			if state != stateBegin {
				return nil, errors.New("ParseAddress: invalid character: ':'")
			}
			isGroup = true
			value = nil
			state = stateDisplayName
			_ = r.SkipSpace()

		case '<':
			if state >= stateLocalPart {
				return nil, errors.New("ParseAddress: invalid character: '<'")
			}
			value = append(value, tok...)
			value = bytes.TrimSpace(value)
			mbox = &Mailbox{
				isAngle: true,
			}
			if len(value) > 0 {
				mbox.Name = value
			}
			value = nil
			state = stateLocalPart

		case '@':
			if state >= stateDomain {
				return nil, errors.New("ParseAddress: invalid character: '@'")
			}
			value = append(value, tok...)
			value = bytes.TrimSpace(value)
			if len(value) == 0 {
				return nil, errors.New("ParseAddress: empty local")
			}
			if mbox == nil {
				mbox = &Mailbox{}
			}
			if !IsValidLocal(value) {
				return nil, fmt.Errorf("ParseAddress: invalid local: '%s'", value)
			}
			mbox.Local = value
			value = nil
			state = stateDomain

		case '>':
			if state > stateDomain || !mbox.isAngle {
				return nil, errors.New("ParseAddress: invalid character: '>'")
			}
			value = append(value, tok...)
			value = bytes.TrimSpace(value)
			if state == stateDomain {
				if !libnet.IsHostnameValid(value) {
					return nil, fmt.Errorf("ParseAddress: invalid domain: '%s'", value)
				}
			}
			mbox.Domain = value
			mboxes = append(mboxes, mbox)
			mbox = nil
			value = nil
			state = stateEnd

		case ';':
			if state < stateDomain || !isGroup {
				return nil, errors.New("ParseAddress: invalid character: ';'")
			}
			if mbox != nil && mbox.isAngle {
				return nil, errors.New("ParseAddress: missing '>'")
			}
			value = append(value, tok...)
			value = bytes.TrimSpace(value)
			switch state {
			case stateDomain:
				if !libnet.IsHostnameValid(value) {
					return nil, fmt.Errorf("ParseAddress: invalid domain: '%s'", value)
				}
				mbox.Domain = value
				mboxes = append(mboxes, mbox)
				mbox = nil
			case stateEnd:
				if len(value) > 0 {
					return nil, fmt.Errorf("ParseAddress: invalid token: '%s'", value)
				}
			}
			isGroup = false
			value = nil
			state = stateGroupEnd
		case ',':
			if state < stateDomain {
				return nil, errors.New("ParseAddress: invalid character: ','")
			}
			if mbox != nil && mbox.isAngle {
				return nil, errors.New("ParseAddress: missing '>'")
			}
			value = append(value, tok...)
			value = bytes.TrimSpace(value)
			switch state {
			case stateDomain:
				if !libnet.IsHostnameValid(value) {
					return nil, fmt.Errorf("ParseAddress: invalid domain: '%s'", value)
				}
				mbox.Domain = value
				mboxes = append(mboxes, mbox)
				mbox = nil
			case stateEnd:
				if len(value) > 0 {
					return nil, fmt.Errorf("ParseAddress: invalid token: '%s'", value)
				}
			}
			value = nil
			state = stateBegin
		case 0:
			if state < stateDomain {
				return nil, errors.New("ParseAddress: empty or invalid address")
			}
			if state != stateEnd && mbox != nil && mbox.isAngle {
				return nil, errors.New("ParseAddress: missing '>'")
			}
			if isGroup {
				return nil, errors.New("ParseAddress: missing ';'")
			}

			value = append(value, tok...)
			value = bytes.TrimSpace(value)
			if state == stateGroupEnd {
				if len(value) > 0 {
					return nil, fmt.Errorf("ParseAddress: trailing text: '%s'", value)
				}
			}

			if state == stateDomain {
				if !libnet.IsHostnameValid(value) {
					return nil, fmt.Errorf("ParseAddress: invalid domain: '%s'", value)
				}
				mbox.Domain = value
				mboxes = append(mboxes, mbox)
				mbox = nil
			}
			goto out
		}
		tok, _, c = r.ReadUntil(seps, nil)
	}
out:
	return mboxes, nil
}

//
// skipComment skip all characters inside parentheses, '(' and ')'.
//
// A comment can contains quoted-pair, which means opening or closing
// parentheses can be escaped using backslash character '\', for example
// "( a \) comment)".
//
// A comment can be nested, for example "(a (comment))"
//
func skipComment(r *libio.Reader) (c byte, err error) {
	seps := []byte{'\\', '(', ')'}
	c = r.SkipUntil(seps)
	for {
		switch c {
		case 0:
			return c, errors.New("ParseAddress: missing comment close parentheses")
		case '\\':
			// We found backslash, skip one character and continue
			// looking for separator.
			r.SkipN(1)
		case '(':
			c, err = skipComment(r)
			if err != nil {
				return c, err
			}
		case ')':
			c = r.SkipSpace()
			if c != '(' {
				goto out
			}
			r.SkipN(1)
		}
		c = r.SkipUntil(seps)
	}
out:
	return c, nil
}
