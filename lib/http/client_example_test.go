// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

package http_test

import (
	"crypto/rand"
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test/mock"
)

func ExampleGenerateFormData() {
	// Mock the random reader for predictable output.
	// NOTE: do not do this on real code.
	rand.Reader = mock.NewRandReader([]byte(`randomseed`))

	var data = &multipart.Form{
		Value: map[string][]string{
			`name`: []string{`test.txt`},
			`size`: []string{`42`},
		},
	}

	var (
		contentType string
		body        string
		err         error
	)
	contentType, body, err = libhttp.GenerateFormData(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(`contentType:`, contentType)
	fmt.Println(`body:`)
	fmt.Println(strings.ReplaceAll(body, "\r\n", "\n"))
	// Output:
	// contentType: multipart/form-data; boundary=72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564
	// body:
	// --72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564
	// Content-Disposition: form-data; name="name"
	//
	// test.txt
	// --72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564
	// Content-Disposition: form-data; name="size"
	//
	// 42
	// --72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564--
}
