// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/email/dkim"
)

//
// Message represent an unpacked internet message format.
//
type Message struct {
	DKIMSignature *dkim.Signature
	dkimStatus    *dkim.Status

	Header Header
	Body   Body
}

//
// NewMultipart create multipart email message with text and HTML bodies.
//
func NewMultipart(from, to, subject, bodyText, bodyHTML []byte) (
	msg *Message, err error,
) {
	var (
		logp      = "NewMultipart"
		timeNow   = time.Unix(Epoch(), 0)
		dateValue = timeNow.Format(DateFormat)
	)

	if dateInUtc {
		dateValue = timeNow.UTC().Format(DateFormat)
	}

	msg = &Message{}

	err = msg.Header.Set(FieldTypeDate, []byte(dateValue))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	err = msg.Header.Set(FieldTypeFrom, from)
	if err != nil {
		return nil, fmt.Errorf("email.NewMultipart: %w", err)
	}

	err = msg.Header.Set(FieldTypeTo, to)
	if err != nil {
		return nil, fmt.Errorf("email.NewMultipart: %w", err)
	}

	err = msg.Header.Set(FieldTypeSubject, subject)
	if err != nil {
		return nil, fmt.Errorf("email.NewMultipart: %w", err)
	}

	err = msg.Header.SetMultipart()
	if err != nil {
		return nil, fmt.Errorf("email.NewMultipart: %w", err)
	}

	if len(bodyText) > 0 {
		mimeText, err := newMIME([]byte(contentTypeTextPlain), bodyText)
		if err != nil {
			return nil, fmt.Errorf("email.NewMultipart: %w", err)
		}
		msg.Body.Add(mimeText)
	}
	if len(bodyHTML) > 0 {
		mimeHTML, err := newMIME([]byte(contentTypeTextHTML), bodyHTML)
		if err != nil {
			return nil, fmt.Errorf("email.NewMultipart: %w", err)
		}
		msg.Body.Add(mimeHTML)
	}

	return msg, nil
}

//
// ParseFile parse message from file.
//
func ParseFile(inFile string) (msg *Message, rest []byte, err error) {
	raw, err := ioutil.ReadFile(inFile)
	if err != nil {
		return nil, nil, fmt.Errorf("email: " + err.Error())
	}

	return ParseMessage(raw)
}

//
// ParseMessage parse the raw message header and body.
//
func ParseMessage(raw []byte) (msg *Message, rest []byte, err error) {
	if len(raw) == 0 {
		return nil, nil, nil
	}

	var (
		logp = "ParseMessage"

		hdr      *Header
		body     *Body
		boundary []byte
	)

	msg = &Message{}

	hdr, rest, err = ParseHeader(raw)
	if err != nil {
		return nil, rest, fmt.Errorf("%s: %w", logp, err)
	}

	boundary = hdr.Boundary()

	body, rest, err = ParseBody(rest, boundary)
	if err != nil {
		return nil, rest, fmt.Errorf("%s: %w", logp, err)
	}

	msg.Header = *hdr
	msg.Body = *body

	return msg, rest, nil
}

//
// AddCC add one or more recipients to the message header CC.
//
func (msg *Message) AddCC(mailboxes string) (err error) {
	err = msg.addMailboxes(FieldTypeCC, []byte(mailboxes))
	if err != nil {
		return fmt.Errorf("AddCC: %w", err)
	}
	return nil
}

//
// AddTo add one or more recipients to the mesage header To.
//
func (msg *Message) AddTo(mailboxes string) (err error) {
	err = msg.addMailboxes(FieldTypeTo, []byte(mailboxes))
	if err != nil {
		return fmt.Errorf("AddTo: %w", err)
	}
	return nil
}

func (msg *Message) addMailboxes(ft FieldType, mailboxes []byte) error {
	mailboxes = bytes.TrimSpace(mailboxes)
	if len(mailboxes) == 0 {
		return nil
	}
	return msg.Header.addMailboxes(ft, mailboxes)
}

//
// DKIMSign sign the message using the private key and signature.
// The only required fields in signature is SDID and Selector, any other
// required fields that are empty will be initialized with default values.
//
// Upon calling this function, any field values in header and body MUST be
// already encoded.
//
func (msg *Message) DKIMSign(pk *rsa.PrivateKey, sig *dkim.Signature) (err error) {
	if pk == nil {
		return fmt.Errorf("email: empty private key for signing")
	}
	if sig == nil {
		return fmt.Errorf("email: empty signature for signing")
	}

	sig.SetDefault()
	msg.setDKIMHeaders(sig)

	// Set the body hash and signature to dummy value, to enable
	// validating it.
	dummy := []byte{0}
	sig.BodyHash = dummy
	sig.Value = dummy

	err = sig.Validate()
	if err != nil {
		return err
	}

	// Reset the body hash and value back to nil.
	sig.BodyHash = nil
	sig.Value = nil
	msg.DKIMSignature = sig

	_, sig.BodyHash = sig.Hash(msg.CanonBody())

	dkimField := &Field{
		Type:     FieldTypeDKIMSignature,
		Name:     fieldNames[FieldTypeDKIMSignature],
		Value:    sig.Pack(false),
		oriName:  []byte("DKIM-Signature"),
		oriValue: sig.Pack(true),
	}

	subHeader := &Header{
		fields: make([]*Field, len(msg.Header.fields)),
	}
	copy(subHeader.fields, msg.Header.fields)

	hh, _ := sig.Hash(msg.CanonHeader(subHeader, dkimField))

	err = sig.Sign(pk, hh)
	if err != nil {
		return err
	}

	// Regenerate the DKIM field again with non empty signature "b=".
	dkimField.Value = sig.Pack(false)
	dkimField.oriValue = sig.Pack(true)

	msg.Header.PushTop(dkimField)

	return nil
}

