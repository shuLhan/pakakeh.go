// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dkim

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

// Signature represents the value of DKIM-Signature header field tag.
type Signature struct {
	// Algorithm used to generate the signature.
	// Valid values is "rsa-sha1" or "rsa-sha256".
	// ("a=", text, REQUIRED).
	Alg *SignAlg

	// Type of canonicalization for header.  Default is "simple".
	// ("c=header/body", text, OPTIONAL).
	CanonHeader *Canon

	// Type of canonicalization for body.  Default is "simple".
	// ("c=header/body", text, OPTIONAL).
	CanonBody *Canon

	// The number of octets in body, after canonicalization, included when
	// computing hash.
	// If nil, its means entire body is signed.
	// If 0, its means the body is not signed.
	// ("l=", text, OPTIONAL).
	BodyLength *uint64

	// QMethod define a type and option used to retrieve the public keys.
	// ("q=type/option", text, OPTIONAL).  Default is "dns/txt".
	QMethod *QueryMethod

	// Version of specification.
	// It MUST have the value "1" for compliant with RFC 6376.
	// ("v=", text, REQUIRED).
	Version []byte

	// Signer domain
	// ("d=", text, REQUIRED).
	SDID []byte

	// The selector subdividing the namespace for the "d=" tag.
	// ("s=", text, REQUIRED).
	Selector []byte

	// List of header field names included in signing or verifying.
	// ("h=", text, REQUIRED).
	Headers [][]byte

	// The hash of canonicalized body data.
	// ("bh=", base64, REQUIRED).
	BodyHash []byte

	// The signature data.
	// ("b=", base64, REQUIRED)
	Value []byte

	// List of header field name and value that present when the message
	// is signed.
	// ("z=", dkim-quoted-printable, OPTIONAL).  Default is null.
	PresentHeaders [][]byte

	// The agent or user identifier.
	// Default is "@" + "d=" value)
	// ("i=", dkim-quoted-printable, OPTIONAL).
	AUID []byte

	// raw contains original Signature field value, for Simple
	// canonicalization.
	raw []byte

	// Time when signature created, in UNIX timestamp.
	// ("t=", text, RECOMMENDED).
	CreatedAt uint64

	// Expiration time, in UNIX timestamp.
	// ("x=", text, RECOMMENDED).
	ExpiredAt uint64
}

// Parse DKIM-Signature field value.
// The signature value MUST be end with CRLF.
func Parse(value []byte) (sig *Signature, err error) {
	if len(value) == 0 {
		return nil, nil
	}

	var l = len(value)
	if value[l-2] != '\r' && value[l-1] != '\n' {
		return nil, errors.New("dkim: value must end with CRLF")
	}

	var (
		p = newParser(value)

		tag *tag
	)

	sig = &Signature{}
	for {
		tag, err = p.fetchTag()
		if err != nil {
			return sig, err
		}
		if tag == nil {
			break
		}
		if tag.key == tagUnknown {
			continue
		}
		err = sig.set(tag)
		if err != nil {
			return sig, err
		}
	}

	sig.raw = value

	return sig, nil
}

// NewSignature create and initialize new signature using SDID ("d=") and
// selector ("s=") and default value for the rest of field.
func NewSignature(sdid, selector []byte) (sig *Signature) {
	sig = &Signature{
		SDID:     sdid,
		Selector: selector,
	}

	sig.SetDefault()

	return sig
}

// Hash compute the hash of input using the defined signature algorithm and
// return their binary and base64 representation.
func (sig *Signature) Hash(in []byte) (h, h64 []byte) {
	if sig.Alg == nil || *sig.Alg == SignAlgRS256 {
		h256 := sha256.Sum256(in)
		h = h256[:]
	} else {
		h1 := sha1.Sum(in)
		h = h1[:]
	}

	h64 = make([]byte, base64.StdEncoding.EncodedLen(len(h)))
	base64.StdEncoding.Encode(h64, h)

	return
}

