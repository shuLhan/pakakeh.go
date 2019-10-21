// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

//
// FieldType represent numerical identification of field name.
//
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

//
// fieldNames contains mapping between field type and their lowercase name.
//
//nolint:gochecknoglobals
var fieldNames = map[FieldType][]byte{
	FieldTypeDate: []byte("date"),

	FieldTypeFrom:    []byte("from"),
	FieldTypeSender:  []byte("sender"),
	FieldTypeReplyTo: []byte("reply-to"),

	FieldTypeTo:  []byte("to"),
	FieldTypeCC:  []byte("cc"),
	FieldTypeBCC: []byte("bcc"),

	FieldTypeMessageID:  []byte("message-id"),
	FieldTypeInReplyTo:  []byte("in-reply-to"),
	FieldTypeReferences: []byte("references"),

	FieldTypeSubject:  []byte("subject"),
	FieldTypeComments: []byte("comments"),
	FieldTypeKeywords: []byte("keywords"),

	FieldTypeResentDate:      []byte("resent-date"),
	FieldTypeResentFrom:      []byte("resent-from"),
	FieldTypeResentSender:    []byte("resent-sender"),
	FieldTypeResentTo:        []byte("resent-to"),
	FieldTypeResentCC:        []byte("resent-cc"),
	FieldTypeResentBCC:       []byte("resent-bcc"),
	FieldTypeResentMessageID: []byte("resent-message-id"),

	FieldTypeReturnPath: []byte("return-path"),
	FieldTypeReceived:   []byte("received"),

	FieldTypeMIMEVersion:             []byte("mime-version"),
	FieldTypeContentType:             []byte("content-type"),
	FieldTypeContentTransferEncoding: []byte("content-transfer-encoding"),
	FieldTypeContentID:               []byte("content-id"),
	FieldTypeContentDescription:      []byte("content-description"),

	FieldTypeDKIMSignature: []byte("dkim-signature"),
}
