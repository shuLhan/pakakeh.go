// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package bytes

import (
	"bytes"
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParser_Read(t *testing.T) {
	var (
		logp = `TestParser`

		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/Parser_Read_test.txt`)
	if err != nil {
		t.Fatal(logp, err)
	}

	var listCase = []string{
		`multiline`,
	}

	var (
		tcase   string
		tag     string
		parser  *Parser
		out     bytes.Buffer
		content []byte
		delims  []byte
		token   []byte
		c       byte
	)

	for _, tcase = range listCase {
		content = tdata.Input[tcase]
		delims = tdata.Input[tcase+`:delims`]

		parser = NewParser(content, delims)

		out.Reset()
		for {
			token, c = parser.Read()
			fmt.Fprintf(&out, "%q %q\n", token, c)
			if c == 0 {
				break
			}
		}
		tag = tcase + `:Read`
		test.Assert(t, tag, string(tdata.Output[tag]), out.String())
	}
}

func TestParser_ReadNoSpace(t *testing.T) {
	var (
		logp = `TestParser`

		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/Parser_ReadNoSpace_test.txt`)
	if err != nil {
		t.Fatal(logp, err)
	}

	var listCase = []string{
		`multiline`,
	}

	var (
		tcase   string
		tag     string
		parser  *Parser
		out     bytes.Buffer
		content []byte
		delims  []byte
		token   []byte
		c       byte
	)

	for _, tcase = range listCase {
		content = tdata.Input[tcase]
		delims = tdata.Input[tcase+`:delims`]
		parser = NewParser(content, delims)
		out.Reset()
		for {
			token, c = parser.ReadNoSpace()
			fmt.Fprintf(&out, "%q %q\n", token, c)
			if c == 0 {
				break
			}
		}
		tag = tcase + `:ReadNoSpace`
		test.Assert(t, tag, string(tdata.Output[tag]), out.String())
	}
}
