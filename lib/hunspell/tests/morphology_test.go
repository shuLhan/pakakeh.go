// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tests

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/hunspell"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestMorphology_parseAnalyze(t *testing.T) {
	cases := []struct {
		line     string
		exp      hunspell.Morphemes
		expError string
	}{{
		line: "analyze(x)=a:b",
		exp: hunspell.Morphemes{
			"a": "b",
		},
	}, {
		line: "	analyze(x)	=	a:b	",
		exp: hunspell.Morphemes{
			"a": "b",
		},
	}, {
		line: "analyze(x) = a:",
		exp: hunspell.Morphemes{
			"a": "",
		},
	}, {
		line: "analyze(x) = :b",
		exp: hunspell.Morphemes{
			"": "b",
		},
	}}

	got := morphology{
		word: "x",
	}
	for _, c := range cases {
		got.analyze = nil

		err := got.parseAnalyze(c.line)
		if err != nil {
			test.Assert(t, c.line, c.expError, err.Error())
			continue
		}

		test.Assert(t, c.line, c.exp, got.analyze)
	}
}

func TestMorphology_parseStem(t *testing.T) {
	cases := []struct {
		line     string
		exp      string
		expError string
	}{{
		line: "stem(x)=x",
		exp:  "x",
	}, {
		line: "	stem(x) = x ",
		exp:  "x",
	}}

	got := morphology{
		word: "x",
	}

	for _, c := range cases {
		err := got.parseStem(c.line)
		if err != nil {
			test.Assert(t, c.line+" error", c.expError, err.Error())
			continue
		}
		test.Assert(t, c.line, c.exp, got.stem)
	}
}