//
// DKIMVerify verify the message signature using DKIM.
//
func (msg *Message) DKIMVerify() (*dkim.Status, error) {
	// Do not run verify again if the message has no DKIM-Signature or
	// already permanent failed.
	if msg.dkimStatus != nil {
		switch msg.dkimStatus.Type {
		case dkim.StatusNoSignature, dkim.StatusOK, dkim.StatusPermFail:
			return msg.dkimStatus, msg.dkimStatus.Error
		}
	}

	msg.dkimStatus = &dkim.Status{}

	// Only process the first DKIM-Signature for now.
	subHeader := msg.Header.DKIM(1)
	if subHeader == nil || len(subHeader.fields) == 0 {
		msg.dkimStatus.Type = dkim.StatusNoSignature
		return msg.dkimStatus, nil
	}

	sig, err := dkim.Parse(subHeader.fields[0].Value)

	if sig != nil && len(sig.SDID) > 0 {
		msg.dkimStatus.SDID = libbytes.Copy(sig.SDID)
	}
	if err != nil {
		msg.dkimStatus.Type = dkim.StatusPermFail
		msg.dkimStatus.Error = err
		return msg.dkimStatus, err
	}

	err = sig.Validate()
	if err != nil {
		msg.dkimStatus.Type = dkim.StatusPermFail
		msg.dkimStatus.Error = err
		return msg.dkimStatus, err
	}

	// Check if the headers really contains "from:" field.
	from := subHeader.Filter(FieldTypeFrom)
	if len(from) == 0 {
		msg.dkimStatus.Type = dkim.StatusPermFail
		msg.dkimStatus.Error = fmt.Errorf("email: missing 'From' field")
		return msg.dkimStatus, msg.dkimStatus.Error
	}

	// Get the public key.
	dname := fmt.Sprintf("%s._domainkey.%s", sig.Selector, sig.SDID)
	key, err := dkim.DefaultKeyPool.Get(dname)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			msg.dkimStatus.Type = dkim.StatusTempFail
		} else {
			msg.dkimStatus.Type = dkim.StatusPermFail
		}
		msg.dkimStatus.Error = err
		return msg.dkimStatus, err
	}

	msg.DKIMSignature = sig

	canonBody := msg.CanonBody()
	_, bh64 := sig.Hash(canonBody)

	if !bytes.Equal(sig.BodyHash, bh64) {
		err = fmt.Errorf("email: body hash did not verify")
		msg.dkimStatus.Type = dkim.StatusPermFail
		msg.dkimStatus.Error = err
		return nil, err
	}

	canonHeader := msg.CanonHeader(subHeader, subHeader.fields[0])
	hh, _ := sig.Hash(canonHeader)

	err = sig.Verify(key, hh)
	if err != nil {
		msg.dkimStatus.Type = dkim.StatusPermFail
		msg.dkimStatus.Error = err
		return msg.dkimStatus, err
	}

	msg.dkimStatus.Type = dkim.StatusOK

	return msg.dkimStatus, nil
}

//
// SetBodyHtml set or replace the message's body HTML content.
//
func (msg *Message) SetBodyHtml(content []byte) (err error) {
	err = msg.setBody([]byte(contentTypeTextHTML), content)
	if err != nil {
		return fmt.Errorf("SetBodyHtml: %w", err)
	}
	return nil
}

//
// SetBodyText set or replace the message body text content.
//
func (msg *Message) SetBodyText(content []byte) (err error) {
	err = msg.setBody([]byte(contentTypeTextPlain), content)
	if err != nil {
		return fmt.Errorf("SetBodyText: %w", err)
	}
	return nil
}

func (msg *Message) setBody(contentType, content []byte) (err error) {
	var (
		mime *MIME
	)
	mime, err = newMIME(contentType, content)
	if err != nil {
		return err
	}
	msg.Body.Set(mime)
	return nil
}

//
// SetCC set or replace the message header CC with one or more mailboxes.
// See AddCC to add another recipient to the CC header.
//
func (msg *Message) SetCC(mailboxes string) (err error) {
	err = msg.setMailboxes(FieldTypeCC, []byte(mailboxes))
	if err != nil {
		return fmt.Errorf("SetCC: %w", err)
	}
	return nil
}

