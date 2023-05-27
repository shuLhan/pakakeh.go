// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email_test

import (
	"fmt"
	"log"

	"github.com/shuLhan/share/lib/email"
)

func ExampleParseContentType() {
	var (
		raw = []byte(`text/plain; key1=val1; key2="value 2"`)

		ct  *email.ContentType
		err error
	)

	ct, err = email.ParseContentType(raw)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ct.String())

	// Output:
	// text/plain; key1=val1; key2="value 2"
}

func ExampleContentType_GetParamValue() {
	var (
		raw = []byte(`text/plain; key1=val1; key2="value 2"`)

		ct  *email.ContentType
		err error
	)

	ct, err = email.ParseContentType(raw)
	if err != nil {
		log.Fatal(err)
	}

	var key = []byte(`notexist`)
	fmt.Printf("%s=%q\n", key, ct.GetParamValue(key))

	key = []byte(`KEY1`)
	fmt.Printf("%s=%q\n", key, ct.GetParamValue(key))

	key = []byte(`key2`)
	fmt.Printf("%s=%q\n", key, ct.GetParamValue(key))

	// Output:
	// notexist=""
	// KEY1="val1"
	// key2="value 2"

}

func ExampleContentType_SetBoundary() {
	var ct = &email.ContentType{}

	ct.SetBoundary([]byte(`42`))
	fmt.Println(ct.String())

	ct.SetBoundary([]byte(`43`))
	fmt.Println(ct.String())

	// Output:
	// /; boundary=42
	// /; boundary=43
}
