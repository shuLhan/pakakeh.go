// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package xmlrpc

import (
	"encoding/xml"
	"testing"

	liberrors "git.sr.ht/~shulhan/pakakeh.go/lib/errors"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestResponse_MarshalText(t *testing.T) {
	cases := []struct {
		desc string
		resp *Response
		exp  string
	}{{
		desc: "With param",
		resp: &Response{
			Param: &Value{
				Kind: Boolean,
				In:   true,
			},
		},
		exp: xml.Header + `<methodResponse><params><param><value><boolean>true</boolean></value></param></params></methodResponse>`,
	}, {
		desc: "With fault",
		resp: &Response{
			E: liberrors.E{
				Code:    404,
				Message: "Not found",
			},
		},
		exp: xml.Header + `<methodResponse>` +
			`<fault><value><struct>` +
			`<member>` +
			`<name>faultCode</name>` +
			`<value><int>404</int></value>` +
			`</member>` +
			`<member>` +
			`<name>faultString</name>` +
			`<value><string>Not found</string></value>` +
			`</member>` +
			`</struct></value></fault>` +
			`</methodResponse>`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := c.resp.MarshalText()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "MarshalText", c.exp, string(got))
	}
}

func TestResponse_UnmarshalText(t *testing.T) {
	cases := []struct {
		desc string
		text string
		exp  Response
	}{{
		desc: "Normal response with string",
		text: `<?xml version="1.0"?>
			<methodResponse>
				<params>
					<param>
						<value><string>West Sumatra</string></value>
					</param>
				</params>
			</methodResponse>`,
		exp: Response{
			Param: &Value{
				Kind: String,
				In:   "West Sumatra",
			},
		},
	}, {
		desc: "Faulty response",
		text: `<?xml version="1.0"?>
			<methodResponse>
				<fault>
					<value>
						<struct>
							<member>
								<name>faultCode</name>
								<value><int>4</int></value>
							</member>
							<member>
								<name>faultString</name>
								<value><string>Too many parameters.</string></value>
							</member>
						</struct>
					</value>
				</fault>
			</methodResponse>`,
		exp: Response{
			E: liberrors.E{
				Code:    4,
				Message: "Too many parameters.",
			},
		},
	}, {
		desc: "Response with array",
		text: `<?xml version="1.0"?>
			<methodResponse>
				<params>
					<param>
					<value>
						<array>
						<data>
							<value><string>North Sumatra</string></value>
							<value><string>West Sumatra</string></value>
							<value><string>South Sumatra</string></value>
						</data>
						</array>
					</value>
					</param>
				</params>
			</methodResponse>`,
		exp: Response{
			Param: &Value{
				Kind: Array,
				ArrayValues: []*Value{{
					Kind: String,
					In:   "North Sumatra",
				}, {
					Kind: String,
					In:   "West Sumatra",
				}, {
					Kind: String,
					In:   "South Sumatra",
				}},
			},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var got Response

		err := got.UnmarshalText([]byte(c.text))
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Response", c.exp, got)
	}
}
