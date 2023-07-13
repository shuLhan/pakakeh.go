// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"encoding/json"
	"fmt"
	"testing"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/test"
)

func TestParseMailbox(t *testing.T) {
	type testCase struct {
		exp *Mailbox
		in  string
	}

	var cases = []testCase{{
		in:  `(empty)`,
		exp: nil,
	}, {
		in: `one@example`,
		exp: &Mailbox{
			Local:   `one`,
			Domain:  `example`,
			Address: `one@example`,
		},
	}, {
		in: `one@example , two@example`,
		exp: &Mailbox{
			Local:   `one`,
			Domain:  `example`,
			Address: `one@example`,
		},
	}}

	var (
		c   testCase
		got *Mailbox
	)
	for _, c = range cases {
		got = ParseMailbox([]byte(c.in))
		test.Assert(t, c.in, c.exp, got)
	}
}

func TestParseMailboxes(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
		exp    string
	}
	var cases = []testCase{{
		desc:   `With empty input`,
		expErr: `ParseMailboxes: empty address`,
	}, {
		desc:   `With comment only`,
		in:     `(comment)`,
		expErr: `ParseMailboxes: empty or invalid address`,
	}, {
		desc:   `With no domain`,
		in:     `(comment)local(comment)`,
		expErr: `ParseMailboxes: empty or invalid address`,
	}, {
		desc:   `With no opening comment`,
		in:     `comment)local@domain`,
		expErr: `ParseMailboxes: parseMailboxText: invalid character ')'`,
	}, {
		desc:   `With no closing comment`,
		in:     `(commentlocal@domain`,
		expErr: `ParseMailboxes: parseMailboxText: missing comment close parentheses`,
	}, {
		desc:   `With no opening bracket`,
		in:     `(comment)local(comment)@domain>`,
		expErr: `ParseMailboxes: parseMailbox: invalid character '>'`,
	}, {
		desc:   `With no closing bracket`,
		in:     `<(comment)local(comment)@domain`,
		expErr: `ParseMailboxes: parseMailbox: missing '>'`,
	}, {
		desc:   `With ':' inside mailbox`,
		in:     `<local:part@domain>`,
		expErr: `ParseMailboxes: parseMailbox: invalid character ':'`,
	}, {
		desc: `With '<' inside local part`,
		in:   `local<part@domain>`,
		exp:  `[local <part@domain>]`,
	}, {
		desc:   `With multiple '<'`,
		in:     `Name <local<part@domain>`,
		expErr: `ParseMailboxes: parseMailbox: invalid character '<'`,
	}, {
		desc:   `With multiple '@'`,
		in:     `Name <local@part@domain>`,
		expErr: `ParseMailboxes: parseMailbox: missing '>'`,
	}, {
		desc: `With no local-part`,
		in:   `Name <local>`,
		exp:  `[Name <@local>]`,
	}, {
		desc:   `With empty local`,
		in:     `@domain`,
		expErr: `ParseMailboxes: empty local`,
	}, {
		desc:   `With empty local`,
		in:     `Name <@domain>`,
		expErr: `ParseMailboxes: parseMailbox: invalid local ''`,
	}, {
		desc:   `With invalid local`,
		in:     `e[ample@domain`,
		expErr: `ParseMailboxes: invalid local 'e[ample'`,
	}, {
		desc:   `With empty domain`,
		in:     `Name <local@>, test@domain`,
		expErr: `ParseMailboxes: parseMailbox: invalid domain ''`,
	}, {
		desc:   `With invalid domain`,
		in:     `Name <local@dom[ain>, test@domain`,
		expErr: `ParseMailboxes: parseMailbox: invalid domain 'dom[ain'`,
	}, {
		desc: `With no bracket, single address`,
		in:   `local@domain`,
		exp:  `[<local@domain>]`,
	}, {
		desc: `With no bracket, comments between single address`,
		in:   `(comment)local(comment)@domain`,
		exp:  `[<local@domain>]`,
	}, {
		desc: `With bracket, comments between local part`,
		in:   `<(comment)local(comment)@domain>`,
		exp:  `[<local@domain>]`,
	}, {
		desc: `With bracket, single address`,
		in:   `<(comment)local(comment)@(comment)domain>`,
		exp:  `[<local@domain>]`,
	}, {
		desc: `With bracket, single address`,
		in:   `<(comment)local(comment)@(comment)domain(comment(comment))>`,
		exp:  `[<local@domain>]`,
	}, {
		desc:   `With ';' on multiple mailboxes`,
		in:     `One <one@example> ; (comment)`,
		expErr: `ParseMailboxes: invalid character ';'`,
	}, {
		desc: `With group list, single address`,
		in:   `Group name: <(c)local(c)@(c)domain(c)>;(c)`,
		exp:  `[<local@domain>]`,
	}, {
		desc:   `With group, missing '>'`,
		in:     `Group name:One <one@example ; (comment)`,
		expErr: `ParseMailboxes: parseMailbox: missing '>'`,
	}, {
		desc:   `With group, missing ';'`,
		in:     `Group name:One <one@example>`,
		expErr: `ParseMailboxes: missing group terminator ';'`,
	}, {
		desc: `With group, without bracket`,
		in:   `Group name: one@example ; (comment)`,
		exp:  `[<one@example>]`,
	}, {
		desc:   `With group, without bracket, invalid domain`,
		in:     `Group name: one@exa[mple ; (comment)`,
		expErr: `ParseMailboxes: parseMailbox: invalid domain 'exa[mple'`,
	}, {
		desc:   `With group, trailing text before ';'`,
		in:     `Group name: <one@example> trail ; (comment)`,
		expErr: `ParseMailboxes: parseMailbox: unknown token 'trail'`,
	}, {
		desc: `With group, trailing text`,
		in:   `Group name: <(c)local(c)@(c)domain(c)>; trail(c)`,
		exp:  `[<local@domain>]`,
	}, {
		desc: `With group, multiple addresses`,
		in:   `(c)Group name(c): <(c)local(c)@(c)domain(c)>, Test One <test@one>;(c)`,
		exp:  `[<local@domain> Test One <test@one>]`,
	}, {
		desc:   `With list, invalid ','`,
		in:     `on,e@example , two@example`,
		expErr: `ParseMailboxes: empty or invalid address`,
	}, {
		desc:   `With list, missing '>'`,
		in:     `<one@example , <two@example>`,
		expErr: `ParseMailboxes: parseMailbox: missing '>'`,
	}, {
		desc:   `With list, invalid local #0`,
		in:     `one@example, @example`,
		expErr: `ParseMailboxes: parseMailbox: empty local`,
	}, {
		desc:   `With list, invalid local #1`,
		in:     `one@example, t[o@example`,
		expErr: `ParseMailboxes: parseMailbox: invalid local 't[o'`,
	}, {
		desc:   `With list, invalid local #2`,
		in:     `one@example, t)o@example`,
		expErr: `ParseMailboxes: parseMailbox: parseMailboxText: invalid character ')'`,
	}, {
		desc:   `With list, invalid local #3`,
		in:     `one@example, <t)o@example>`,
		expErr: `ParseMailboxes: parseMailbox: parseMailboxText: invalid character ')'`,
	}, {
		desc:   `With list, invalid domain #0`,
		in:     `one@ex[ample , <two@example>`,
		expErr: `ParseMailboxes: parseMailbox: invalid domain 'ex[ample'`,
	}, {
		desc:   `With list, invalid domain #1`,
		in:     `one@example, two@exa[mple`,
		expErr: `ParseMailboxes: parseMailbox: invalid domain 'exa[mple'`,
	}, {
		desc:   `With list, invalid domain #2`,
		in:     `one@example, two@exa)mple`,
		expErr: `ParseMailboxes: parseMailbox: parseMailboxText: invalid character ')'`,
	}, {
		desc:   `With list, trailing text #0`,
		in:     `<one@example> trail, <two@example>`,
		expErr: `ParseMailboxes: parseMailbox: unknown token 'trail'`,
	}, {
		desc:   `With list, trailing text #1`,
		in:     `<one@example> ), <two@example>`,
		expErr: `ParseMailboxes: parseMailbox: parseMailboxText: invalid character ')'`,
	}, {
		desc: `RFC 5322 example`,
		in: "A Group(Some people)\r\n" +
			"        :Chris Jones <c@(Chris's host.)public.example>,\r\n" +
			"            joe@example.org,\r\n" +
			"     John <jdoe@one.test> (my dear friend); (the end of the group)\r\n",
		exp: `[Chris Jones <c@public.example> <joe@example.org> John <jdoe@one.test>]`,
	}, {
		desc: `With null address (for Return-Path)`,
		in:   `<>`,
		exp:  `[<>]`,
	}, {
		desc: `With null address (for Return-Path) #2`,
		in:   `<(comment(comment))(comment)>`,
		exp:  `[<>]`,
	}}

	var (
		c      testCase
		mboxes []*Mailbox
		got    string
		err    error
	)

	for _, c = range cases {
		t.Log(c.desc)

		mboxes, err = ParseMailboxes([]byte(c.in))
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		got = fmt.Sprintf(`%+v`, mboxes)
		test.Assert(t, `Mailboxes`, c.exp, got)
	}
}

