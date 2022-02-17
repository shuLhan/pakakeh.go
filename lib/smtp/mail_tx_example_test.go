// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"

	"github.com/shuLhan/share/lib/email"
)

func ExampleNewMailTx() {
	// Example on how to create MailTx Data using email package [1].
	//
	// [1] github.com/shuLhan/share/lib/email

	var (
		txFrom      = "Postmaster <postmaster@mail.example.com>"
		fromAddress = []byte("Noreply <noreply@example.com>")
		subject     = []byte(`Example subject`)
		bodyText    = []byte(`Email body as plain text`)
		bodyHtml    = []byte(`Email body as <b>HTML</b>`)
		toAddresses = []byte("John <john@example.com>, Jane <jane@example.com>")

		recipients []string
		mboxes     []*email.Mailbox
		msg        *email.Message
		mailtx     *MailTx
		data       []byte
		err        error
	)

	rand.Seed(42)

	mboxes, err = email.ParseMailboxes(toAddresses)
	if err != nil {
		log.Fatal(err)
	}
	for _, mbox := range mboxes {
		recipients = append(recipients, mbox.Address)
	}

	msg, err = email.NewMultipart(
		fromAddress,
		toAddresses,
		subject,
		bodyText,
		bodyHtml,
	)
	if err != nil {
		log.Fatal(err)
	}

	// The From parameter is not necessary equal to the fromAddress.
	// The From in MailTx define the account that authorize or allowed
	// sending the email on behalf of fromAddress domain, while the
	// fromAddress define the address that viewed by recipients.
	mailtx = NewMailTx(txFrom, recipients, msg.Pack())

	fmt.Printf("Tx From: %s\n", mailtx.From)
	fmt.Printf("Tx Recipients: %s\n", mailtx.Recipients)

	// In order to make the example Output works, we need to replace all
	// CRLF with LF.
	data = bytes.ReplaceAll(mailtx.Data, []byte("\r\n"), []byte("\n"))

	fmt.Printf("Tx Data:\n%s", data)
	//Output:
	//Tx From: Postmaster <postmaster@mail.example.com>
	//Tx Recipients: [john@example.com jane@example.com]
	//Tx Data:
	//from: Noreply <noreply@example.com>
	//to: John <john@example.com>, Jane <jane@example.com>
	//subject: Example subject
	//mime-version: 1.0
	//content-type: multipart/alternative; boundary=1b4df158039f7cce49f0a64b0ea7b7dd
	//
	//--1b4df158039f7cce49f0a64b0ea7b7dd
	//content-type: text/plain; charset="utf-8"
	//mime-version: 1.0
	//content-transfer-encoding: quoted-printable
	//
	//Email body as plain text
	//--1b4df158039f7cce49f0a64b0ea7b7dd
	//content-type: text/html; charset="utf-8"
	//mime-version: 1.0
	//content-transfer-encoding: quoted-printable
	//
	//Email body as <b>HTML</b>
	//--1b4df158039f7cce49f0a64b0ea7b7dd--
}
