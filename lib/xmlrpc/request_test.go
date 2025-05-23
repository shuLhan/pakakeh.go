// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package xmlrpc

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

type testStruct struct {
	X int32
	Y bool
}

func TestRequest_MarshalText(t *testing.T) {
	type testCase struct {
		methodName string
		params     []any
	}

	var cases = []testCase{{
		methodName: "method.name",
		params: []any{
			"param-string",
		},
	}, {
		methodName: "test.struct",
		params: []any{
			testStruct{
				X: 1,
				Y: true,
			},
		},
	}, {
		methodName: "test.array",
		params: []any{
			[]string{"a", "b"},
		},
	}}

	var (
		c     testCase
		req   Request
		tdata *test.Data
		got   []byte
		exp   []byte
		err   error
	)

	tdata, err = test.LoadData("testdata/marshal_test.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, c = range cases {
		req, err = NewRequest(c.methodName, c.params)
		if err != nil {
			t.Fatal(err)
		}

		got, err = req.MarshalText()
		if err != nil {
			t.Fatal(err)
		}

		exp = tdata.Output[c.methodName]
		test.Assert(t, "Pack", string(exp), string(got))
	}
}

func TestRequest_UnmarshalText(t *testing.T) {
	var (
		tdata    *test.Data
		name     string
		xmlInput []byte
		err      error
	)

	tdata, err = test.LoadData("testdata/unmarshal_test.txt")
	if err != nil {
		t.Fatal(err)
	}

	for name, xmlInput = range tdata.Input {
		t.Run(name, func(t *testing.T) {
			var (
				req  *Request
				exp  string
				got  string
				xmlb []byte
				err  error
			)

			exp = string(tdata.Output[name])

			req = &Request{}
			err = req.UnmarshalText(xmlInput)
			if err != nil {
				got = err.Error()
			} else {
				xmlb, err = req.MarshalText()
				if err != nil {
					t.Fatal(err)
				}
				got = string(xmlb)
			}

			test.Assert(t, name, exp, got)
		})
	}
}
