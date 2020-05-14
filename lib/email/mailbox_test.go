// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"fmt"
	"testing"

	libio "github.com/shuLhan/share/lib/io"
	"github.com/shuLhan/share/lib/test"
)

func TestParseMailboxes(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    string
	}{{
		desc:   "With empty input",
		expErr: "ParseMailboxes: empty address",
	}, {
		desc:   "With comment only",
		in:     "(comment)",
		expErr: "ParseMailboxes: empty or invalid address",
	}, {
		desc:   "With no domain",
		in:     "(comment)local(comment)",
		expErr: "ParseMailboxes: empty or invalid address",
	}, {
		desc:   "With no opening comment",
		in:     "comment)local@domain",
		expErr: "ParseMailboxes: invalid local: 'comment)local'",
	}, {
		desc:   "With no closing comment",
		in:     "(commentlocal@domain",
		expErr: "missing comment close parentheses",
	}, {
		desc:   "With no opening bracket",
		in:     "(comment)local(comment)@domain>",
		expErr: "ParseMailboxes: invalid character: '>'",
	}, {
		desc:   "With no closing bracket",
		in:     "<(comment)local(comment)@domain",
		expErr: "ParseMailboxes: missing '>'",
	}, {
		desc:   "With ':' inside mailbox",
		in:     "<local:part@domain>",
		expErr: "ParseMailboxes: invalid character: ':'",
	}, {
		desc: "With '<' inside local part",
		in:   "local<part@domain>",
		exp:  "[local <part@domain>]",
	}, {
		desc:   "With multiple '<'",
		in:     "Name <local<part@domain>",
		expErr: "ParseMailboxes: invalid character: '<'",
	}, {
		desc:   "With multiple '@'",
		in:     "Name <local@part@domain>",
		expErr: "ParseMailboxes: invalid character: '@'",
	}, {
		desc: "With no domain",
		in:   "Name <local>",
		exp:  "[Name <@local>]",
	}, {
		desc:   "With empty local",
		in:     "Name <@domain>",
		expErr: "ParseMailboxes: empty local",
	}, {
		desc:   "With empty domain",
		in:     "Name <local@>, test@domain",
		expErr: "ParseMailboxes: invalid domain: ''",
	}, {
		desc:   "With invalid domain",
		in:     "Name <local@dom[ain>, test@domain",
		expErr: "ParseMailboxes: invalid domain: 'dom[ain'",
	}, {
		desc: "With no bracket, single address",
		in:   "local@domain",
		exp:  "[<local@domain>]",
	}, {
		desc: "With no bracket, comments between single address",
		in:   "(comment)local(comment)@domain",
		exp:  "[<local@domain>]",
	}, {
		desc: "With bracket, comments between local part",
		in:   "<(comment)local(comment)@domain>",
		exp:  "[<local@domain>]",
	}, {
		desc: "With bracket, single address",
		in:   "<(comment)local(comment)@(comment)domain>",
		exp:  "[<local@domain>]",
	}, {
		desc: "With bracket, single address",
		in:   "<(comment)local(comment)@(comment)domain(comment(comment))>",
		exp:  "[<local@domain>]",
	}, {
		desc:   "With ';' on multiple mailboxes",
		in:     "One <one@example> ; (comment)",
		expErr: "ParseMailboxes: invalid character: ';'",
	}, {
		desc: "With group list, single address",
		in:   "Group name: <(c)local(c)@(c)domain(c)>;(c)",
		exp:  "[<local@domain>]",
	}, {
		desc:   "With group, missing '>'",
		in:     "Group name:One <one@example ; (comment)",
		expErr: "ParseMailboxes: missing '>'",
	}, {
		desc:   "With group, missing ';'",
		in:     "Group name:One <one@example>",
		expErr: "ParseMailboxes: missing ';'",
	}, {
		desc: "With group, without bracket",
		in:   "Group name: one@example ; (comment)",
		exp:  "[<one@example>]",
	}, {
		desc:   "With group, without bracket, invalid domain",
		in:     "Group name: one@exa[mple ; (comment)",
		expErr: "ParseMailboxes: invalid domain: 'exa[mple'",
	}, {
		desc:   "With group, trailing text before ';'",
		in:     "Group name: <one@example> trail ; (comment)",
		expErr: "ParseMailboxes: invalid token: 'trail'",
	}, {
		desc:   "With group, trailing text",
		in:     "Group name: <(c)local(c)@(c)domain(c)>; trail(c)",
		expErr: "ParseMailboxes: trailing text: 'trail'",
	}, {
		desc: "With group, multiple addresses",
		in:   "(c)Group name(c): <(c)local(c)@(c)domain(c)>, Test One <test@one>;(c)",
		exp:  "[<local@domain> Test One <test@one>]",
	}, {
		desc:   "With list, invalid ','",
		in:     "on,e@example , two@example",
		expErr: "ParseMailboxes: invalid character: ','",
	}, {
		desc:   "With list, missing '>'",
		in:     "<one@example , <two@example>",
		expErr: "ParseMailboxes: missing '>'",
	}, {
		desc:   "With list, invalid domain",
		in:     "one@ex[ample , <two@example>",
		expErr: "ParseMailboxes: invalid domain: 'ex[ample'",
	}, {
		desc:   "With list, invalid domain",
		in:     "one@example, two@exa[mple",
		expErr: "ParseMailboxes: invalid domain: 'exa[mple'",
	}, {
		desc:   "With list, trailing text after '>'",
		in:     "<one@example> trail, <two@example>",
		expErr: "ParseMailboxes: invalid token: 'trail'",
	}, {
		desc: "RFC 5322 example",
		in: "A Group(Some people)\r\n" +
			"        :Chris Jones <c@(Chris's host.)public.example>,\r\n" +
			"            joe@example.org,\r\n" +
			"     John <jdoe@one.test> (my dear friend); (the end of the group)\r\n",
		exp: "[Chris Jones <c@public.example> <joe@example.org> John <jdoe@one.test>]",
	}, {
		desc: "With null address (for Return-Path)",
		in:   "<>",
		exp:  "[<>]",
	}, {
		desc: "With null address (for Return-Path)",
		in:   "<(comment(comment))(comment)>",
		exp:  "[<>]",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		mboxes, err := ParseMailboxes([]byte(c.in))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		got := fmt.Sprintf("%+v", mboxes)
		test.Assert(t, "Mailboxes", c.exp, got, true)
	}
}

func TestSkipComment(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    string
	}{{
		desc:   "With empty input",
		in:     "",
		expErr: "missing comment close parentheses",
	}, {
		desc: "With empty comment",
		in:   "()",
	}, {
		desc: "With quoted-pair",
		in:   `(\)) x`,
		exp:  `x`,
	}, {
		desc:   "With no closing parentheses",
		in:     `(\) x`,
		expErr: "missing comment close parentheses",
	}, {
		desc:   "With invalid nested comments",
		in:     `((comment x`,
		expErr: "missing comment close parentheses",
	}, {
		desc:   "With invalid nested comments",
		in:     `((comment) x`,
		expErr: "missing comment close parentheses",
	}, {
		desc: "With nested comments",
		in:   `(((\(comment))) x`,
		exp:  `x`,
	}, {
		desc: "With multiple comments",
		in:   `(comment) (comment) x`,
		exp:  `x`,
	}}

	r := &libio.Reader{}

	for _, c := range cases {
		t.Log(c.desc)
		r.Init([]byte(c.in))

		r.SkipN(1)
		_, err := skipComment(r)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		got := string(r.Rest())

		test.Assert(t, "rest", c.exp, got, true)
	}
}