//
// SetFrom set or replace the message header From with mailbox.
// If the mailbox parameter is empty, nothing will changes.
//
func (msg *Message) SetFrom(mailbox string) (err error) {
	err = msg.setMailboxes(FieldTypeFrom, []byte(mailbox))
	if err != nil {
		return fmt.Errorf("SetFrom: %w", err)
	}
	return nil
}

//
// SetSubject set or replace the subject.
// It will do nothing if the subject is empty.
//
func (msg *Message) SetSubject(subject string) {
	subject = strings.TrimSpace(subject)
	if len(subject) == 0 {
		return
	}
	_ = msg.Header.Set(FieldTypeSubject, []byte(subject))
}

//
// SetTo set or replace the message header To with one or more mailboxes.
// See AddTo to add another recipient to the To header.
//
func (msg *Message) SetTo(mailboxes string) (err error) {
	err = msg.setMailboxes(FieldTypeTo, []byte(mailboxes))
	if err != nil {
		return fmt.Errorf("SetTo: %w", err)
	}
	return nil
}

func (msg *Message) setMailboxes(ft FieldType, mailboxes []byte) error {
	mailboxes = bytes.TrimSpace(mailboxes)
	if len(mailboxes) == 0 {
		return nil
	}
	return msg.Header.Set(ft, mailboxes)
}

//
// String return the text representation of Message object.
//
func (msg *Message) String() string {
	var sb strings.Builder

	sb.Write(msg.Header.Relaxed())
	sb.WriteByte(cr)
	sb.WriteByte(lf)
	sb.WriteString(msg.Body.String())

	return sb.String()
}

//
// CanonBody return the canonical representation of Message.
//
func (msg *Message) CanonBody() (body []byte) {
	if msg.DKIMSignature.CanonBody == nil || *msg.DKIMSignature.CanonBody == dkim.CanonSimple {
		body = msg.Body.Simple()
	} else {
		body = msg.Body.Relaxed()
	}

	switch {
	case msg.DKIMSignature.BodyLength == nil:
		// Hash entire body ...
	case *msg.DKIMSignature.BodyLength == 0:
		// Body is not hashed.
		body = nil
	default:
		body = body[:*msg.DKIMSignature.BodyLength]
	}

	return
}

//
// CanonHeader generate the canonicalization of sub-header and DKIM-Signature
// field.
//
func (msg *Message) CanonHeader(subHeader *Header, dkimField *Field) []byte {
	var bb bytes.Buffer

	canonType := dkim.CanonRelaxed
	if msg.DKIMSignature.CanonHeader == nil || *msg.DKIMSignature.CanonHeader == dkim.CanonSimple {
		canonType = dkim.CanonSimple
	}

	for x := 0; x < len(msg.DKIMSignature.Headers); x++ {
		signedField := subHeader.popByName(msg.DKIMSignature.Headers[x])
		if signedField == nil {
			continue
		}
		if canonType == dkim.CanonSimple {
			bb.Write(signedField.Simple())
		} else {
			bb.Write(signedField.Relaxed())
		}
	}

	// The last one to hash is DKIM-Signature itself without "b=" value
	// and CRLF.
	if canonType == dkim.CanonSimple {
		bb.Write(dkimField.oriName)
		bb.WriteByte(':')
		bb.Write(dkim.Canonicalize(dkimField.oriValue))
	} else {
		bb.Write(dkimField.Name)
		bb.WriteByte(':')
		bb.Write(dkim.Canonicalize(dkimField.Value))
	}

	return bb.Bytes()
}

//
// Pack the message for sending.
//
func (msg *Message) Pack() (out []byte) {
	var buf bytes.Buffer

	boundary := msg.Header.Boundary()

	for _, f := range msg.Header.fields {
		if f.Type == FieldTypeContentType {
			fmt.Fprintf(&buf, "%s: %s\r\n", f.Name, f.ContentType.String())
		} else {
			fmt.Fprintf(&buf, "%s: %s", f.Name, f.Value)
		}
	}
	buf.WriteString("\r\n")
	for _, mime := range msg.Body.Parts {
		if len(boundary) > 0 {
			fmt.Fprintf(&buf, "--%s\r\n", boundary)
		}
		for _, f := range mime.Header.fields {
			if f.Type == FieldTypeContentType {
				fmt.Fprintf(&buf, "%s: %s\r\n", f.Name,
					f.ContentType.String())
			} else {
				fmt.Fprintf(&buf, "%s: %s", f.Name, f.Value)
			}
		}
		buf.WriteString("\r\n")
		buf.Write(mime.Content)
	}
	if len(boundary) > 0 {
		fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	}
	return buf.Bytes()
}

//
// setDKIMHeaders set the DKIM signature headers ("h=") value with current
// list of headers name in message.
//
func (msg *Message) setDKIMHeaders(sig *dkim.Signature) {
	if len(sig.Headers) > 0 {
		return
	}

	sig.Headers = make([][]byte, 0, len(msg.Header.fields))

	for _, f := range msg.Header.fields {
		sig.Headers = append(sig.Headers, f.Name)
	}
}
