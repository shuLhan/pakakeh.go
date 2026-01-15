// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package dns

import (
	"bytes"
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestZoneParserDecodeString(t *testing.T) {
	type testCase struct {
		exp      string
		expError string
		in       []byte
	}

	var cases = []testCase{{
		in:  []byte(`"a\\b \"c\."`),
		exp: `a\\b "c\.`,
	}, {
		in:       []byte(`a\12a`),
		expError: `decodeString: invalid digits: \12a`,
	}, {
		in:       []byte(`a\12`),
		expError: `decodeString: invalid digits length: \12`,
	}, {
		in:       []byte(`a\999`),
		expError: `decodeString: invalid octet: \999`,
	}, {
		in:  []byte(`a\032b c`),
		exp: `a b`,
	}, {
		in:  []byte(`a\032b\.c`),
		exp: `a b.c`,
	}}

	var (
		zp = &zoneParser{}

		c   testCase
		got []byte
		err error
	)
	for _, c = range cases {
		got, err = zp.decodeString(c.in)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}
		test.Assert(t, string(c.in), c.exp, string(got))
	}
}

func TestZoneParser_next(t *testing.T) {
	var (
		logp = `TestZoneParser_next`

		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/zoneParser_next_test.txt`)
	if err != nil {
		t.Fatal(logp, err)
	}

	var listCase = []string{
		`comments`,
		`multiline`,
	}
	var (
		tag  string
		buf  bytes.Buffer
		zone Zone
		zp   zoneParser
	)
	for _, tag = range listCase {
		buf.Reset()
		zp.Reset(tdata.Input[tag], &zone)
		for {
			err = zp.next()
			if err != nil {
				t.Logf(`err:%s`, err)
				break
			}
			fmt.Fprintf(&buf, "%q %q\n", zp.token, zp.delim)
		}
		test.Assert(t, tag, string(tdata.Output[tag]), buf.String())
	}
}
