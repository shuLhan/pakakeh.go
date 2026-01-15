// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package email_test

import (
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/email"
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

	var key = `notexist`
	fmt.Printf("%s=%q\n", key, ct.GetParamValue(key))

	key = `KEY1`
	fmt.Printf("%s=%q\n", key, ct.GetParamValue(key))

	key = `key2`
	fmt.Printf("%s=%q\n", key, ct.GetParamValue(key))

	// Output:
	// notexist=""
	// KEY1="val1"
	// key2="value 2"
}

func ExampleContentType_SetBoundary() {
	var ct = &email.ContentType{}

	ct.SetBoundary(`42`)
	fmt.Println(ct.String())

	ct.SetBoundary(`43`)
	fmt.Println(ct.String())

	// Output:
	// /; boundary=42
	// /; boundary=43
}
