// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

// FieldType represent numerical identification of field name.
type FieldType int

// List of valid email envelope headers.
const (
	FieldTypeOptional FieldType = iota
	// The origination date field, RFC 5322 section 3.6.1.
	FieldTypeDate
	// Originator fields, RFC 5322 section 3.6.2.
	FieldTypeFrom
	FieldTypeSender
	FieldTypeReplyTo
	// Destination fields, RFC 5322 section 3.6.3.
	FieldTypeTo
	FieldTypeCC
	FieldTypeBCC
	// Identitication fields, RFC 5322 section 3.6.4.
	FieldTypeMessageID
	FieldTypeInReplyTo
	FieldTypeReferences
	// Informational fields, RFC 5322 section 3.6.5.
	FieldTypeSubject
	FieldTypeComments
	FieldTypeKeywords
	// Resent fields, RFC 5322 section 3.6.6.
	FieldTypeResentDate
	FieldTypeResentFrom
	FieldTypeResentSender
	FieldTypeResentTo
	FieldTypeResentCC
	FieldTypeResentBCC
	FieldTypeResentMessageID
	// Trace fields, RFC 5322 section 3.6.7.
	FieldTypeReturnPath
	FieldTypeReceived

	// MIME header fields, RFC 2045
	FieldTypeMIMEVersion
	FieldTypeContentType
	FieldTypeContentTransferEncoding
	FieldTypeContentID
	FieldTypeContentDescription

	// DKIM Signature, RFC 6376.
	FieldTypeDKIMSignature
)

// fieldNames contains mapping between field type and their lowercase name.
var fieldNames = map[FieldType]string{
	FieldTypeDate: `date`,

	FieldTypeFrom:    `from`,
	FieldTypeSender:  `sender`,
	FieldTypeReplyTo: `reply-to`,

	FieldTypeTo:  `to`,
	FieldTypeCC:  `cc`,
	FieldTypeBCC: `bcc`,

	FieldTypeMessageID:  `message-id`,
	FieldTypeInReplyTo:  `in-reply-to`,
	FieldTypeReferences: `references`,

	FieldTypeSubject:  `subject`,
	FieldTypeComments: `comments`,
	FieldTypeKeywords: `keywords`,

	FieldTypeResentDate:      `resent-date`,
	FieldTypeResentFrom:      `resent-from`,
	FieldTypeResentSender:    `resent-sender`,
	FieldTypeResentTo:        `resent-to`,
	FieldTypeResentCC:        `resent-cc`,
	FieldTypeResentBCC:       `resent-bcc`,
	FieldTypeResentMessageID: `resent-message-id`,

	FieldTypeReturnPath: `return-path`,
	FieldTypeReceived:   `received`,

	FieldTypeMIMEVersion:             `mime-version`,
	FieldTypeContentType:             `content-type`,
	FieldTypeContentTransferEncoding: `content-transfer-encoding`,
	FieldTypeContentID:               `content-id`,
	FieldTypeContentDescription:      `content-description`,

	FieldTypeDKIMSignature: `dkim-signature`,
}