// Pack the Signature into stream.  Each non empty tag field is printed,
// ordered by tag priority: required, recommended, and optional.
// Recommended and optional field values will be printed only if its not
// empty.
func (sig *Signature) Pack(simple bool) []byte {
	bb := new(bytes.Buffer)
	var sigAlg = signAlgNames[SignAlgRS256]

	if sig.Alg != nil {
		sigAlg = signAlgNames[*sig.Alg]
	}

	_, _ = fmt.Fprintf(bb, "v=%s; a=%s; d=%s; s=%s;",
		sig.Version, sigAlg, sig.SDID, sig.Selector)
	wrap(bb, simple)

	_, _ = fmt.Fprintf(bb, "h=%s;", bytes.Join(sig.Headers, sepColon))
	wrap(bb, simple)
	_, _ = fmt.Fprintf(bb, "bh=%s;", sig.BodyHash)
	wrap(bb, simple)
	_, _ = fmt.Fprintf(bb, "b=%s;", sig.Value)
	wrap(bb, simple)

	if sig.CreatedAt > 0 {
		_, _ = fmt.Fprintf(bb, "t=%d; ", sig.CreatedAt)
	}
	if sig.ExpiredAt > 0 {
		_, _ = fmt.Fprintf(bb, "x=%d; ", sig.ExpiredAt)
	}

	if sig.CanonHeader != nil {
		_, _ = fmt.Fprintf(bb, "c=%s", canonNames[*sig.CanonHeader])

		if sig.CanonBody != nil {
			_, _ = fmt.Fprintf(bb, "/%s;", canonNames[*sig.CanonBody])
		} else {
			bb.WriteByte(';')
		}
		wrap(bb, simple)
	}

	if len(sig.PresentHeaders) > 0 {
		bb.WriteString("z=")
		for x := range len(sig.PresentHeaders) {
			if x > 0 {
				bb.WriteByte('|')
				wrap(bb, simple)
			}
			bb.Write(sig.PresentHeaders[x])
		}
		bb.WriteByte(';')
		wrap(bb, simple)
	}

	if len(sig.AUID) > 0 {
		_, _ = fmt.Fprintf(bb, "i=%s; ", sig.AUID)
	}
	if sig.BodyLength != nil {
		_, _ = fmt.Fprintf(bb, "l=%d; ", *sig.BodyLength)
	}
	if sig.QMethod != nil {
		_, _ = fmt.Fprintf(bb, "q=%s/%s",
			queryTypeNames[sig.QMethod.Type],
			queryOptionNames[sig.QMethod.Option])
	}
	bb.WriteByte('\r')
	bb.WriteByte('\n')

	return bb.Bytes()
}

func wrap(bb *bytes.Buffer, simple bool) {
	if simple {
		bb.WriteByte('\r')
		bb.WriteByte('\n')
	}
	bb.WriteByte(' ')
}

// SetDefault signature field's values.
//
// The default values are "sha-rsa256" for signing algorithm, and
// "relaxed/relaxed" for canonicalization in header and body.
func (sig *Signature) SetDefault() {
	if len(sig.Version) == 0 {
		sig.Version = append(sig.Version, '1')
	}
	if sig.Alg == nil {
		signAlg := SignAlgRS256
		sig.Alg = &signAlg
	}
	if sig.CanonHeader == nil {
		canonHeader := CanonRelaxed
		sig.CanonHeader = &canonHeader
	}
	if sig.CanonBody == nil {
		canonBody := CanonRelaxed
		sig.CanonBody = &canonBody
	}
}

// Sign compute the signature of message hash header using specific private
// key and store the base64 result in Signature.Value ("b=").
func (sig *Signature) Sign(pk *rsa.PrivateKey, hashHeader []byte) (err error) {
	if pk == nil {
		return errors.New(`email/dkim: empty private key for signing`)
	}

	cryptoHash := crypto.SHA256
	if sig.Alg != nil && *sig.Alg == SignAlgRS1 {
		cryptoHash = crypto.SHA1
	}

	rng := rand.Reader
	b, err := rsa.SignPKCS1v15(rng, pk, cryptoHash, hashHeader)
	if err != nil {
		err = fmt.Errorf("email/dkim: failed to sign message: %w", err)
		return err
	}

	sig.Value = make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(sig.Value, b)

	return nil
}

