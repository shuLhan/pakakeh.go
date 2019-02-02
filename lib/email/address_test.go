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

func TestParseAddress(t *testing.T) {
	cases := []struct {
		desc   string
		in     []byte
		expErr string
		exp    string
	}{{
		desc:   "With empty input",
		expErr: "ParseAddress: empty address",
	}, {
		desc:   "With comment only",
		in:     []byte("(comment)"),
		expErr: "ParseAddress: empty or invalid address",
	}, {
		desc:   "With no domain",
		in:     []byte("(comment)local(comment)"),
		expErr: "ParseAddress: empty or invalid address",
	}, {
		desc:   "With no opening comment",
		in:     []byte("comment)local@domain"),
		expErr: "ParseAddress: invalid local: 'comment)local'",
	}, {
		desc:   "With no closing comment",
		in:     []byte("(commentlocal@domain"),
		expErr: "ParseAddress: missing comment close parentheses",
	}, {
		desc:   "With no opening bracket",
		in:     []byte("(comment)local(comment)@domain>"),
		expErr: "ParseAddress: invalid character: '>'",
	}, {
		desc:   "With no closing bracket",
		in:     []byte("<(comment)local(comment)@domain"),
		expErr: "ParseAddress: missing '>'",
	}, {
		desc:   "With ':' inside mailbox",
		in:     []byte("<local:part@domain>"),
		expErr: "ParseAddress: invalid character: ':'",
	}, {
		desc: "With '<' inside local part",
		in:   []byte("local<part@domain>"),
		exp:  "[local <part@domain>]",
	}, {
		desc:   "With multiple '<'",
		in:     []byte("Name <local<part@domain>"),
		expErr: "ParseAddress: invalid character: '<'",
	}, {
		desc:   "With multiple '@'",
		in:     []byte("Name <local@part@domain>"),
		expErr: "ParseAddress: invalid character: '@'",
	}, {
		desc:   "With empty local",
		in:     []byte("Name <@domain>"),
		expErr: "ParseAddress: empty local",
	}, {
		desc:   "With empty domain",
		in:     []byte("Name <local@>, test@domain"),
		expErr: "ParseAddress: invalid domain: ''",
	}, {
		desc:   "With invalid domain",
		in:     []byte("Name <local@dom[ain>, test@domain"),
		expErr: "ParseAddress: invalid domain: 'dom[ain'",
	}, {
		desc: "With no bracket, single address",
		in:   []byte("local@domain"),
		exp:  "[<local@domain>]",
	}, {
		desc: "With no bracket, comments between single address",
		in:   []byte("(comment)local(comment)@domain"),
		exp:  "[<local@domain>]",
	}, {
		desc: "With bracket, comments between local part",
		in:   []byte("<(comment)local(comment)@domain>"),
		exp:  "[<local@domain>]",
	}, {
		desc: "With bracket, single address",
		in:   []byte("<(comment)local(comment)@(comment)domain>"),
		exp:  "[<local@domain>]",
	}, {
		desc: "With bracket, single address",
		in:   []byte("<(comment)local(comment)@(comment)domain(comment(comment))>"),
		exp:  "[<local@domain>]",
	}, {
		desc:   "With ';' on multiple mailboxes",
		in:     []byte("One <one@example> ; (comment)"),
		expErr: "ParseAddress: invalid character: ';'",
	}, {
		desc: "With group list, single address",
		in:   []byte("Group name: <(c)local(c)@(c)domain(c)>;(c)"),
		exp:  "[<local@domain>]",
	}, {
		desc:   "With group, missing '>'",
		in:     []byte("Group name:One <one@example ; (comment)"),
		expErr: "ParseAddress: missing '>'",
	}, {
		desc:   "With group, missing ';'",
		in:     []byte("Group name:One <one@example>"),
		expErr: "ParseAddress: missing ';'",
	}, {
		desc: "With group, without bracket",
		in:   []byte("Group name: one@example ; (comment)"),
		exp:  "[<one@example>]",
	}, {
		desc:   "With group, without bracket, invalid domain",
		in:     []byte("Group name: one@exa[mple ; (comment)"),
		expErr: "ParseAddress: invalid domain: 'exa[mple'",
	}, {
		desc:   "With group, trailing text before ';'",
		in:     []byte("Group name: <one@example> trail ; (comment)"),
		expErr: "ParseAddress: invalid token: 'trail'",
	}, {
		desc:   "With group, trailing text",
		in:     []byte("Group name: <(c)local(c)@(c)domain(c)>; trail(c)"),
		expErr: "ParseAddress: trailing text: 'trail'",
	}, {
		desc: "With group, multiple addresses",
		in:   []byte("(c)Group name(c): <(c)local(c)@(c)domain(c)>, Test One <test@one>;(c)"),
		exp:  "[<local@domain> Test One <test@one>]",
	}, {
		desc:   "With list, invalid ','",
		in:     []byte("on,e@example , two@example"),
		expErr: "ParseAddress: invalid character: ','",
	}, {
		desc:   "With list, missing '>'",
		in:     []byte("<one@example , <two@example>"),
		expErr: "ParseAddress: missing '>'",
	}, {
		desc:   "With list, invalid domain",
		in:     []byte("one@ex[ample , <two@example>"),
		expErr: "ParseAddress: invalid domain: 'ex[ample'",
	}, {
		desc:   "With list, invalid domain",
		in:     []byte("one@example, two@exa[mple"),
		expErr: "ParseAddress: invalid domain: 'exa[mple'",
	}, {
		desc:   "With list, trailing text after '>'",
		in:     []byte("<one@example> trail, <two@example>"),
		expErr: "ParseAddress: invalid token: 'trail'",
	}, {
		desc: "RFC 5322 example",
		in: []byte("A Group(Some people)\r\n" +
			"        :Chris Jones <c@(Chris's host.)public.example>,\r\n" +
			"            joe@example.org,\r\n" +
			"     John <jdoe@one.test> (my dear friend); (the end of the group)\r\n"),
		exp: "[Chris Jones <c@public.example> <joe@example.org> John <jdoe@one.test>]",
	}, {
		desc: "With null address (for Return-Path)",
		in:   []byte("<>"),
		exp:  "[<>]",
	}, {
		desc: "With null address (for Return-Path)",
		in:   []byte("<(comment(comment))(comment)>"),
		exp:  "[<>]",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		mboxes, err := ParseAddress(c.in)
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
		in     []byte
		expErr string
		exp    []byte
	}{{
		desc:   "With empty input",
		in:     []byte(""),
		expErr: "ParseAddress: missing comment close parentheses",
	}, {
		desc: "With empty comment",
		in:   []byte("()"),
		exp:  []byte{},
	}, {
		desc: "With quoted-pair",
		in:   []byte(`(\)) x`),
		exp:  []byte("x"),
	}, {
		desc:   "With no closing parentheses",
		in:     []byte(`(\) x`),
		expErr: "ParseAddress: missing comment close parentheses",
	}, {
		desc:   "With invalid nested comments",
		in:     []byte(`((comment x`),
		expErr: "ParseAddress: missing comment close parentheses",
	}, {
		desc:   "With invalid nested comments",
		in:     []byte(`((comment) x`),
		expErr: "ParseAddress: missing comment close parentheses",
	}, {
		desc: "With nested comments",
		in:   []byte(`(((\(comment))) x`),
		exp:  []byte("x"),
	}, {
		desc: "With multiple comments",
		in:   []byte(`(comment) (comment) x`),
		exp:  []byte("x"),
	}}

	r := &libio.Reader{}

	for _, c := range cases {
		t.Log(c.desc)
		r.InitBytes(c.in)

		r.SkipN(1)
		_, err := skipComment(r)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		got := r.Rest()

		test.Assert(t, "rest", c.exp, got, true)
	}
}
