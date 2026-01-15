// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package email_test

import (
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/email"
)

func ExampleHeader_Filter() {
	// Overwrite the email.Epoch to make the example works.
	email.Epoch = func() int64 {
		return 1645600000
	}

	var (
		fromAddress = []byte("Noreply <noreply@example.com>")
		toAddresses = []byte("John <john@example.com>, Jane <jane@example.com>")
		subject     = []byte(`Example subject`)
		bodyText    = []byte(`Email body as plain text`)
		bodyHTML    = []byte(`Email body as <b>HTML</b>`)

		msg *email.Message
		err error
	)

	msg, err = email.NewMultipart(
		fromAddress,
		toAddresses,
		subject,
		bodyText,
		bodyHTML,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = msg.Pack()

	var (
		fields []*email.Field
		field  *email.Field
	)

	fields = msg.Header.Filter(email.FieldTypeFrom)

	for _, field = range fields {
		fmt.Printf("%s: %s\n", field.Name, field.Value)
	}
	// Output:
	// from: Noreply <noreply@example.com>
}
