// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

type testStruct struct {
	X int32
	Y bool
}

func TestRequest(t *testing.T) {
	cases := []struct {
		methodName string
		params     []interface{}
		exp        string
	}{{
		methodName: "method.name",
		params: []interface{}{
			"param-string",
		},
		exp: "<methodCall><methodName>method.name</methodName>" +
			"<params>" +
			"<param>" +
			"<value><string>param-string</string></value>" +
			"</param>" +
			"</params>" +
			"</methodCall>",
	}, {
		methodName: "test.struct",
		params: []interface{}{
			testStruct{
				X: 1,
				Y: true,
			},
		},
		exp: "<methodCall><methodName>test.struct</methodName>" +
			"<params><param><value><struct>" +
			"<member>" +
			"<name>X</name><value><int>1</int></value>" +
			"</member>" +
			"<member>" +
			"<name>Y</name><value><boolean>true</boolean></value>" +
			"</member>" +
			"</struct></value></param></params>" +
			"</methodCall>",
	}, {
		methodName: "test.array",
		params: []interface{}{
			[]string{"a", "b"},
		},
		exp: "<methodCall><methodName>test.array</methodName>" +
			"<params><param><value><array><data>" +
			"<value><string>a</string></value>" +
			"<value><string>b</string></value>" +
			"</data></array></value></param></params>" +
			"</methodCall>",
	}}

	for _, c := range cases {
		req, err := NewRequest(c.methodName, c.params)
		if err != nil {
			t.Fatal(err)
		}

		got, err := req.MarshalText()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Pack", c.exp, string(got), true)
	}
}