// Relaxed return the "relaxed" canonicalization of Signature.
func (sig *Signature) Relaxed() []byte {
	return sig.Pack(false)
}

// Simple return the "simple" canonicalization of Signature.
func (sig *Signature) Simple() []byte {
	if len(sig.raw) == 0 {
		return sig.Pack(true)
	}
	return sig.raw
}

// Validate the signature's tag values.
//
// Rules of tags,
//
// *  Tags MUST NOT duplicate.  This was already handled by parser.
//
// *  All the required fields MUST NOT be empty or nil.
//
// *  The "h=" tag MUST include the "From" header field
//
// *  The value of the "x=" tag MUST be greater than the value of the "t=" tag
// if both are present.
//
// *  The "d=" value MUST be the same or parent domain of "i="
func (sig *Signature) Validate() (err error) {
	if len(sig.Version) == 0 || sig.Version[0] != '1' {
		return fmt.Errorf("dkim: invalid version: '%s'", sig.Version)
	}
	if sig.Alg == nil {
		return errEmptySignAlg
	}
	if len(sig.SDID) == 0 {
		return errEmptySDID
	}
	if len(sig.Selector) == 0 {
		return errEmptySelector
	}
	if len(sig.Headers) == 0 {
		return errEmptyHeader
	}

	err = sig.validateHeaders()
	if err != nil {
		return err
	}

	if len(sig.BodyHash) == 0 {
		return errEmptyBodyHash
	}
	if len(sig.Value) == 0 {
		return errEmptySignature
	}

	err = sig.validateTime()
	if err != nil {
		return err
	}

	err = sig.validateAUID()

	return err
}

// Verify the signature value ("b=") using DKIM public key record and computed
// hash of message header.
func (sig *Signature) Verify(key *Key, headerHash []byte) (err error) {
	if key == nil {
		return errors.New(`email/dkim: key record is empty`)
	}
	if key.RSA == nil {
		return errors.New(`email/dkim: public key is empty`)
	}

	sigValue := make([]byte, base64.StdEncoding.DecodedLen(len(sig.Value)))
	n, err := base64.StdEncoding.Decode(sigValue, sig.Value)
	if err != nil {
		return fmt.Errorf("email/dkim: failed to decode signature: %w", err)
	}
	sigValue = sigValue[:n]

	cryptoHash := crypto.SHA256
	if sig.Alg != nil && *sig.Alg == SignAlgRS1 {
		cryptoHash = crypto.SHA1
	}

	err = rsa.VerifyPKCS1v15(key.RSA, cryptoHash, headerHash, sigValue)
	if err != nil {
		err = fmt.Errorf("email/dkim: verification failed: %w", err)
	}

	return err
}

// set the signature field value with value from tag.
func (sig *Signature) set(t *tag) (err error) {
	if t == nil {
		return
	}

	var l uint64

	switch t.key {
	case tagVersion:
		if len(t.value) != 1 || t.value[0] != '1' {
			return fmt.Errorf("dkim: invalid version: '%s'", t.value)
		}
		sig.Version = t.value

	case tagAlg:
		for k, name := range signAlgNames {
			if bytes.Equal(t.value, name) {
				k := k
				sig.Alg = &k
				return nil
			}
		}
		return fmt.Errorf("dkim: unknown algorithm: '%s'", t.value)

	case tagSDID:
		if len(t.value) == 0 {
			return errEmptySDID
		}
		sig.SDID = t.value

	case tagHeaders:
		if len(t.value) == 0 {
			return errEmptyHeader
		}
		headers := bytes.Split(t.value, sepColon)
		for x := range len(headers) {
			headers[x] = bytes.ToLower(bytes.TrimSpace(headers[x]))
			sig.Headers = append(sig.Headers, headers[x])
		}
		err = sig.validateHeaders()

	case tagSelector:
		if len(t.value) == 0 {
			return errEmptySelector
		}
		sig.Selector = t.value

	case tagBodyHash:
		if len(t.value) == 0 {
			return errEmptyBodyHash
		}
		sig.BodyHash = t.value

	case tagSignature:
		if len(t.value) == 0 {
			return errEmptySignature
		}
		sig.Value = t.value

	case tagCreatedAt:
		sig.CreatedAt, err = strconv.ParseUint(string(t.value), 10, 64)
		if err != nil {
			return errors.New("dkim: t=: " + err.Error())
		}
		err = sig.validateTime()

	case tagExpiredAt:
		sig.ExpiredAt, err = strconv.ParseUint(string(t.value), 10, 64)
		if err != nil {
			return errors.New("dkim: x=: " + err.Error())
		}
		err = sig.validateTime()
	case tagCanon:
		sig.CanonHeader, sig.CanonBody, err = unpackCanons(t.value)

	case tagPresentHeaders:
		z := bytes.Split(t.value, sepVBar)
		for x := range len(z) {
			z[x] = bytes.TrimSpace(z[x])
			sig.PresentHeaders = append(sig.PresentHeaders, z[x])
		}

	case tagAUID:
		sig.AUID = t.value
		err = sig.validateAUID()

	case tagBodyLength:
		l, err = strconv.ParseUint(string(t.value), 10, 64)
		if err == nil {
			sig.BodyLength = &l
		}

	case tagQueryMethod:
		sig.setQueryMethods(t.value)
	}

	return err
}

