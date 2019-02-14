// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io/ioutil"
	"log"
	"strings"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/email/dkim"
)

//
// Message represent an unpacked internet message format.
//
type Message struct {
	Header        *Header
	Body          *Body
	DKIMSignature *dkim.Signature
	dkimStatus    *dkim.Status
	hasher        hash.Hash
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

	msg = &Message{}

	msg.Header, rest, err = ParseHeader(raw)
	if err != nil {
		return nil, rest, err
	}

	boundary := msg.Header.Boundary()

	msg.Body, rest, err = ParseBody(rest, boundary)

	return msg, rest, err
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
		msg.dkimStatus.SDID = *libbytes.Copy(sig.SDID)
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

	msg.createHasher()

	_, err = msg.dkimVerifyBody()
	if err != nil {
		msg.dkimStatus.Type = dkim.StatusPermFail
		msg.dkimStatus.Error = err
		msg.hasher = nil
		return nil, err
	}

	msg.hasher.Reset()

	msg.dkimHashHeaders(subHeader)

	hashed := msg.hasher.Sum(nil)

	err = sig.Verify(key, hashed)
	if err != nil {
		msg.dkimStatus.Type = dkim.StatusPermFail
		msg.dkimStatus.Error = err
		msg.hasher = nil
		return msg.dkimStatus, err
	}

	msg.dkimStatus.Type = dkim.StatusOK
	msg.hasher = nil

	return msg.dkimStatus, nil
}

//
// String return the text representation of Message object.
//
func (msg *Message) String() string {
	var sb strings.Builder

	if msg.Header != nil {
		sb.WriteString(msg.Header.String())
	}
	sb.WriteByte(cr)
	sb.WriteByte(lf)
	if msg.Body != nil {
		sb.WriteString(msg.Body.String())
	}

	return sb.String()
}

func (msg *Message) createHasher() {
	switch *msg.DKIMSignature.Alg {
	case dkim.SignAlgRS1:
		msg.hasher = sha1.New()
	case dkim.SignAlgRS256:
		msg.hasher = sha256.New()
	}
}

func (msg *Message) dkimVerifyBody() (h []byte, err error) {
	var body []byte

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
	if len(body) == 0 {
		return
	}

	msg.hasher.Write(body)

	h = msg.hasher.Sum(nil)
	bodyHash := make([]byte, base64.StdEncoding.EncodedLen(len(h)))
	base64.StdEncoding.Encode(bodyHash, h)

	if !bytes.Equal(msg.DKIMSignature.BodyHash, bodyHash) {
		err = fmt.Errorf("email: body hash did not verify")
		return nil, err
	}

	return h, nil
}

//
// dkimHashHeaders compute hash for each header om "h=" DKIMSignature.Headers
// followed by the canonicalization of DKIM-Signature itself.
//
func (msg *Message) dkimHashHeaders(subHeader *Header) {
	for x := 0; x < len(msg.DKIMSignature.Headers); x++ {
		signedField := subHeader.popByName(msg.DKIMSignature.Headers[x])
		if signedField == nil {
			log.Printf("email: dkimHashHeaders: field '%s' not found\n",
				msg.DKIMSignature.Headers[x])
			continue
		}

		msg.dkimHashField(signedField)
	}

	// The last one to hash is DKIM-Signature itself without "b=" value
	// and CRLF.
	var canonDKIM []byte
	if msg.DKIMSignature.CanonHeader == nil || *msg.DKIMSignature.CanonHeader == dkim.CanonSimple {
		canonDKIM = append(canonDKIM, subHeader.fields[0].oriName...)
		canonDKIM = append(canonDKIM, ':')
		v := dkim.Canonicalize(subHeader.fields[0].oriValue)
		canonDKIM = append(canonDKIM, v...)
	} else {
		canonDKIM = append(canonDKIM, subHeader.fields[0].Name...)
		canonDKIM = append(canonDKIM, ':')
		v := dkim.Canonicalize(subHeader.fields[0].Value)
		canonDKIM = append(canonDKIM, v...)
	}

	msg.hasher.Write(canonDKIM)
}

func (msg *Message) dkimHashField(f *Field) {
	var fb []byte
	if msg.DKIMSignature.CanonHeader == nil || *msg.DKIMSignature.CanonHeader == dkim.CanonSimple {
		fb = f.Simple()
	} else {
		fb = f.Relaxed()
	}
	msg.hasher.Write(fb)
}
