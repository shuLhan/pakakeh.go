// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// Key represent a DKIM key record.
type Key struct {
	// REQUIRED fields.

	// Public contains public key data.
	// ("p=", base64, REQUIRED)
	Public []byte

	// RECOMMENDED fields.

	// Version of DKIM key record.
	// ("v=", plain-text, RECOMMENDED, default is "DKIM1")
	Version []byte

	// OPTIONAL fields.

	// Type of key.
	// ("k=", plain-text, OPTIONAL, default is "rsa").
	Type *KeyType

	// HashAlgs contains list of hash algorithm that might be used.
	// ("h=", plain-text colon-separated,  OPTIONAL, defaults to allowing
	// all algorithms)
	HashAlgs []HashAlg

	// Notes that might be interest to human.
	// ("n=", qp-section, OPTIONAL, default is empty)
	Notes []byte

	// Services contains list of service types to which this record
	// applies.
	// ("s=", plain-text colon-separated, OPTIONAL, default is "*").
	Services [][]byte

	// Flags contains list of flags.
	// ("t=", plain-text colon-separated, OPTIONAL, default is no flags set)
	Flags []KeyFlag

	// RSA contains parsed Public key.
	RSA *rsa.PublicKey

	// ExpiredAt define time when the key will be expired.
	// This is a local value derived from lookup time + RR TTL.
	ExpiredAt int64
}

// LookupKey DKIM (public) key using specific query method and DKIM domain
// name (selector plus SDID).
func LookupKey(qmethod QueryMethod, dname string) (key *Key, err error) {
	if len(dname) == 0 {
		return nil, nil
	}
	if qmethod.Type == QueryTypeDNS {
		key, err = lookupDNS(qmethod.Option, dname)
	}
	return key, err
}

// ParseTXT parse DNS TXT resource record into Key.
func ParseTXT(txt []byte, ttl uint32) (key *Key, err error) {
	if len(txt) == 0 {
		return nil, nil
	}

	p := newParser(txt)

	key = &Key{}
	for {
		tag, err := p.fetchTag()
		if err != nil {
			return nil, err
		}
		if tag == nil {
			break
		}
		err = key.set(tag)
		if err != nil {
			return nil, err
		}
	}

	key.ExpiredAt = time.Now().Unix() + int64(ttl)

	return key, nil
}

// Pack the key to be used in DNS TXT record.
func (key *Key) Pack() string {
	if key == nil {
		return ""
	}

	var bb strings.Builder
	sep := []byte{';', ' '}

	if len(key.Version) > 0 {
		bb.WriteString("v=")
		bb.Write(key.Version)
	}
	if len(key.Public) > 0 {
		if bb.Len() > 0 {
			bb.Write(sep)
		}
		bb.WriteString("p=")
		bb.Write(key.Public)
	}

	if key.Type != nil {
		if bb.Len() > 0 {
			bb.Write(sep)
		}
		bb.WriteString("k=")
		bb.Write(keyTypeNames[*key.Type])
	}

	if len(key.HashAlgs) > 0 {
		if bb.Len() > 0 {
			bb.Write(sep)
		}
		bb.WriteString("h=")
		bb.Write(packHashAlgs(key.HashAlgs))
	}
	if len(key.Notes) > 0 {
		if bb.Len() > 0 {
			bb.Write(sep)
		}
		bb.WriteString("n=")
		bb.Write(key.Notes)
	}
	if len(key.Services) > 0 {
		if bb.Len() > 0 {
			bb.Write(sep)
		}
		bb.WriteString("s=")
		bb.Write(bytes.Join(key.Services, sepColon))
	}
	if len(key.Flags) > 0 {
		if bb.Len() > 0 {
			bb.Write(sep)
		}
		bb.WriteString("t=")
		bb.Write(packKeyFlags(key.Flags))
	}

	return bb.String()
}

// IsExpired will return true if key ExpiredAt time is less than current time;
// otherwise it will return false.
func (key *Key) IsExpired() bool {
	return key.ExpiredAt < time.Now().Unix()
}

func (key *Key) set(t *tag) (err error) {
	if t == nil {
		return nil
	}
	switch t.key {
	case tagDNSPublicKey:
		pkey, err := base64.RawStdEncoding.DecodeString(string(t.value))
		if err != nil {
			err = fmt.Errorf("dkim: error decode public key: %w", err)
			return err
		}
		pk, err := x509.ParsePKIXPublicKey(pkey)
		if err != nil {
			err = fmt.Errorf("dkim: error parsing public key: %w", err)
			return err
		}

		key.RSA = pk.(*rsa.PublicKey)
		key.Public = t.value

	case tagDNSVersion:
		key.Version = t.value

	case tagDNSHashAlgs:
		key.HashAlgs = unpackHashAlgs(t.value)

	case tagDNSKeyType:
		key.Type = parseKeyType(t.value)

	case tagDNSNotes:
		key.Notes = t.value

	case tagDNSServices:
		key.Services = bytes.Split(t.value, sepColon)

	case tagDNSFlags:
		key.Flags = unpackKeyFlags(t.value)
	}
	return nil
}
