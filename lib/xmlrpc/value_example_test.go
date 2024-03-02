// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"fmt"
	"log"
)

func ExampleValue_GetFieldAsBoolean() {
	var (
		xmlb = `<?xml version="1.0"?>
<methodResponse>
<params>
	<param>
		<value>
			<struct>
				<member>
					<name>boolean_false</name>
					<value><boolean>0</boolean></value>
				</member>
				<member>
					<name>boolean_true</name>
					<value><boolean>1</boolean></value>
				</member>
				<member>
					<name>string_0</name>
					<value><string>0</string></value>
				</member>
				<member>
					<name>string_1</name>
					<value><string>1</string></value>
				</member>
			</struct>
		</value>
	</param>
</params>
</methodResponse>
`
		res = &Response{}

		err error
	)

	err = res.UnmarshalText([]byte(xmlb))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Get boolean field as string:")
	fmt.Println(res.Param.GetFieldAsString("boolean_false"))
	fmt.Println(res.Param.GetFieldAsString("boolean_true"))
	fmt.Println("Get boolean field as boolean:")
	fmt.Println(res.Param.GetFieldAsBoolean("boolean_false"))
	fmt.Println(res.Param.GetFieldAsBoolean("boolean_true"))
	fmt.Println("Get string field as boolean:")
	fmt.Println(res.Param.GetFieldAsBoolean("string_0"))
	fmt.Println(res.Param.GetFieldAsBoolean("string_1"))
	// Output:
	// Get boolean field as string:
	// false
	// true
	// Get boolean field as boolean:
	// false
	// true
	// Get string field as boolean:
	// false
	// false
}
