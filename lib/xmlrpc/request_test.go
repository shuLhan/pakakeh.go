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

func TestRequest_MarshalText(t *testing.T) {
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

		test.Assert(t, "Pack", c.exp, string(got))
	}
}

func TestRequest_UnmarshalText(t *testing.T) {
	cases := []struct {
		desc     string
		in       string
		exp      *Request
		expError string
	}{{
		desc: "Multiple param",
		in: `<?xml version="1.0"?>
			<methodCall>
				<methodName>method.name</methodName>
				<params>
					<param>
						<value>
							<string>
								param-string
							</string>
						</value>
					</param>
					<param>
						<value>
							<int>
								1
							</int>
						</value>
					</param>
				</params>
			</methodCall>`,
		exp: &Request{
			MethodName: "method.name",
			Params: []*Value{{
				Kind: String,
				In:   "param-string",
			}, {
				Kind: Integer,
				In:   int32(1),
			}},
		},
	}, {
		desc: "Param as struct",
		in: `<?xml version="1.0"?>
			<methodCall>
				<methodName>test.struct</methodName>
				<params>
					<param>
						<value>
							<struct>
								<member>
									<name>X</name>
									<value><int>1</int></value>
								</member>
								<member>
									<name>Y</name>
									<value><boolean>true</boolean></value>
								</member>
							</struct>
						</value>
					</param>
				</params>
			</methodCall>`,
		exp: &Request{
			MethodName: "test.struct",
			Params: []*Value{{
				Kind: Struct,
				StructMembers: []*Member{{
					Name: "X",
					Value: &Value{
						Kind: Integer,
						In:   int32(1),
					},
				}, {
					Name: "Y",
					Value: &Value{
						Kind: Boolean,
						In:   true,
					},
				}},
			}},
		},
	}, {
		desc: "Param as array",
		in: `<?xml version="1.0"?>
			<methodCall><methodName>test.array</methodName>
				<params><param><value><array><data>
					<value><string>a</string></value>
					<value><string>b</string></value>
				</data></array></value></param></params>
			</methodCall>`,
		exp: &Request{
			MethodName: "test.array",
			Params: []*Value{{
				Kind: Array,
				ArrayValues: []*Value{{
					Kind: String,
					In:   "a",
				}, {
					Kind: String,
					In:   "b",
				}},
			}},
		},
	}}

	for _, c := range cases {
		t.Logf(c.desc)

		got := &Request{}
		err := got.UnmarshalText([]byte(c.in))
		if err != nil {
			test.Assert(t, "Unmarshal", c.expError, err.Error())
			continue
		}

		test.Assert(t, "Unmarshal", c.exp, got)
	}
}
