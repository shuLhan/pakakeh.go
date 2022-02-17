// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

//
// MailTx define a mail transaction.
//
type MailTx struct {
	Postpone time.Time

	// Received contains the time when the message arrived on server.
	// This field is ignored in Client.Send().
	Received time.Time

	// ID of message.
	// This field is ignored in Client.Send().
	ID string

	// From contains originator address.
	// This field is required in Client.Send().
	From string

	// Recipients contains list of the destination address.
	// This field is required in Client.Send().
	Recipients []string

	// Data contains content of message.
	// This field is optional in Client.Send().
	Data []byte

	Retry int
}

//
// NewMailTx create and return new mail object.
//
func NewMailTx(from string, to []string, data []byte) (mail *MailTx) {
	mail = &MailTx{
		From:       from,
		Recipients: to,
		Received:   time.Now().Round(0),
	}

	mail.ID = strconv.FormatInt(mail.Received.UnixNano(), 10)
	mail.Data = make([]byte, len(data))
	copy(mail.Data, data)

	return
}

//
// Reset all mail attributes to their zero value.
//
func (mail *MailTx) Reset() {
	mail.ID = ""
	mail.From = ""
	mail.Recipients = nil
	mail.Data = nil
}

//
// isTerminated will return true if data is end with "\r\n.\r\n".
//
func (mail *MailTx) isTerminated() bool {
	l := len(mail.Data)
	if l < 5 {
		return false
	}
	return bytes.Equal(mail.Data[l-5:l], []byte{'\r', '\n', '.', '\r', '\n'})
}

//
// postpone the mail transaction.
//
func (mail *MailTx) postpone() {
	mail.Retry++
	mail.Postpone = mail.Received.Add(time.Duration(mail.Retry*30) * time.Minute)
}

func (mail *MailTx) isPostponed() bool {
	return mail.Postpone.After(time.Now())
}

//
// seal the mail envelope by inserting trace information into message content.
//
func (mail *MailTx) seal(clientDomain, clientAddress, localAddress string) {
	line := fmt.Sprintf("FROM %s (%s)\r\n\tBY %s WITH SMTP ID %s;\r\n\t%s",
		clientDomain, clientAddress, localAddress, mail.ID,
		mail.Received.Format(time.RFC1123Z))
	mail.Data = append([]byte(line), mail.Data...)
}
