// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package strings

import (
	"strings"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestCleanURI(t *testing.T) {
	var (
		tdata *test.Data
		err   error
		exp   string
		got   string
	)

	tdata, err = test.LoadData(`testdata/clean_uri_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	exp = string(tdata.Output[`default`])
	got = CleanURI(string(tdata.Input[`default`]))

	test.Assert(t, ``, exp, got)
}

func TestCleanWikiMarkup(t *testing.T) {
	var (
		tdata *test.Data
		err   error
		exp   string
		got   string
	)

	tdata, err = test.LoadData(`testdata/clean_wiki_markup_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	exp = string(tdata.Output[`default`])
	got = CleanWikiMarkup(string(tdata.Input[`default`]))

	test.Assert(t, ``, exp, got)
}

func TestMergeSpaces(t *testing.T) {
	type testCase struct {
		text     string
		exp      string
		withline bool
	}
	var cases = []testCase{{
		text: "   a\n\nb c   d\n\n",
		exp:  " a\n\nb c d\n\n",
	}, {
		text: " \t a \t ",
		exp:  " a ",
	}, {
		text:     "   a\n\nb c   d\n\n",
		withline: true,
		exp:      " a\nb c d\n",
	}}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		got = MergeSpaces(c.text, c.withline)
		test.Assert(t, ``, c.exp, got)
	}
}

func TestReverse(t *testing.T) {
	type testCase struct {
		input string
		exp   string
	}
	var cases = []testCase{{
		input: `The quick bròwn 狐 jumped over the lazy 犬`,
		exp:   `犬 yzal eht revo depmuj 狐 nwòrb kciuq ehT`,
	}}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		got = Reverse(c.input)
		test.Assert(t, `Reverse`, c.exp, got)
	}
}

func TestSingleSpace(t *testing.T) {
	type testCase struct {
		in  string
		exp string
	}

	var cases = []testCase{{
		// Empty input.
	}, {
		in:  " \t\v\r\n\r\n\fa \t\v\r\n\r\n\f",
		exp: " a ",
	}}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		got = SingleSpace(c.in)
		test.Assert(t, c.in, c.exp, got)
	}
}

func TestSplit(t *testing.T) {
	var (
		tdata *test.Data
		err   error
		name  string
		got   []string
		exp   []string
		raw   []byte
	)

	tdata, err = test.LoadData(`testdata/split_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	for name, raw = range tdata.Input {
		got = Split(string(raw), true, true)
		raw = tdata.Output[name]
		exp = strings.Fields(string(raw))
		test.Assert(t, name, exp, got)
	}
}

func TestTrimNonAlnum(t *testing.T) {
	type testCase struct {
		text string
		exp  string
	}
	var cases = []testCase{
		{`[[alpha]]`, `alpha`},
		{`[[alpha`, `alpha`},
		{`alpha]]`, `alpha`},
		{`alpha`, `alpha`},
		{`alpha0`, `alpha0`},
		{`1alpha`, `1alpha`},
		{`1alpha0`, `1alpha0`},
		{`[][][]`, ``},
	}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		got = TrimNonAlnum(c.text)
		test.Assert(t, ``, c.exp, got)
	}
}
