// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	libbytes "github.com/shuLhan/share/lib/bytes"
	libjson "github.com/shuLhan/share/lib/json"
	libnet "github.com/shuLhan/share/lib/net"
)

// Mailbox represent an invidual mailbox.
type Mailbox struct {
	Address string // address contains the combination of "local@domain"
	Name    []byte
	Local   []byte
	Domain  []byte
	isAngle bool
}

// String return the text representation of mailbox.
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

// ParseMailbox parse the raw address(es) and return the first mailbox in the
// list.
// If the raw parameter is empty or no mailbox present or mailbox format is
// invalid, it will return nil.
func ParseMailbox(raw []byte) (mbox *Mailbox) {
	mboxes, err := ParseMailboxes(raw)
	if err != nil {
		return nil
	}
	if len(mboxes) > 0 {
		mbox = mboxes[0]
	}
	return mbox
}

// ParseMailboxes parse raw [address] into single or multiple mailboxes.
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
// [address]: https://www.rfc-editor.org/rfc/rfc5322.html#section-3.4
func ParseMailboxes(raw []byte) (mboxes []*Mailbox, err error) {
	var logp = `ParseMailboxes`

	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, fmt.Errorf(`%s: empty address`, logp)
	}

	var (
		delims = []byte{'(', ':', '<', '@', '>', ',', ';'}
		parser = libbytes.NewParser(raw, delims)

		token   []byte
		isGroup bool
		c       byte
		mbox    *Mailbox
	)

	token, c, err = parseMailboxText(parser)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	if c == 0 || c == '>' || c == ',' || c == ';' {
		return nil, fmt.Errorf(`%s: empty or invalid address`, logp)
	}

	token = bytes.TrimSpace(token)
	mbox = &Mailbox{}
	if c == ':' {
		// We are parsing group of mailbox.
		isGroup = true
	} else if c == '<' {
		mbox.isAngle = true
		mbox.Name = token
	} else if c == '@' {
		if len(token) == 0 {
			return nil, fmt.Errorf(`%s: empty local`, logp)
		}
		if !IsValidLocal(token) {
			return nil, fmt.Errorf(`%s: invalid local '%s'`, logp, token)
		}
		mbox.Local = token
	}

	c, err = parseMailbox(mbox, parser, c)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	mboxes = append(mboxes, mbox)

	for c == ',' {
		mbox = &Mailbox{}
		c, err = parseMailbox(mbox, parser, c)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		mboxes = append(mboxes, mbox)
	}

	if isGroup {
		if c != ';' {
			return nil, fmt.Errorf(`%s: missing group terminator ';'`, logp)
		}
	} else {
		if c != 0 {
			return nil, fmt.Errorf(`%s: invalid character '%c'`, logp, c)
		}
	}

	return mboxes, nil
}