// setQueryMethods parse list of query methods and set Signature.QueryMethod
// based on first match.
func (sig *Signature) setQueryMethods(v []byte) {
	methods := bytes.Split(v, sepColon)

	for _, m := range methods {
		var qtype, qopt []byte

		kv := bytes.Split(m, sepSlash)
		switch len(kv) {
		case 0:
		case 1:
			qtype = kv[0]
		case 2:
			qtype = kv[0]
			qopt = kv[1]
		}
		err := sig.setQueryMethod(qtype, qopt)
		if err != nil {
			sig.QMethod = nil
			// Ignore error, use default query method.
		}
		qtype = nil
		qopt = nil
	}
}

// setQueryMethod set Signature query type and option.
func (sig *Signature) setQueryMethod(qtype, qopt []byte) (err error) {
	if len(qtype) == 0 {
		return nil
	}

	sig.QMethod = &QueryMethod{}

	found := false
	for k, typ := range queryTypeNames {
		if bytes.Equal(qtype, typ) {
			sig.QMethod.Type = k
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("dkim: unknown query type: '%s'", qtype)
	}
	if len(qopt) == 0 {
		return nil
	}

	found = false
	for k, opt := range queryOptionNames {
		if bytes.Equal(qopt, opt) {
			sig.QMethod.Option = k
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("dkim: unknown query option: '%s'", qopt)
	}

	return nil
}

// validateHeaders validate value of header tag "h=" that it MUST contains
// "from".
func (sig *Signature) validateHeaders() (err error) {
	for x := range len(sig.Headers) {
		if bytes.Equal(sig.Headers[x], []byte("from")) {
			return nil
		}
	}
	return errFromHeader
}

func (sig *Signature) validateTime() (err error) {
	if sig.ExpiredAt == 0 || sig.CreatedAt == 0 {
		return nil
	}
	if sig.ExpiredAt < sig.CreatedAt {
		return errCreatedTime
	}
	if sig.ExpiredAt > math.MaxInt64 {
		// According to RFC 6376,
		// "To avoid denial-of-service attacks, implementations MAY
		// consider any value longer than 12 digits to be
		// infinite.".
		sig.ExpiredAt = math.MaxInt64
		return nil
	}
	exp := time.Unix(int64(sig.ExpiredAt), 0)
	now := time.Now().Add(time.Hour * -1).Unix()
	if uint64(now) > sig.ExpiredAt {
		return fmt.Errorf("dkim: signature is expired at '%s'", exp.UTC())
	}

	return nil
}

func (sig *Signature) validateAUID() (err error) {
	if len(sig.AUID) == 0 {
		return nil
	}

	bb := bytes.Split(sig.AUID, []byte{'@'})
	if len(bb) != 2 {
		return fmt.Errorf("dkim: no domain in AUID 'i=' tag: '%s'", sig.AUID)
	}
	if !bytes.HasSuffix(bb[1], sig.SDID) {
		return fmt.Errorf("dkim: invalid AUID: '%s'", sig.AUID)
	}

	return nil
}
