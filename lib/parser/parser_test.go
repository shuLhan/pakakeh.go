// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParser_AddDelimiters(t *testing.T) {
	p := &Parser{
		delims: "/:",
	}

	cases := []struct {
		delims string
		exp    string
	}{{
		exp: "/:",
	}, {
		delims: " \t",
		exp:    "/: \t",
	}, {
		delims: " \t",
		exp:    "/: \t",
	}}

	for _, c := range cases {
		p.AddDelimiters(c.delims)
		test.Assert(t, "p.delims", c.exp, p.delims, true)
	}
}

func TestParser_Lines(t *testing.T) {
	cases := []struct {
		desc    string
		content string
		exp     []string
	}{{
		desc: "With empty content",
		exp:  []string{},
	}, {
		desc:    "With single empty line",
		content: "\n",
		exp:     []string{},
	}, {
		desc:    "With single empty line",
		content: " \t\r\f\n",
		exp:     []string{},
	}, {
		desc:    "With one line, at the end",
		content: " \t\r\f\ntest",
		exp: []string{
			"test",
		},
	}, {
		desc:    "With one line, in the middle",
		content: " \t\r\f\ntest \t\r\f\n",
		exp: []string{
			"test",
		},
	}, {
		desc:    "With two lines",
		content: "A \t\f\r\n \nB \t\f\r\n",
		exp: []string{
			"A",
			"B",
		},
	}, {
		desc:    "With three lines",
		content: "A \t\f\r\n \n\n\nB\n \t\f\r\nC",
		exp: []string{
			"A",
			"B",
			"C",
		},
	}}

	p := New("", "")

	for _, c := range cases {
		t.Log(c.desc)

		p.Load(c.content, "")

		got := p.Lines()

		test.Assert(t, "Lines()", c.exp, got, true)
	}
}

func TestParser_Stop(t *testing.T) {
	p := New("\t test \ntest", "")

	cases := []struct {
		exp string
	}{{
		exp: " test \ntest",
	}, {
		exp: "test \ntest",
	}, {
		exp: "\ntest",
	}, {
		exp: "test",
	}, {
		exp: "",
	}}

	var got string
	for _, c := range cases {
		_, _ = p.Token()
		got, _ = p.Stop()
		test.Assert(t, "Stop", c.exp, got, true)
		p.Load(got, "")
	}
}

func TestParser_Token(t *testing.T) {
	p := New("\t test \ntest", "")

	cases := []struct {
		expToken string
		expDelim rune
	}{{
		expDelim: '\t',
	}, {
		expDelim: ' ',
	}, {
		expToken: "test",
		expDelim: ' ',
	}, {
		expDelim: '\n',
	}, {
		expToken: "test",
	}}

	for _, c := range cases {
		gotToken, gotDelim := p.Token()

		test.Assert(t, "token", c.expToken, gotToken, true)
		test.Assert(t, "delim", c.expDelim, gotDelim, true)
	}
}

func TestParser_TokenEscaped(t *testing.T) {
	p := New("\t te\\ st \ntest", "")

	cases := []struct {
		expToken string
		expDelim rune
	}{{
		expDelim: '\t',
	}, {
		expDelim: ' ',
	}, {
		expToken: "te st",
		expDelim: ' ',
	}, {
		expDelim: '\n',
	}, {
		expToken: "test",
	}}

	for _, c := range cases {
		gotToken, gotDelim := p.TokenEscaped('\\')

		test.Assert(t, "token", c.expToken, gotToken, true)
		test.Assert(t, "delim", c.expDelim, gotDelim, true)
	}
}

func TestParser_SkipLine(t *testing.T) {
	cases := []struct {
		desc     string
		content  string
		expToken string
		expDelim rune
	}{{
		desc: "With empty content",
	}, {
		desc:     "With empty line",
		content:  "\ntest\n",
		expToken: "test",
		expDelim: '\n',
	}, {
		desc:    "With single line",
		content: "test\n",
	}, {
		desc:     "With two lines",
		content:  "test 1\ntest 2",
		expToken: "test",
		expDelim: ' ',
	}}

	p := New("", "")

	for _, c := range cases {
		t.Log(c.desc)

		p.Load(c.content, "")

		p.SkipLine()

		gotToken, gotDelim := p.Token()

		test.Assert(t, "token", c.expToken, gotToken, true)
		test.Assert(t, "delim", c.expDelim, gotDelim, true)
	}
}

func TestParser_Open(t *testing.T) {
	cases := []struct {
		desc       string
		file       string
		expError   string
		expContent string
	}{{
		desc:     "With not existing file",
		file:     "testdata/xxx",
		expError: "open testdata/xxx: no such file or directory",
	}, {
		desc:       "With file exist",
		file:       "testdata/test.txt",
		expContent: "test\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		p, err := Open(c.file, "")
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}

		test.Assert(t, "content", c.expContent, p.v, true)
	}
}

func TestParser_RemoveDelimiters(t *testing.T) {
	p := &Parser{
		delims: "/: \t",
	}
	cases := []struct {
		delims string
		exp    string
	}{{
		exp: "/: \t",
	}, {
		delims: "/",
		exp:    ": \t",
	}, {
		delims: "///",
		exp:    ": \t",
	}, {
		delims: "\t :",
		exp:    "",
	}}

	for _, c := range cases {
		p.RemoveDelimiters(c.delims)
		test.Assert(t, "p.delims", c.exp, p.delims, true)
	}
}

func TestParser_SkipHorizontalSpaces(t *testing.T) {
	cases := []struct {
		desc     string
		content  string
		expToken string
		expRune  rune
		expDelim rune
	}{{
		desc: "With empty content",
	}, {
		desc:     "With empty line",
		content:  " \t\r\f\n",
		expRune:  '\n',
		expDelim: '\n',
	}, {
		desc:     "With single line",
		content:  "test\n",
		expRune:  't',
		expToken: "test",
		expDelim: '\n',
	}, {
		desc:     "With space in the beginning",
		content:  " \t\f\rtest 1\ntest 2",
		expRune:  't',
		expToken: "test",
		expDelim: ' ',
	}}

	p := New("", "")

	for _, c := range cases {
		t.Log(c.desc)

		p.Load(c.content, "")

		got := p.SkipHorizontalSpaces()

		test.Assert(t, "rune", c.expRune, got, true)

		gotToken, gotDelim := p.Token()

		test.Assert(t, "token", c.expToken, gotToken, true)
		test.Assert(t, "delim", c.expDelim, gotDelim, true)
	}
}
