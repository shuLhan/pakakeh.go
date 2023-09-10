// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestMergeDictionaries(t *testing.T) {
	cases := []struct {
		desc    string
		outFile string
		exp     string
		inFiles []string
		expN    int
	}{{
		desc:    "Without input file",
		outFile: "testdata/out.dic",
	}, {
		desc:    "With single input file",
		outFile: "testdata/out.dic",
		inFiles: []string{
			"testdata/in.dic",
		},
		exp: `3
hello
try/A
work/B
`,
	}, {
		desc:    "With two input files",
		outFile: "testdata/out.dic",
		inFiles: []string{
			"testdata/in.dic",
			"testdata/in2.dic",
		},
		exp: `5
a
c
hello
try/AC
work/B
`,
	}}

	for _, c := range cases {
		n, err := MergeDictionaries(c.outFile, c.inFiles...)
		if err != nil {
			t.Fatalf("%s: %s", c.desc, err)
		}

		if n == 0 {
			test.Assert(t, c.desc, c.expN, n)
			continue
		}

		got, err := os.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("%s: %s", c.desc, err)
		}

		test.Assert(t, c.desc, c.exp, string(got))
	}
}