// parseMailbox continue parsing single mailbox based on previous delimiter
// prevd.
// On success it will return the last delimiter that is not for mailbox,
// either ',' for list of mailbox or ';' for end of group.
func parseMailbox(mbox *Mailbox, parser *libbytes.Parser, prevd byte) (c byte, err error) {
	var (
		logp = `parseMailbox`

		value []byte
	)

	c = prevd
	if c == ':' || c == ',' {
		// Get the name or local part.
		value, c, err = parseMailboxText(parser)
		if err != nil {
			return c, fmt.Errorf(`%s: %w`, logp, err)
		}

		if c == '<' {
			mbox.isAngle = true
			mbox.Name = value
		} else if c == '@' {
			if len(value) == 0 {
				return c, fmt.Errorf(`%s: empty local`, logp)
			}
			if !IsValidLocal(value) {
				return c, fmt.Errorf(`%s: invalid local '%s'`, logp, value)
			}
			mbox.Local = value
		}
	}
	if c == '<' {
		// Get the local or domain part.
		value, c, err = parseMailboxText(parser)
		if err != nil {
			return c, fmt.Errorf(`%s: %w`, logp, err)
		}
		if c == '>' {
			// No local part, only domain part, "<domain>".
			goto domain
		}
		if c != '@' {
			return c, fmt.Errorf(`%s: invalid character '%c'`, logp, c)
		}
		if !IsValidLocal(value) {
			return c, fmt.Errorf(`%s: invalid local '%s'`, logp, value)
		}
		mbox.Local = value
	}

	// Get the domain part.
	value, c, err = parseMailboxText(parser)
	if err != nil {
		return c, fmt.Errorf(`%s: %w`, logp, err)
	}
domain:
	if len(value) != 0 {
		if !libnet.IsHostnameValid(value, false) {
			return c, fmt.Errorf(`%s: invalid domain '%s'`, logp, value)
		}
	}
	if len(mbox.Local) != 0 && len(value) == 0 {
		return c, fmt.Errorf(`%s: invalid domain '%s'`, logp, value)
	}
	mbox.Domain = value
	mbox.Address = fmt.Sprintf(`%s@%s`, mbox.Local, mbox.Domain)

	if mbox.isAngle {
		if c != '>' {
			return c, fmt.Errorf(`%s: missing '>'`, logp)
		}
		value, c, err = parseMailboxText(parser)
		if err != nil {
			return c, fmt.Errorf(`%s: %w`, logp, err)
		}
		if len(value) != 0 {
			return c, fmt.Errorf(`%s: unknown token '%s'`, logp, value)
		}
	}

	if c == ',' || c == ';' || c == 0 {
		return c, nil
	}

	return c, fmt.Errorf(`%s: invalid character '%c'`, logp, c)
}

// parseMailboxText parse text (display-name, local-part, or domain) probably
// with comment inside.
func parseMailboxText(parser *libbytes.Parser) (text []byte, c byte, err error) {
	var (
		logp   = `parseMailboxText`
		delims = []byte{'\\', ')'}

		token []byte
	)

	parser.AddDelimiters(delims)
	defer parser.RemoveDelimiters(delims)

	token, c = parser.ReadNoSpace()
	for {
		text = append(text, token...)

		if c == ')' {
			return nil, c, fmt.Errorf(`%s: invalid character '%c'`, logp, c)
		}
		if c == '\\' {
			token, _ = parser.ReadN(1)
			text = append(text, token...)
			token, c = parser.ReadNoSpace()
			continue
		}
		if c == '(' {
			err = skipComment(parser)
			if err != nil {
				return nil, 0, fmt.Errorf(`%s: %w`, logp, err)
			}
			token, c = parser.ReadNoSpace()
			continue
		}
		break
	}
	text = bytes.TrimSpace(text)
	return text, c, nil
}

// skipComment skip all characters inside parentheses, '(' and ')'.
//
// A comment can contains quoted-pair, which means opening or closing
// parentheses can be escaped using backslash character '\', for example
// "( a \) comment)".
//
// A comment can be nested, for example "(a (comment))"
func skipComment(parser *libbytes.Parser) (err error) {
	var c = parser.Skip()
	for {
		if c == 0 {
			return errors.New(`missing comment close parentheses`)
		}
		if c == ')' {
			break
		}
		if c == '(' {
			err = skipComment(parser)
			if err != nil {
				return err
			}
			c = parser.Skip()
			continue
		}
		// c == '\\'
		// We found backslash, skip one character and continue
		// looking for next delimiter.
		parser.SkipN(1)
		c = parser.Skip()
	}
	return nil
}

func (mbox *Mailbox) UnmarshalJSON(b []byte) (err error) {
	// Replace \u003c and \u003e escaped characters back to '<' and '>'.
	b, err = libjson.Unescape(b, false)
	if err != nil {
		return err
	}
	if b[0] == '"' {
		b = b[1:]
	}
	if b[len(b)-1] == '"' {
		b = b[:len(b)-1]
	}
	got := ParseMailbox(b)
	if got == nil {
		return nil
	}
	*mbox = *got
	return nil
}

func (mbox *Mailbox) MarshalJSON() (b []byte, err error) {
	return []byte(`"` + mbox.String() + `"`), nil
}
