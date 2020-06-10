// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func cleanup() {
	// Cleaning up TestRmdirEmptyAll
	_ = os.Remove("testdata/file")
	_ = os.RemoveAll("testdata/a")
	_ = os.RemoveAll("testdata/dirempty")
}

func TestMain(m *testing.M) {
	cleanup()

	s := m.Run()

	cleanup()

	os.Exit(s)
}

func TestCopy(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		out    string
		expErr string
		exp    string
	}{{
		desc:   "Without output file",
		in:     "testdata/input.txt",
		expErr: `Copy: failed to open output file: open : no such file or directory`,
	}, {
		desc:   "Without input file",
		out:    "testdata/output.txt",
		expErr: `Copy: failed to open input file: open : no such file or directory`,
	}, {
		desc: "With input and output",
		in:   "testdata/input.txt",
		out:  "testdata/output.txt",
		exp: `Copyright (c) 2018 M. Shulhan (ms@kilabit.info). All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of M. Shulhan, nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
`,
	}}

	for _, c := range cases {
		err := Copy(c.out, c.in)
		if err != nil {
			test.Assert(t, c.desc, c.expErr, err.Error(), true)
			continue
		}

		got, err := ioutil.ReadFile(c.out)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.desc, c.exp, string(got), true)
	}
}
