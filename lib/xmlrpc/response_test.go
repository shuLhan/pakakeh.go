// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestResponse(t *testing.T) {
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
			Param: Value{
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
			FaultCode:    4,
			FaultMessage: "Too many parameters.",
			IsFault:      true,
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
			Param: Value{
				Kind: Array,
				Values: []Value{{
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
		var got Response

		err := got.UnmarshalText([]byte(c.text))
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Response", c.exp, got, true)
	}
}
