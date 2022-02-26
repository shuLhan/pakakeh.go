// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"time"

	"github.com/shuLhan/share/lib/email"
)

func ExampleNewMailTx() {
	// Example on how to create MailTx Data using email package [1].
	//
	// [1] github.com/shuLhan/share/lib/email

	// Overwrite the email.Epoch to make the example works.
	email.Epoch = func() int64 {
		return 1645600000
	}

	var (
		txFrom      = "Postmaster <postmaster@mail.example.com>"
		fromAddress = []byte("Noreply <noreply@example.com>")
		toAddresses = []byte("John <john@example.com>, Jane <jane@example.com>")
		subject     = []byte(`Example subject`)
		bodyText    = []byte(`Email body as plain text`)
		bodyHtml    = []byte(`Email body as <b>HTML</b>`)
		timeNowUtc  = time.Unix(email.Epoch(), 0).UTC()
		dateNowUtc  = timeNowUtc.Format(email.DateFormat)

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
	data, _ = msg.Pack()
	mailtx = NewMailTx(txFrom, recipients, data)

	fmt.Printf("Tx From: %s\n", mailtx.From)
	fmt.Printf("Tx Recipients: %s\n", mailtx.Recipients)

	// In order to make the example Output works, we need to replace all
	// CRLF with LF and "date:" with the system timezone.
	data = bytes.ReplaceAll(mailtx.Data, []byte("\r\n"), []byte("\n"))

	//fmt.Printf("timeNowUtc: %s\n", timeNowUtc)
	//fmt.Printf("dateNowUtc: %s\n", dateNowUtc)

	reDate := regexp.MustCompile(`^date: Wed(.*) \+....`)
	data = reDate.ReplaceAll(data, []byte(`date: `+dateNowUtc))

	fmt.Printf("Tx Data:\n%s", data)
	//Output:
	//Tx From: Postmaster <postmaster@mail.example.com>
	//Tx Recipients: [john@example.com jane@example.com]
	//Tx Data:
	//date: Wed, 23 Feb 2022 07:06:40 +0000
	//from: Noreply <noreply@example.com>
	//to: John <john@example.com>, Jane <jane@example.com>
	//subject: Example subject
	//mime-version: 1.0
	//content-type: multipart/alternative; boundary=1b4df158039f7cce49f0a64b0ea7b7dd
	//
	//--1b4df158039f7cce49f0a64b0ea7b7dd
	//mime-version: 1.0
	//content-type: text/plain; charset="utf-8"
	//content-transfer-encoding: quoted-printable
	//
	//Email body as plain text
	//--1b4df158039f7cce49f0a64b0ea7b7dd
	//mime-version: 1.0
	//content-type: text/html; charset="utf-8"
	//content-transfer-encoding: quoted-printable
	//
	//Email body as <b>HTML</b>
	//--1b4df158039f7cce49f0a64b0ea7b7dd--
}