func TestParseMailboxText(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    string
	}{{
		desc: `With empty input`,
		in:   ``,
	}, {
		desc: `With empty comment`,
		in:   `()`,
	}, {
		desc: `With quoted-pair`,
		in:   `(\)) x`,
		exp:  `x`,
	}, {
		desc:   `With no closing parentheses`,
		in:     `(\) x`,
		expErr: `parseMailboxText: missing comment close parentheses`,
	}, {
		desc:   `With invalid nested comments`,
		in:     `((comment x`,
		expErr: `parseMailboxText: missing comment close parentheses`,
	}, {
		desc:   `With invalid nested comments`,
		in:     `((comment) x`,
		expErr: `parseMailboxText: missing comment close parentheses`,
	}, {
		desc: "With nested comments",
		in:   `(((\(comment))) x`,
		exp:  `x`,
	}, {
		desc: `With multiple comments`,
		in:   `(comment) (comment) x`,
		exp:  `x`,
	}, {
		desc: `With escaped char`,
		in:   `\(x(comment)`,
		exp:  `(x`,
	}}

	var (
		parser = libbytes.NewParser(nil, nil)
		got    []byte
		err    error
	)

	for _, c := range cases {
		t.Log(c.desc)

		parser.Reset([]byte(c.in), []byte{'\\', '(', ')'})

		got, _, err = parseMailboxText(parser)
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		test.Assert(t, `text`, c.exp, string(got))
	}
}

type ADT struct {
	Address *Mailbox `json:"address"`
}

func TestMailbox_UnmarshalJSON(t *testing.T) {
	jsonRaw := `{"address":"Name <local@domain>"}`

	got := &ADT{}
	err := json.Unmarshal([]byte(jsonRaw), got)
	if err != nil {
		t.Fatal(err)
	}

	exp := &ADT{
		Address: &Mailbox{
			Name:    `Name`,
			Local:   `local`,
			Domain:  `domain`,
			Address: `local@domain`,
			isAngle: true,
		},
	}

	test.Assert(t, "UnmarshalJSON", exp, got)
}

func TestMailbox_MarshalJSON(t *testing.T) {
	adt := &ADT{
		Address: &Mailbox{
			Name:    `Name`,
			Local:   `local`,
			Domain:  `domain`,
			Address: `local@domain`,
			isAngle: true,
		},
	}

	got, err := json.Marshal(adt)
	if err != nil {
		t.Fatal(err)
	}

	exp := `{"address":"Name \u003clocal@domain\u003e"}`

	test.Assert(t, "MarshalJSON", exp, string(got))

	un := &ADT{}

	err = json.Unmarshal(got, un)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "UnmarshalJSON", adt, un)
}
